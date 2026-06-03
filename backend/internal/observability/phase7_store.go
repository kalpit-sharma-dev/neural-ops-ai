package observability

import (
	"time"

	"github.com/google/uuid"
)

func (s *Store) seedPhase7() {
	now := time.Now().UTC()
	s.abacPolicy = ABACPolicy{
		Enabled: true,
		Rules: []ABACPolicyRule{
			{ID: "abac-1", Effect: "allow", Action: "read", Resource: "logs:*", Condition: "user.department == resource.team"},
			{ID: "abac-2", Effect: "deny", Action: "export", Resource: "pii:*", Condition: "user.clearance < 3"},
		},
		UpdatedAt: now,
	}
	s.dataResidency = DataResidencyPolicy{
		PrimaryRegion: "us-east-1", AllowedRegions: []string{"us-east-1", "eu-west-1"},
		PIIStorageRegion: "us-east-1", CrossBorderDenied: true, UpdatedAt: now,
	}
	s.branding = BrandingTheme{
		ProductName: "NeuralOps", LogoURL: "/brand/logo.svg",
		PrimaryColor: "#2563eb", AccentColor: "#0ea5e9",
		SupportEmail: "support@neuralops.ai", CustomDomain: "observe.acme-corp.com",
		UpdatedAt: now,
	}
	s.mspTenants = []MSPTenant{
		{ID: "t-acme", Name: "Acme Corp", Slug: "acme", Plan: "enterprise", Status: "active", UserCount: 42, Region: "us-east-1", CreatedAt: now.Add(-180 * 24 * time.Hour)},
		{ID: "t-globex", Name: "Globex", Slug: "globex", Plan: "pro", Status: "active", UserCount: 18, Region: "eu-west-1", CreatedAt: now.Add(-90 * 24 * time.Hour)},
		{ID: "t-initech", Name: "Initech", Slug: "initech", Plan: "trial", Status: "suspended", UserCount: 5, Region: "us-east-1", CreatedAt: now.Add(-14 * 24 * time.Hour)},
	}
}

func (s *Store) GetABACPolicy() ABACPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.abacPolicy
}

func (s *Store) PutABACPolicy(p ABACPolicy) ABACPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	s.abacPolicy = p
	return p
}

func (s *Store) GetDataResidency() DataResidencyPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dataResidency
}

func (s *Store) PutDataResidency(p DataResidencyPolicy) DataResidencyPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	s.dataResidency = p
	return p
}

func (s *Store) GetBranding() BrandingTheme {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.branding
}

func (s *Store) PutBranding(b BrandingTheme) BrandingTheme {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.UpdatedAt = time.Now().UTC()
	s.branding = b
	return b
}

func (s *Store) ListMSPTenants() []MSPTenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]MSPTenant, len(s.mspTenants))
	copy(out, s.mspTenants)
	return out
}

// CreateMSPTenant registers a child tenant for MSP operators (ADM-03).
func (s *Store) CreateMSPTenant(t MSPTenant) MSPTenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if t.ID == "" {
		t.ID = "t-" + uuid.New().String()[:8]
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.Status == "" {
		t.Status = "active"
	}
	s.mspTenants = append(s.mspTenants, t)
	return t
}

func (s *Store) CreateExportJob(exportType, destination string) ExportJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	fin := now.Add(3 * time.Second)
	job := ExportJob{
		ID:           "exp-" + uuid.New().String()[:8],
		Type:         exportType,
		Destination:  destination,
		Status:       "completed",
		RowsExported: 125000,
		StartedAt:    now,
		FinishedAt:   &fin,
		Message:      "Export delivered to destination",
	}
	switch exportType {
	case "warehouse":
		job.RowsExported = 2_400_000
		job.Message = "Parquet batch written to warehouse bucket"
	case "bi":
		job.RowsExported = 890_000
		job.Message = "Dataset refresh triggered for BI connector"
	case "events":
		job.RowsExported = 450_000
		job.Message = "Event stream checkpoint published"
	}
	s.exportJobs[job.ID] = job
	return job
}
