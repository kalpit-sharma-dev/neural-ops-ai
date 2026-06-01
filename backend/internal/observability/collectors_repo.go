package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CollectorsRepo reads production collector telemetry from Postgres.
type CollectorsRepo struct {
	pool *pgxpool.Pool
}

// NewCollectorsRepo creates a collectors repository.
func NewCollectorsRepo(pool *pgxpool.Pool) *CollectorsRepo {
	return &CollectorsRepo{pool: pool}
}

func (r *CollectorsRepo) available() bool {
	return r != nil && r.pool != nil
}

// ListHosts returns host inventory from collector.
func (r *CollectorsRepo) ListHosts(ctx context.Context, tenantID string) ([]HostSummary, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT name, status, cpu_percent, memory_percent, disk_percent, zone
FROM collector_hosts WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]HostSummary, 0)
	for rows.Next() {
		var h HostSummary
		if err := rows.Scan(&h.Name, &h.Status, &h.CPU, &h.Memory, &h.Disk, &h.Zone); err != nil {
			return nil, err
		}
		h.ID = h.Name
		out = append(out, h)
	}
	return out, rows.Err()
}

// ListK8sClusters returns K8s clusters from collector.
func (r *CollectorsRepo) ListK8sClusters(ctx context.Context, tenantID string) ([]K8sCluster, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, name, nodes, pods, health, namespace_count
FROM collector_k8s_clusters WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]K8sCluster, 0)
	for rows.Next() {
		var c K8sCluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Nodes, &c.Pods, &c.Health, &c.NamespaceCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListK8sPods returns pods from collector.
func (r *CollectorsRepo) ListK8sPods(ctx context.Context, tenantID, namespace string) ([]K8sPod, error) {
	if !r.available() {
		return nil, nil
	}
	query := `
SELECT name, namespace, node, status, cpu_percent, memory_percent, restarts
FROM collector_k8s_pods WHERE tenant_id = $1`
	args := []any{tenantID}
	if namespace != "" {
		query += ` AND namespace = $2`
		args = append(args, namespace)
	}
	query += ` ORDER BY namespace, name LIMIT 500`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]K8sPod, 0)
	for rows.Next() {
		var p K8sPod
		if err := rows.Scan(&p.Name, &p.Namespace, &p.Node, &p.Status, &p.CPU, &p.Memory, &p.Restarts); err != nil {
			return nil, err
		}
		p.ID = p.Namespace + "/" + p.Name
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListRUMSessions returns RUM sessions.
func (r *CollectorsRepo) ListRUMSessions(ctx context.Context, tenantID string, limit int) ([]RUMSession, error) {
	if !r.available() {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
SELECT session_key, user_id, page, device, country, duration_ms, errors, lcp, started_at
FROM rum_sessions WHERE tenant_id = $1 ORDER BY started_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RUMSession, 0)
	for rows.Next() {
		var s RUMSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Page, &s.Device, &s.Country, &s.DurationMs, &s.Errors, &s.LCP, &s.StartedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpsertRUMSession stores or updates a RUM session beacon.
func (r *CollectorsRepo) UpsertRUMSession(ctx context.Context, tenantID string, s RUMSession) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	if s.ID == "" {
		s.ID = uuid.New().String()[:16]
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO rum_sessions (tenant_id, session_key, user_id, page, device, country, duration_ms, errors, lcp, started_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,COALESCE($10, NOW()), NOW())
ON CONFLICT (tenant_id, session_key) DO UPDATE SET
  user_id=EXCLUDED.user_id, page=EXCLUDED.page, device=EXCLUDED.device, country=EXCLUDED.country,
  duration_ms=EXCLUDED.duration_ms, errors=EXCLUDED.errors, lcp=EXCLUDED.lcp, updated_at=NOW()`,
		tenantID, s.ID, s.UserID, s.Page, s.Device, s.Country, s.DurationMs, s.Errors, s.LCP, s.StartedAt,
	)
	return err
}

// AppendReplayEvent stores one session replay event.
func (r *CollectorsRepo) AppendReplayEvent(ctx context.Context, tenantID, sessionKey string, seq int, eventType string, payload map[string]any) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	payloadJSON, _ := json.Marshal(payload)
	_, err := r.pool.Exec(ctx, `
INSERT INTO rum_replay_events (tenant_id, session_key, seq, event_type, payload)
VALUES ($1,$2,$3,$4,$5)`, tenantID, sessionKey, seq, eventType, payloadJSON)
	return err
}

// ReplayEvent is one stored replay frame.
type ReplayEvent struct {
	Seq       int            `json:"seq"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	RecordedAt time.Time     `json:"recordedAt"`
}

// ListReplayEvents returns replay events for a session.
func (r *CollectorsRepo) ListReplayEvents(ctx context.Context, tenantID, sessionKey string) ([]ReplayEvent, error) {
	if !r.available() {
		return nil, fmt.Errorf("postgres unavailable")
	}
	rows, err := r.pool.Query(ctx, `
SELECT seq, event_type, payload, recorded_at
FROM rum_replay_events WHERE tenant_id = $1 AND session_key = $2 ORDER BY seq`, tenantID, sessionKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ReplayEvent, 0)
	for rows.Next() {
		var e ReplayEvent
		var payloadJSON []byte
		if err := rows.Scan(&e.Seq, &e.Type, &payloadJSON, &e.RecordedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payloadJSON, &e.Payload)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListSyntheticMonitors returns synthetic monitors from DB.
func (r *CollectorsRepo) ListSyntheticMonitors(ctx context.Context, tenantID string) ([]SyntheticMonitor, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, name, monitor_type, url, interval_label, locations, enabled, last_status, last_run_at
FROM synthetic_monitors WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SyntheticMonitor, 0)
	for rows.Next() {
		var m SyntheticMonitor
		var locJSON []byte
		var lastRun *time.Time
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.Interval, &locJSON, &m.Enabled, &m.LastStatus, &lastRun); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(locJSON, &m.Locations)
		if lastRun != nil {
			m.LastRunAt = *lastRun
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// SaveSyntheticMonitor upserts a monitor.
func (r *CollectorsRepo) SaveSyntheticMonitor(ctx context.Context, tenantID string, m SyntheticMonitor) (SyntheticMonitor, error) {
	if !r.available() {
		return m, fmt.Errorf("postgres unavailable")
	}
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	locJSON, _ := json.Marshal(m.Locations)
	err := r.pool.QueryRow(ctx, `
INSERT INTO synthetic_monitors (id, tenant_id, name, monitor_type, url, interval_label, locations, enabled, last_status, last_run_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (id) DO UPDATE SET
  name=EXCLUDED.name, url=EXCLUDED.url, enabled=EXCLUDED.enabled, last_status=EXCLUDED.last_status, last_run_at=EXCLUDED.last_run_at
RETURNING id::text`, m.ID, tenantID, m.Name, m.Type, m.URL, m.Interval, locJSON, m.Enabled, m.LastStatus, m.LastRunAt,
	).Scan(&m.ID)
	return m, err
}

// ListSyntheticRuns returns runs for a monitor.
func (r *CollectorsRepo) ListSyntheticRuns(ctx context.Context, tenantID, monitorID string) ([]SyntheticRun, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, monitor_id::text, status, latency_ms, location, ran_at
FROM synthetic_runs WHERE tenant_id = $1 AND monitor_id = $2 ORDER BY ran_at DESC LIMIT 100`, tenantID, monitorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SyntheticRun, 0)
	for rows.Next() {
		var run SyntheticRun
		if err := rows.Scan(&run.ID, &run.MonitorID, &run.Status, &run.LatencyMs, &run.Location, &run.RanAt); err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

// InsertSyntheticRun records a synthetic execution.
func (r *CollectorsRepo) InsertSyntheticRun(ctx context.Context, tenantID, monitorID, status string, latencyMs int64, location string) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO synthetic_runs (tenant_id, monitor_id, status, latency_ms, location)
VALUES ($1,$2,$3,$4,$5)`, tenantID, monitorID, status, latencyMs, location)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
UPDATE synthetic_monitors SET last_status=$3, last_run_at=NOW() WHERE tenant_id=$1 AND id=$2`, tenantID, monitorID, status)
	return err
}

// EnabledSyntheticMonitors returns monitors to execute.
func (r *CollectorsRepo) EnabledSyntheticMonitors(ctx context.Context, tenantID string) ([]SyntheticMonitor, error) {
	monitors, err := r.ListSyntheticMonitors(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]SyntheticMonitor, 0)
	for _, m := range monitors {
		if m.Enabled {
			out = append(out, m)
		}
	}
	return out, nil
}
