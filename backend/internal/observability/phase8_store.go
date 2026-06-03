package observability

import "time"

func (s *Store) seedPhase8() {
	now := time.Now().UTC()
	s.nfrBenchmarks = []NFRBenchmark{
		{ID: "bm-ingest", Name: "Log ingestion throughput", Category: "ingestion", Target: ">= 500k eps", Actual: "612k eps", Unit: "eps", Pass: true, MeasuredAt: now},
		{ID: "bm-query", Name: "Cross-signal query p95", Category: "query", Target: "<= 800ms", Actual: "420ms", Unit: "ms", Pass: true, MeasuredAt: now},
		{ID: "bm-alert", Name: "Alert evaluation latency p99", Category: "alerting", Target: "<= 30s", Actual: "12s", Unit: "s", Pass: true, MeasuredAt: now},
		{ID: "bm-api", Name: "Gateway API availability", Category: "availability", Target: ">= 99.95%", Actual: "99.97%", Unit: "percent", Pass: true, MeasuredAt: now},
	}
	s.nfrReliability = []NFRReliabilityDrill{
		{ID: "dr-1", Name: "Multi-region failover", Type: "dr", Region: "us-east-1 → us-west-2", Status: "passed", RTOSeconds: 185, RTOSLOSeconds: 300, Pass: true, ExecutedAt: now.Add(-7 * 24 * time.Hour)},
		{ID: "ha-1", Name: "Gateway pod disruption", Type: "ha", Region: "us-east-1", Status: "passed", RTOSeconds: 42, RTOSLOSeconds: 60, Pass: true, ExecutedAt: now.Add(-3 * 24 * time.Hour)},
		{ID: "ha-2", Name: "ClickHouse replica loss", Type: "ha", Region: "eu-west-1", Status: "passed", RTOSeconds: 95, RTOSLOSeconds: 120, Pass: true, ExecutedAt: now.Add(-14 * 24 * time.Hour)},
	}
	s.nfrA11y = NFRA11yReport{
		Standard: "WCAG", Level: "2.1 AA", PagesAudited: 12,
		ViolationsCritical: 0, ViolationsSerious: 0, Pass: true, LastAuditAt: now.Add(-2 * 24 * time.Hour),
	}
	s.nfrLocales = []NFRLocale{
		{Code: "en", Name: "English", CoveragePct: 100, Enabled: true},
		{Code: "es", Name: "Español", CoveragePct: 78, Enabled: true},
		{Code: "de", Name: "Deutsch", CoveragePct: 72, Enabled: true},
		{Code: "fr", Name: "Français", CoveragePct: 65, Enabled: false},
	}
	s.nfrCertification = NFRCertificationReport{
		Version: "2026.06", SignedAt: now, OverallPass: true,
		BenchmarksPass: true, ReliabilityPass: true, AccessibilityPass: true, I18nReady: true,
		EvidenceURIs: []string{
			"s3://neuralops-compliance/benchmarks/2026-06-report.pdf",
			"s3://neuralops-compliance/dr/2026-06-drill.json",
			"s3://neuralops-compliance/a11y/axe-2026-06.json",
		},
	}
}

func (s *Store) ListNFRBenchmarks() []NFRBenchmark {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NFRBenchmark, len(s.nfrBenchmarks))
	copy(out, s.nfrBenchmarks)
	return out
}

func (s *Store) ListNFRReliability() []NFRReliabilityDrill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NFRReliabilityDrill, len(s.nfrReliability))
	copy(out, s.nfrReliability)
	return out
}

func (s *Store) NFRA11yReport() NFRA11yReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nfrA11y
}

func (s *Store) ListNFRLocales() []NFRLocale {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NFRLocale, len(s.nfrLocales))
	copy(out, s.nfrLocales)
	return out
}

func (s *Store) NFRCertification() NFRCertificationReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nfrCertification
}
