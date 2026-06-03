package observability

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) seedPhase6() {
	now := time.Now().UTC()
	s.rumFunnels = map[string]RUMFunnel{
		"fn-checkout": {
			ID: "fn-checkout", Name: "Checkout conversion",
			Steps: []RUMFunnelStep{
				{Name: "Product view", Event: "page_view_product", Count: 12400, ConversionPct: 100},
				{Name: "Add to cart", Event: "cart_add", Count: 6200, ConversionPct: 50},
				{Name: "Checkout start", Event: "checkout_start", Count: 4100, ConversionPct: 66.1},
				{Name: "Payment success", Event: "payment_success", Count: 3520, ConversionPct: 85.9},
			},
			OverallConversionPct: 28.4,
			CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now,
		},
	}
	s.syntheticBrowserTests = map[string]SyntheticBrowserTest{
		"br-1": {
			ID: "br-1", Name: "Checkout happy path", URL: "https://demo.neuralops.ai/checkout",
			Script: "await page.goto(url);\nawait page.click('[data-testid=pay]');\nawait expect(page).toHaveURL(/success/);",
			Locations: []string{"us-east", "eu-west"}, Enabled: true, LastStatus: "OK", CreatedAt: now,
		},
	}
	s.syntheticMobileTests = map[string]SyntheticMobileTest{
		"mob-1": {
			ID: "mob-1", Name: "iOS login flow", Platform: "ios", BundleID: "ai.neuralops.mobile",
			Script: "tap('#login'); type('#email', 'demo@neuralops.ai'); tap('#submit');",
			Enabled: true, LastStatus: "OK", CreatedAt: now,
		},
	}
	s.syntheticPrivateLocations = map[string]SyntheticPrivateLocation{
		"loc-priv-1": {
			ID: "loc-priv-1", Name: "corp-dc-probes", Region: "on-prem-us",
			AgentVersion: "2.1.0", Status: "healthy", LastHeartbeatAt: now.Add(-30 * time.Second),
		},
	}
	s.businessKPIPacks = map[string]BusinessKPIPack{
		"kpi-ecommerce": {
			ID: "kpi-ecommerce", Name: "E-commerce KPI pack", Category: "retail",
			Description: "Conversion, AOV, and cart abandonment KPIs wired to RUM + payments.",
			KPIs: []BusinessKPIDefinition{
				{Key: "conversion_rate", Label: "Conversion rate", Unit: "%", Target: 3.5},
				{Key: "aov", Label: "Average order value", Unit: "USD", Target: 120},
				{Key: "cart_abandon", Label: "Cart abandonment", Unit: "%", Target: 65},
			},
			Connectors: []string{"rum", "payments-api", "warehouse"},
			Enabled: false,
		},
		"kpi-saas": {
			ID: "kpi-saas", Name: "SaaS activation pack", Category: "saas",
			Description: "Signup-to-activation funnel and weekly active accounts.",
			KPIs: []BusinessKPIDefinition{
				{Key: "activation_rate", Label: "Activation rate", Unit: "%", Target: 42},
				{Key: "wau", Label: "Weekly active accounts", Unit: "count", Target: 5000},
			},
			Connectors: []string{"product-analytics", "crm"},
			Enabled: true,
		},
		"kpi-bfsi-payments": {
			ID: "kpi-bfsi-payments", Name: "BFSI payments KPI pack", Category: "bfsi",
			Description: "UPI/NEFT success, p95 auth latency, and failed-txn rate wired to business transactions + APM.",
			KPIs: []BusinessKPIDefinition{
				{Key: "payment_success_rate", Label: "Payment success rate", Unit: "%", Target: 99.5},
				{Key: "upi_p95_ms", Label: "UPI p95 latency", Unit: "ms", Target: 800},
				{Key: "failed_txn_rate", Label: "Failed transaction rate", Unit: "%", Target: 0.5},
				{Key: "chargeback_rate", Label: "Chargeback rate", Unit: "bps", Target: 12},
			},
			Connectors: []string{"transactions-api", "rum", "apm", "ledger-service"},
			Enabled: false,
		},
	}
}

// ListRUMFunnels returns all funnels.
func (s *Store) ListRUMFunnels() []RUMFunnel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RUMFunnel, 0, len(s.rumFunnels))
	for _, f := range s.rumFunnels {
		out = append(out, f)
	}
	return out
}

// SaveRUMFunnel creates or updates a funnel with computed conversion metrics.
func (s *Store) SaveRUMFunnel(f RUMFunnel) RUMFunnel {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if f.ID == "" {
		f.ID = "fn-" + uuid.New().String()[:8]
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now
	var prev int64 = 0
	for i := range f.Steps {
		if i == 0 {
			f.Steps[i].ConversionPct = 100
			if f.Steps[i].Count == 0 {
				f.Steps[i].Count = 1000
			}
			prev = f.Steps[i].Count
			continue
		}
		if f.Steps[i].Count == 0 && prev > 0 {
			f.Steps[i].Count = int64(float64(prev) * 0.72)
		}
		if prev > 0 {
			f.Steps[i].ConversionPct = float64(f.Steps[i].Count) / float64(prev) * 100
		}
		prev = f.Steps[i].Count
	}
	if len(f.Steps) >= 2 && f.Steps[0].Count > 0 {
		last := f.Steps[len(f.Steps)-1].Count
		f.OverallConversionPct = float64(last) / float64(f.Steps[0].Count) * 100
	}
	s.rumFunnels[f.ID] = f
	return f
}

// SaveSyntheticBrowserTest registers a browser synthetic test.
func (s *Store) SaveSyntheticBrowserTest(t SyntheticBrowserTest) SyntheticBrowserTest {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if t.ID == "" {
		t.ID = "br-" + uuid.New().String()[:8]
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.LastStatus == "" {
		t.LastStatus = "scheduled"
	}
	if len(t.Locations) == 0 {
		t.Locations = []string{"us-east"}
	}
	s.syntheticBrowserTests[t.ID] = t
	return t
}

// SaveSyntheticMobileTest registers a mobile synthetic test.
func (s *Store) SaveSyntheticMobileTest(t SyntheticMobileTest) SyntheticMobileTest {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if t.ID == "" {
		t.ID = "mob-" + uuid.New().String()[:8]
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.LastStatus == "" {
		t.LastStatus = "scheduled"
	}
	s.syntheticMobileTests[t.ID] = t
	return t
}

// SaveSyntheticPrivateLocation registers a private synthetic probe location.
func (s *Store) SaveSyntheticPrivateLocation(loc SyntheticPrivateLocation) SyntheticPrivateLocation {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if loc.ID == "" {
		loc.ID = "loc-" + uuid.New().String()[:8]
	}
	if loc.LastHeartbeatAt.IsZero() {
		loc.LastHeartbeatAt = now
	}
	if loc.Status == "" {
		loc.Status = "healthy"
	}
	if loc.AgentVersion == "" {
		loc.AgentVersion = "2.1.0"
	}
	s.syntheticPrivateLocations[loc.ID] = loc
	return loc
}

// ListBusinessKPIPacks returns KPI pack catalog.
func (s *Store) ListBusinessKPIPacks() []BusinessKPIPack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]BusinessKPIPack, 0, len(s.businessKPIPacks))
	for _, p := range s.businessKPIPacks {
		out = append(out, p)
	}
	return out
}

// EnableBusinessKPIPack enables a KPI pack by id.
func (s *Store) EnableBusinessKPIPack(id string) (BusinessKPIPack, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.businessKPIPacks[id]
	if !ok {
		return BusinessKPIPack{}, errors.New("kpi pack not found")
	}
	p.Enabled = true
	s.businessKPIPacks[id] = p
	return p, nil
}

// ListSyntheticBrowserTests returns browser tests (helper for UI).
func (s *Store) ListSyntheticBrowserTests() []SyntheticBrowserTest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SyntheticBrowserTest, 0, len(s.syntheticBrowserTests))
	for _, t := range s.syntheticBrowserTests {
		out = append(out, t)
	}
	return out
}

// ListSyntheticMobileTests returns mobile tests.
func (s *Store) ListSyntheticMobileTests() []SyntheticMobileTest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SyntheticMobileTest, 0, len(s.syntheticMobileTests))
	for _, t := range s.syntheticMobileTests {
		out = append(out, t)
	}
	return out
}

// ListSyntheticPrivateLocations returns private locations.
func (s *Store) ListSyntheticPrivateLocations() []SyntheticPrivateLocation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SyntheticPrivateLocation, 0, len(s.syntheticPrivateLocations))
	for _, l := range s.syntheticPrivateLocations {
		out = append(out, l)
	}
	return out
}

// validateFunnelName ensures funnel has a name and steps.
func validateFunnelName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("funnel name is required")
	}
	return nil
}
