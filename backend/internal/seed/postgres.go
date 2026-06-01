package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/security"
)

// PostgresOptions configures relational seed loading.
type PostgresOptions struct {
	DSN string
}

// SeedPostgres loads tenant, users, incidents, alerts, graph, and deployments.
func SeedPostgres(ctx context.Context, opts PostgresOptions) error {
	pool, err := db.NewPool(ctx, opts.DSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	now := time.Now().UTC()
	deployments := GenerateDeployments(now)

	if err := ensureDemoTenant(ctx, pool); err != nil {
		return err
	}
	if err := ensureDemoUser(ctx, pool); err != nil {
		return err
	}
	if err := seedCoreIncidents(ctx, pool, now); err != nil {
		return err
	}
	if err := seedHistoricalIncidents(ctx, pool, now); err != nil {
		return err
	}
	if err := seedAlerts(ctx, pool, now); err != nil {
		return err
	}
	if err := seedServiceGraph(ctx, pool); err != nil {
		return err
	}
	return seedDeploymentRecords(ctx, pool, deployments, now)
}

func ensureDemoTenant(ctx context.Context, pool *pgxpool.Pool) error {
	id := uuid.MustParse(DemoTenantID)
	_, err := pool.Exec(ctx, `
INSERT INTO tenants (id, name, plan_tier, subscription_status)
VALUES ($1, $2, 'enterprise', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`, id, DemoTenantName)
	return err
}

func ensureDemoUser(ctx context.Context, pool *pgxpool.Pool) error {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	tenantID := uuid.MustParse(DemoTenantID)
	_, err := pool.Exec(ctx, `
INSERT INTO users (id, tenant_id, email, role)
VALUES ($1, $2, $3, 'ADMIN')
ON CONFLICT (tenant_id, email) DO UPDATE SET role = EXCLUDED.role`,
		userID, tenantID, DemoEmail,
	)
	if err != nil {
		return err
	}
	keyHash := security.HashAPIKey("demo-api-key")
	_, err = pool.Exec(ctx, `
INSERT INTO api_keys (id, tenant_id, key_hash, name, role)
VALUES ($1, $2, $3, 'Demo API Key', 'ADMIN')
ON CONFLICT (key_hash) DO NOTHING`,
		uuid.MustParse("00000000-0000-0000-0000-000000000201"), tenantID, keyHash,
	)
	return err
}

func seedCoreIncidents(ctx context.Context, pool *pgxpool.Pool, now time.Time) error {
	timeline := NewScenarioTimeline(now)
	incidents := []struct {
		id, title, summary, severity, status, fingerprint string
		services                                          []string
		startTime, resolvedTime                           *time.Time
	}{
		{
			id: UPIOutageIncidentID,
			title: "UPI Service Outage",
			summary: "Deployment v2.3.1 introduced blocking DB calls causing 1,247 failed UPI transactions before rollback.",
			severity: "P1", status: "RESOLVED", fingerprint: "seed-upi-outage",
			services: []string{"upi-service", "payment-api", "ledger-service", "notification-service"},
			startTime: &timeline.OutageStart,
			resolvedTime: &timeline.ResolvedAt,
		},
		{
			id: "00000000-0000-0000-0000-000000000302",
			title: "Auth Service Latency Spike",
			summary: "Elevated p99 latency on auth-service due to Redis connection pool exhaustion.",
			severity: "P2", status: "INVESTIGATING", fingerprint: "seed-auth-latency",
			services: []string{"auth-service", "api-gateway"},
			startTime: ptrTime(now.Add(-20 * time.Minute)),
		},
		{
			id: "00000000-0000-0000-0000-000000000303",
			title: "Ledger Reconciliation Delay",
			summary: "Batch reconciliation lag detected in ledger-service overnight window.",
			severity: "P3", status: "OPEN", fingerprint: "seed-ledger-delay",
			services: []string{"ledger-service", "cbs-adapter"},
			startTime: ptrTime(now.Add(-6 * time.Hour)),
		},
	}

	for _, item := range incidents {
		_, err := pool.Exec(ctx, `
INSERT INTO incidents (
  id, title, summary, severity, status, affected_services, root_service, error_category,
  fingerprint, blast_radius, start_time, resolved_time, tenant_id, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
ON CONFLICT (id) DO UPDATE SET
  summary = EXCLUDED.summary,
  status = EXCLUDED.status,
  resolved_time = EXCLUDED.resolved_time,
  updated_at = NOW()`,
			item.id, item.title, item.summary, item.severity, item.status, item.services,
			item.services[0], "TIMEOUT", item.fingerprint, item.services,
			item.startTime, item.resolvedTime, DemoTenantID,
		)
		if err != nil {
			return err
		}
	}

	_, err := pool.Exec(ctx, `
INSERT INTO deployment_incident_correlations (deployment_id, incident_id, correlation_strength, detected_at)
VALUES ($1, $2, 0.94, $3)
ON CONFLICT DO NOTHING`,
		UPIOutageDeploymentID, UPIOutageIncidentID, timeline.DeployAt,
	)
	return err
}

func seedHistoricalIncidents(ctx context.Context, pool *pgxpool.Pool, now time.Time) error {
	severities := []string{"P1", "P2", "P3", "P4"}
	statuses := []string{"OPEN", "INVESTIGATING", "RESOLVED"}
	for i := 4; i <= 30; i++ {
		service := DemoServices[i%len(DemoServices)]
		start := now.Add(-time.Duration(i) * 6 * time.Hour)
		var resolved *time.Time
		if i%3 == 0 {
			resolved = ptrTime(start.Add(2 * time.Hour))
		}
		_, err := pool.Exec(ctx, `
INSERT INTO incidents (
  id, title, summary, severity, status, affected_services, root_service, error_category,
  fingerprint, blast_radius, start_time, resolved_time, tenant_id, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
ON CONFLICT DO NOTHING`,
			uuid.New(), fmt.Sprintf("Historical incident #%d", i),
			fmt.Sprintf("Automated seed incident for %s.", service),
			severities[i%len(severities)], statuses[i%len(statuses)],
			[]string{service}, service, "TIMEOUT", fmt.Sprintf("seed-incident-%d", i),
			[]string{service}, start, resolved, DemoTenantID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedAlerts(ctx context.Context, pool *pgxpool.Pool, now time.Time) error {
	alerts := []struct {
		id, name, service, title, description, severity, fingerprint, status string
		offset                                                               time.Duration
	}{
		{"00000000-0000-0000-0000-000000000401", "HighErrorRate", "upi-service",
			"UPI error rate above threshold", "Error rate exceeded 5% for upi-service over 5 minutes.",
			"P1", "seed-alert-upi-errors", "FIRING", -10 * time.Minute},
		{"00000000-0000-0000-0000-000000000402", "LatencyP99", "auth-service",
			"Auth p99 latency elevated", "p99 latency above 800ms for auth-service.",
			"P2", "seed-alert-auth-latency", "ACKNOWLEDGED", -30 * time.Minute},
		{"00000000-0000-0000-0000-000000000403", "QueueDepth", "notification-service",
			"Notification queue depth high", "Pending notification queue depth exceeded 10k messages.",
			"P3", "seed-alert-notify-queue", "FIRING", -5 * time.Minute},
	}
	for _, item := range alerts {
		firedAt := now.Add(item.offset)
		_, err := pool.Exec(ctx, `
INSERT INTO alerts (
  id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
  occurrence_count, labels, fired_at, last_seen_at, deduplicated, escalation_level, created_at, updated_at
) VALUES ($1,$2,'PROMETHEUS',$3,$4,$5,$6,$7,$8,$9,1,'{}',$10,$10,false,0,NOW(),NOW())
ON CONFLICT (id) DO NOTHING`,
			item.id, DemoTenantID, item.name, item.service, item.title, item.description,
			item.severity, item.status, item.fingerprint, firedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedServiceGraph(ctx context.Context, pool *pgxpool.Pool) error {
	edges := [][2]string{
		{"api-gateway", "auth-service"},
		{"api-gateway", "upi-service"},
		{"upi-service", "payment-api"},
		{"payment-api", "ledger-service"},
		{"payment-api", "notification-service"},
		{"ledger-service", "cbs-adapter"},
		{"upi-service", "fraud-service"},
	}
	for _, edge := range edges {
		_, err := pool.Exec(ctx, `
INSERT INTO service_dependencies (id, tenant_id, source_service, target_service, call_count, p99_latency_ms, p50_latency_ms, error_rate)
VALUES ($1, $2, $3, $4, 1200, 45, 18, 0.002)
ON CONFLICT (source_service, target_service) DO UPDATE SET
  tenant_id = EXCLUDED.tenant_id, call_count = EXCLUDED.call_count,
  p99_latency_ms = EXCLUDED.p99_latency_ms, updated_at = NOW()`,
			uuid.New(), DemoTenantID, edge[0], edge[1],
		)
		if err != nil {
			return err
		}
	}
	for i, service := range DemoServices {
		_, err := pool.Exec(ctx, `
INSERT INTO service_health (id, tenant_id, service_name, health_score, error_rate, latency_p99, anomaly_score, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (tenant_id, service_name) DO UPDATE SET
  health_score = EXCLUDED.health_score, error_rate = EXCLUDED.error_rate,
  latency_p99 = EXCLUDED.latency_p99, anomaly_score = EXCLUDED.anomaly_score, updated_at = NOW()`,
			uuid.New(), DemoTenantID, service, 100-float64(i%4)*8,
			0.001+float64(i%3)*0.003, 40+float64(i*12), float64(i%5)*12,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedDeploymentRecords(ctx context.Context, pool *pgxpool.Pool, deployments []DeploymentRecord, now time.Time) error {
	for _, dep := range deployments {
		_, err := pool.Exec(ctx, `
INSERT INTO deployments (id, tenant_id, service, version, environment, deployed_by, deployed_at, change_type, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
ON CONFLICT (id) DO NOTHING`,
			dep.ID, DemoTenantID, dep.Service, dep.Version, dep.Environment,
			dep.DeployedBy, dep.DeployedAt, dep.ChangeType,
			fmt.Sprintf(`{"category":"%s"}`, dep.Category),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }
