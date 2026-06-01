package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/alerting/model"
)

// ListEscalationPolicies returns escalation policies for tenant.
func (s *Store) ListEscalationPolicies(ctx context.Context, tenantID string) ([]model.EscalationPolicy, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, name, service_pattern, levels, enabled, created_at, updated_at
FROM escalation_policies WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.EscalationPolicy, 0)
	for rows.Next() {
		var p model.EscalationPolicy
		var levelsJSON []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.ServicePattern, &levelsJSON, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(levelsJSON, &p.Levels)
		out = append(out, p)
	}
	return out, rows.Err()
}

// CreateEscalationPolicy inserts a policy.
func (s *Store) CreateEscalationPolicy(ctx context.Context, p *model.EscalationPolicy) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	levelsJSON, _ := json.Marshal(p.Levels)
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
INSERT INTO escalation_policies (id, tenant_id, name, service_pattern, levels, enabled, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`,
		p.ID, p.TenantID, p.Name, p.ServicePattern, levelsJSON, p.Enabled, now)
	if err == nil {
		p.CreatedAt = now
		p.UpdatedAt = now
	}
	return err
}

// UpdateEscalationPolicy updates a policy.
func (s *Store) UpdateEscalationPolicy(ctx context.Context, p *model.EscalationPolicy) error {
	levelsJSON, _ := json.Marshal(p.Levels)
	tag, err := s.pool.Exec(ctx, `
UPDATE escalation_policies SET name=$2, service_pattern=$3, levels=$4, enabled=$5, updated_at=NOW()
WHERE id=$1 AND tenant_id=$6`,
		p.ID, p.Name, p.ServicePattern, levelsJSON, p.Enabled, p.TenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}
	return nil
}

// DeleteEscalationPolicy removes a policy.
func (s *Store) DeleteEscalationPolicy(ctx context.Context, tenantID string, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM escalation_policies WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}
	return nil
}

// MatchEscalationPolicy finds best policy for service.
func (s *Store) MatchEscalationPolicy(ctx context.Context, tenantID, service string) (*model.EscalationPolicy, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, name, service_pattern, levels, enabled, created_at, updated_at
FROM escalation_policies WHERE tenant_id = $1 AND enabled = true`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fallback *model.EscalationPolicy
	for rows.Next() {
		var p model.EscalationPolicy
		var levelsJSON []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.ServicePattern, &levelsJSON, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(levelsJSON, &p.Levels)
		if p.ServicePattern == "" {
			copy := p
			fallback = &copy
			continue
		}
		if service != "" && matchPattern(p.ServicePattern, service) {
			return &p, nil
		}
	}
	return fallback, rows.Err()
}

func matchPattern(pattern, service string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	return pattern == service || (len(pattern) > 0 && pattern[len(pattern)-1] == '*' &&
		len(service) >= len(pattern)-1 && service[:len(pattern)-1] == pattern[:len(pattern)-1])
}
