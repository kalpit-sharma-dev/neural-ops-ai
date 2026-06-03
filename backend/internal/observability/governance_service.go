package observability

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GovernanceService evaluates ABAC and data residency policies per tenant.
type GovernanceService struct {
	pool *pgxpool.Pool
	mem  *Store
}

// NewGovernanceService creates a governance evaluator backed by Postgres with mem fallback.
func NewGovernanceService(pool *pgxpool.Pool, mem *Store) *GovernanceService {
	if mem == nil {
		mem = NewStore()
	}
	return &GovernanceService{pool: pool, mem: mem}
}

func (g *GovernanceService) abacPolicy(ctx context.Context, tenantID string) ABACPolicy {
	if g.pool != nil {
		if p, err := loadABACFromPostgres(ctx, g.pool, tenantID); err == nil {
			return p
		}
	}
	return g.mem.GetABACPolicy()
}

func (g *GovernanceService) residencyPolicy(ctx context.Context, tenantID string) DataResidencyPolicy {
	if g.pool != nil {
		if p, err := loadResidencyFromPostgres(ctx, g.pool, tenantID); err == nil {
			return p
		}
	}
	return g.mem.GetDataResidency()
}

// EvaluateABAC returns whether the request is allowed under tenant ABAC rules.
func (g *GovernanceService) EvaluateABAC(ctx context.Context, tenantID, role, email, method, path string) (bool, string) {
	policy := g.abacPolicy(ctx, tenantID)
	if !policy.Enabled {
		return true, ""
	}
	action := strings.ToLower(method)
	if action == "get" || action == "head" {
		action = "read"
	} else if action == "post" || action == "put" || action == "patch" || action == "delete" {
		action = strings.ToLower(method)
	}
	resource := path
	clearance := clearanceForRole(role)
	department := departmentFromEmail(email)
	for _, rule := range policy.Rules {
		if !matchABAC(rule, action, resource, clearance, department) {
			continue
		}
		if strings.EqualFold(rule.Effect, "deny") {
			return false, "abac rule " + rule.ID + " denied " + action + " on " + resource
		}
		return true, ""
	}
	return true, ""
}

// ResidencyAllowsIngest checks whether telemetry ingest is allowed for a region header.
func (g *GovernanceService) ResidencyAllowsIngest(ctx context.Context, tenantID, region string) (bool, string) {
	p := g.residencyPolicy(ctx, tenantID)
	if region == "" {
		region = p.PrimaryRegion
	}
	allowed := false
	for _, r := range p.AllowedRegions {
		if strings.EqualFold(r, region) {
			allowed = true
			break
		}
	}
	if !allowed && p.CrossBorderDenied {
		return false, "data residency policy blocks region " + region
	}
	return true, ""
}

func clearanceForRole(role string) int {
	switch strings.ToUpper(role) {
	case "ADMIN":
		return 5
	case "SRE":
		return 4
	case "ALERT_MANAGER":
		return 3
	case "DEVELOPER":
		return 3
	default:
		return 1
	}
}

func departmentFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "platform"
	}
	domain := strings.Split(parts[1], ".")[0]
	if domain == "neuralops" {
		return "platform"
	}
	return domain
}

func matchABAC(rule ABACPolicyRule, action, resource string, clearance int, department string) bool {
	if rule.Action != "*" && !strings.Contains(action, strings.ToLower(rule.Action)) {
		return false
	}
	if rule.Resource != "*" {
		pattern := strings.TrimSuffix(rule.Resource, "/*")
		pattern = strings.TrimSuffix(pattern, "*")
		if !strings.Contains(resource, pattern) {
			return false
		}
	}
	cond := rule.Condition
	cond = strings.ReplaceAll(cond, "user.clearance", strconv.Itoa(clearance))
	cond = strings.ReplaceAll(cond, "user.department", department)
	cond = strings.ReplaceAll(cond, "resource.team", department)
	if strings.Contains(cond, "< 3") && clearance < 3 {
		return true
	}
	if strings.Contains(cond, "==") {
		parts := strings.Split(cond, "==")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "user.department" {
			return strings.TrimSpace(parts[1]) == department
		}
	}
	if strings.TrimSpace(cond) == "" {
		return true
	}
	return strings.EqualFold(rule.Effect, "allow")
}
