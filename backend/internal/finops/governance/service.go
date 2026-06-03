package governance

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Service manages FinOps governance policies.
type Service struct {
	mu       sync.RWMutex
	policies map[string][]domain.GovernancePolicy
}

// NewService creates governance service with defaults.
func NewService() *Service {
	s := &Service{policies: map[string][]domain.GovernancePolicy{}}
	s.seed("default")
	return s
}

func (s *Service) seed(tenantID string) {
	now := time.Now().UTC()
	s.policies[tenantID] = []domain.GovernancePolicy{
		{
			ID: "gov-1", Name: "Default FinOps governance", ResidencyRegion: "us-east-1",
			AllowedScopes: []string{"payments", "platform"}, ChargebackMode: "showback",
			Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
	}
}

// List returns governance policies for tenant.
func (s *Service) List(tenantID string) []domain.GovernancePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.GovernancePolicy(nil), s.policies[tenantID]...)
}

// GetChargebackMode returns default chargeback mode for tenant.
func (s *Service) GetChargebackMode(tenantID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.policies[tenantID] {
		if p.Enabled && p.ChargebackMode != "" {
			return p.ChargebackMode
		}
	}
	return "showback"
}

// Upsert creates or updates a governance policy.
func (s *Service) Upsert(tenantID string, p domain.GovernancePolicy) domain.GovernancePolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = "gov-" + uuid.NewString()[:8]
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = time.Now().UTC()
	found := false
	for i, x := range s.policies[tenantID] {
		if x.ID == p.ID {
			s.policies[tenantID][i] = p
			found = true
			break
		}
	}
	if !found {
		s.policies[tenantID] = append(s.policies[tenantID], p)
	}
	return p
}
