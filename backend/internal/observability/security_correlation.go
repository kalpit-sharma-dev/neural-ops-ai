package observability

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// SecurityCorrelationService links findings to incidents.
type SecurityCorrelationService struct {
	mem *Store
	srs *SRSRepo
}

// NewSecurityCorrelationService wires correlation helpers.
func NewSecurityCorrelationService(mem *Store, srs *SRSRepo) *SecurityCorrelationService {
	return &SecurityCorrelationService{mem: mem, srs: srs}
}

// CorrelateFinding attaches or creates an incident for a security finding.
func (s *SecurityCorrelationService) CorrelateFinding(ctx context.Context, tenantID, findingID string) (SecurityFinding, error) {
	f, err := s.loadFinding(ctx, tenantID, findingID)
	if err != nil {
		return SecurityFinding{}, err
	}
	if f.IncidentID != "" {
		return f, nil
	}
	incidentID := s.matchExistingIncident(f)
	if incidentID == "" {
		incidentID = "inc-sec-" + uuid.New().String()[:8]
	}
	f.IncidentID = incidentID
	s.mem.LinkSecurityFindingIncident(findingID, incidentID)
	if s.srs != nil && s.srs.available() {
		_ = s.srs.SaveSecurityFinding(ctx, tenantID, f)
	}
	return f, nil
}

func (s *SecurityCorrelationService) loadFinding(ctx context.Context, tenantID, id string) (SecurityFinding, error) {
	if s.srs != nil && s.srs.available() {
		if f, err := s.srs.GetSecurityFinding(ctx, tenantID, id); err == nil {
			return f, nil
		}
	}
	for _, f := range s.mem.ListSecurityFindings() {
		if f.ID == id {
			return f, nil
		}
	}
	return SecurityFinding{}, fmt.Errorf("finding not found")
}

func (s *SecurityCorrelationService) matchExistingIncident(f SecurityFinding) string {
	svc := strings.ToLower(f.Service)
	if svc == "" {
		return ""
	}
	// Demo correlation: payment-service maps to seeded inc-1.
	if strings.Contains(svc, "payment") {
		return "inc-1"
	}
	return ""
}
