package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
)

// AlertFilter filters alert queries.
type AlertFilter struct {
	TenantID string
	Status   string
	Severity string
	Service  string
	Source   string
	Limit    int
	Offset   int
}

// Store persists alerting data.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates an alerting repository.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close closes the pool.
func (s *Store) Close() { s.pool.Close() }

// Pool exposes the underlying connection pool for health checks.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// FindByFingerprint returns active alert by fingerprint.
func (s *Store) FindByFingerprint(ctx context.Context, tenantID, fingerprint string) (*model.AlertRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
       group_id, occurrence_count, labels, linked_incident_id, ai_explanation, fired_at, last_seen_at,
       resolved_at, acknowledged_at, suppressed_until, deduplicated, escalation_level, created_at, updated_at
FROM alerts
WHERE tenant_id = $1 AND fingerprint = $2 AND status IN ('FIRING', 'ACKNOWLEDGED', 'SUPPRESSED')
ORDER BY last_seen_at DESC LIMIT 1`, tenantID, fingerprint)
	return scanAlert(row)
}

// CreateAlert inserts a new alert.
func (s *Store) CreateAlert(ctx context.Context, alert *model.AlertRecord) error {
	labelsJSON, _ := json.Marshal(alert.Labels)
	_, err := s.pool.Exec(ctx, `
INSERT INTO alerts (
  id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
  group_id, occurrence_count, labels, linked_incident_id, ai_explanation, fired_at, last_seen_at,
  resolved_at, acknowledged_at, suppressed_until, deduplicated, escalation_level, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,NOW(),NOW())`,
		alert.ID, alert.TenantID, alert.Source, alert.AlertName, alert.Service, alert.Title, alert.Description,
		alert.Severity, alert.Status, alert.Fingerprint, alert.GroupID, alert.OccurrenceCount, labelsJSON,
		alert.LinkedIncidentID, alert.AIExplanation, alert.FiredAt, alert.LastSeenAt, alert.ResolvedAt,
		alert.AcknowledgedAt, alert.SuppressedUntil, alert.Deduplicated, alert.EscalationLevel,
	)
	return err
}

// UpdateAlert updates an alert record.
func (s *Store) UpdateAlert(ctx context.Context, alert *model.AlertRecord) error {
	labelsJSON, _ := json.Marshal(alert.Labels)
	_, err := s.pool.Exec(ctx, `
UPDATE alerts SET
  title=$2, description=$3, severity=$4, status=$5, group_id=$6, occurrence_count=$7, labels=$8,
  linked_incident_id=$9, ai_explanation=$10, last_seen_at=$11, resolved_at=$12, acknowledged_at=$13,
  suppressed_until=$14, deduplicated=$15, escalation_level=$16, updated_at=NOW()
WHERE id=$1`,
		alert.ID, alert.Title, alert.Description, alert.Severity, alert.Status, alert.GroupID,
		alert.OccurrenceCount, labelsJSON, alert.LinkedIncidentID, alert.AIExplanation, alert.LastSeenAt,
		alert.ResolvedAt, alert.AcknowledgedAt, alert.SuppressedUntil, alert.Deduplicated, alert.EscalationLevel,
	)
	return err
}

// GetAlert returns alert by ID.
func (s *Store) GetAlert(ctx context.Context, id uuid.UUID) (*model.AlertRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
       group_id, occurrence_count, labels, linked_incident_id, ai_explanation, fired_at, last_seen_at,
       resolved_at, acknowledged_at, suppressed_until, deduplicated, escalation_level, created_at, updated_at
FROM alerts WHERE id = $1`, id)
	return scanAlert(row)
}

// ListAlerts lists alerts with filters.
func (s *Store) ListAlerts(ctx context.Context, filter AlertFilter) ([]model.AlertRecord, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query := `
SELECT id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
       group_id, occurrence_count, labels, linked_incident_id, ai_explanation, fired_at, last_seen_at,
       resolved_at, acknowledged_at, suppressed_until, deduplicated, escalation_level, created_at, updated_at
FROM alerts WHERE 1=1`
	args := make([]any, 0, 8)
	idx := 1
	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, filter.TenantID)
		idx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, filter.Status)
		idx++
	}
	if filter.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", idx)
		args = append(args, filter.Severity)
		idx++
	}
	if filter.Service != "" {
		query += fmt.Sprintf(" AND service = $%d", idx)
		args = append(args, filter.Service)
		idx++
	}
	if filter.Source != "" {
		query += fmt.Sprintf(" AND source = $%d", idx)
		args = append(args, filter.Source)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY last_seen_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, filter.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.AlertRecord, 0)
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *alert)
	}
	return results, rows.Err()
}

// FindGroupCandidate finds an alert in the same service within a time window.
func (s *Store) FindGroupCandidate(ctx context.Context, tenantID, service string, since time.Time) (*uuid.UUID, error) {
	var groupID uuid.UUID
	err := s.pool.QueryRow(ctx, `
SELECT COALESCE(group_id, id) FROM alerts
WHERE tenant_id = $1 AND service = $2 AND last_seen_at >= $3 AND status IN ('FIRING', 'ACKNOWLEDGED')
ORDER BY last_seen_at DESC LIMIT 1`, tenantID, service, since).Scan(&groupID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &groupID, nil
}

// ListActiveSilences returns active silences.
func (s *Store) ListActiveSilences(ctx context.Context, tenantID string, now time.Time) ([]model.Silence, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, service_pattern, alert_name_pattern, reason, starts_at, ends_at, created_at
FROM alert_silences
WHERE tenant_id = $1 AND starts_at <= $2 AND ends_at >= $2`, tenantID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Silence, 0)
	for rows.Next() {
		var silence model.Silence
		if err := rows.Scan(&silence.ID, &silence.TenantID, &silence.ServicePattern, &silence.AlertNamePattern,
			&silence.Reason, &silence.StartsAt, &silence.EndsAt, &silence.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, silence)
	}
	return out, rows.Err()
}

// CreateSilence inserts a silence window.
func (s *Store) CreateSilence(ctx context.Context, silence *model.Silence) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO alert_silences (id, tenant_id, service_pattern, alert_name_pattern, reason, starts_at, ends_at, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`,
		silence.ID, silence.TenantID, silence.ServicePattern, silence.AlertNamePattern,
		silence.Reason, silence.StartsAt, silence.EndsAt,
	)
	return err
}

// ListUnacknowledgedP1 returns firing P1 alerts without acknowledgement.
func (s *Store) ListUnacknowledgedP1(ctx context.Context, tenantID string) ([]model.AlertRecord, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, source, alert_name, service, title, description, severity, status, fingerprint,
       group_id, occurrence_count, labels, linked_incident_id, ai_explanation, fired_at, last_seen_at,
       resolved_at, acknowledged_at, suppressed_until, deduplicated, escalation_level, created_at, updated_at
FROM alerts
WHERE tenant_id = $1 AND severity = 'P1' AND status = 'FIRING' AND acknowledged_at IS NULL`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAlerts(rows)
}

// UpdateRule updates an alert rule.
func (s *Store) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	labelsJSON, _ := json.Marshal(rule.Labels)
	tag, err := s.pool.Exec(ctx, `
UPDATE alert_rules SET name=$2, source=$3, service_pattern=$4, severity=$5, enabled=$6, labels=$7, updated_at=NOW()
WHERE id=$1 AND tenant_id=$8`,
		rule.ID, rule.Name, rule.Source, rule.ServicePattern, rule.Severity, rule.Enabled, labelsJSON, rule.TenantID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// DeleteRule removes an alert rule.
func (s *Store) DeleteRule(ctx context.Context, tenantID string, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM alert_rules WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

// GetRule returns a rule by ID.
func (s *Store) GetRule(ctx context.Context, tenantID string, id uuid.UUID) (*model.AlertRule, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, tenant_id, name, source, service_pattern, severity, enabled, labels, created_at, updated_at
FROM alert_rules WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	var rule model.AlertRule
	var labelsJSON []byte
	if err := row.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Source, &rule.ServicePattern,
		&rule.Severity, &rule.Enabled, &labelsJSON, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(labelsJSON, &rule.Labels)
	return &rule, nil
}

// UpdateChannel updates a notification channel.
func (s *Store) UpdateChannel(ctx context.Context, channel *model.NotificationChannel) error {
	configJSON, _ := json.Marshal(channel.Config)
	tag, err := s.pool.Exec(ctx, `
UPDATE notification_channels SET name=$2, channel_type=$3, config=$4, enabled=$5, updated_at=NOW()
WHERE id=$1 AND tenant_id=$6`,
		channel.ID, channel.Name, channel.ChannelType, configJSON, channel.Enabled, channel.TenantID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// DeleteChannel removes a notification channel.
func (s *Store) DeleteChannel(ctx context.Context, tenantID string, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM notification_channels WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// ListSilences lists silences for tenant (active and future).
func (s *Store) ListSilences(ctx context.Context, tenantID string) ([]model.Silence, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, service_pattern, alert_name_pattern, reason, starts_at, ends_at, created_at
FROM alert_silences WHERE tenant_id = $1 ORDER BY starts_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Silence, 0)
	for rows.Next() {
		var silence model.Silence
		if err := rows.Scan(&silence.ID, &silence.TenantID, &silence.ServicePattern, &silence.AlertNamePattern,
			&silence.Reason, &silence.StartsAt, &silence.EndsAt, &silence.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, silence)
	}
	return out, rows.Err()
}

// CreateRule inserts alert rule.
func (s *Store) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	labelsJSON, _ := json.Marshal(rule.Labels)
	_, err := s.pool.Exec(ctx, `
INSERT INTO alert_rules (id, tenant_id, name, source, service_pattern, severity, enabled, labels, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`,
		rule.ID, rule.TenantID, rule.Name, rule.Source, rule.ServicePattern, rule.Severity, rule.Enabled, labelsJSON,
	)
	return err
}

// ListRules lists alert rules.
func (s *Store) ListRules(ctx context.Context, tenantID string) ([]model.AlertRule, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, name, source, service_pattern, severity, enabled, labels, created_at, updated_at
FROM alert_rules WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.AlertRule, 0)
	for rows.Next() {
		var rule model.AlertRule
		var labelsJSON []byte
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Source, &rule.ServicePattern,
			&rule.Severity, &rule.Enabled, &labelsJSON, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(labelsJSON, &rule.Labels)
		out = append(out, rule)
	}
	return out, rows.Err()
}

// CreateChannel inserts notification channel.
func (s *Store) CreateChannel(ctx context.Context, channel *model.NotificationChannel) error {
	configJSON, _ := json.Marshal(channel.Config)
	_, err := s.pool.Exec(ctx, `
INSERT INTO notification_channels (id, tenant_id, name, channel_type, config, enabled, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
		channel.ID, channel.TenantID, channel.Name, channel.ChannelType, configJSON, channel.Enabled,
	)
	return err
}

// ListChannels lists notification channels.
func (s *Store) ListChannels(ctx context.Context, tenantID string) ([]model.NotificationChannel, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, name, channel_type, config, enabled, created_at, updated_at
FROM notification_channels WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.NotificationChannel, 0)
	for rows.Next() {
		var channel model.NotificationChannel
		var configJSON []byte
		if err := rows.Scan(&channel.ID, &channel.TenantID, &channel.Name, &channel.ChannelType,
			&configJSON, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(configJSON, &channel.Config)
		out = append(out, channel)
	}
	return out, rows.Err()
}

// ListEnabledChannels returns enabled channels for dispatch.
func (s *Store) ListEnabledChannels(ctx context.Context, tenantID string) ([]model.NotificationChannel, error) {
	channels, err := s.ListChannels(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]model.NotificationChannel, 0, len(channels))
	for _, channel := range channels {
		if channel.Enabled {
			out = append(out, channel)
		}
	}
	return out, nil
}

// FindOpenIncidentByService finds open incident for service enrichment.
func (s *Store) FindOpenIncidentByService(ctx context.Context, service string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
SELECT id FROM incidents
WHERE status IN ('OPEN', 'INVESTIGATING') AND $1 = ANY(affected_services)
ORDER BY start_time DESC LIMIT 1`, service).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func scanAlerts(rows pgx.Rows) ([]model.AlertRecord, error) {
	out := make([]model.AlertRecord, 0)
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *alert)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAlert(row rowScanner) (*model.AlertRecord, error) {
	var alert model.AlertRecord
	var labelsJSON []byte
	var groupID *uuid.UUID
	var linkedIncidentID *uuid.UUID
	if err := row.Scan(
		&alert.ID, &alert.TenantID, &alert.Source, &alert.AlertName, &alert.Service, &alert.Title,
		&alert.Description, &alert.Severity, &alert.Status, &alert.Fingerprint, &groupID,
		&alert.OccurrenceCount, &labelsJSON, &linkedIncidentID, &alert.AIExplanation, &alert.FiredAt,
		&alert.LastSeenAt, &alert.ResolvedAt, &alert.AcknowledgedAt, &alert.SuppressedUntil,
		&alert.Deduplicated, &alert.EscalationLevel, &alert.CreatedAt, &alert.UpdatedAt,
	); err != nil {
		return nil, err
	}
	alert.GroupID = groupID
	alert.LinkedIncidentID = linkedIncidentID
	_ = json.Unmarshal(labelsJSON, &alert.Labels)
	if alert.Labels == nil {
		alert.Labels = map[string]string{}
	}
	return &alert, nil
}

// MatchSilence checks if alert matches a silence rule.
func MatchSilence(alert model.IncomingAlert, silence model.Silence) bool {
	if silence.ServicePattern != "" && !strings.Contains(strings.ToLower(alert.Service), strings.ToLower(silence.ServicePattern)) {
		return false
	}
	if silence.AlertNamePattern != "" && !strings.Contains(strings.ToLower(alert.AlertName), strings.ToLower(silence.AlertNamePattern)) {
		return false
	}
	return true
}

// NormalizeSource ensures valid alert source enum.
func NormalizeSource(source domain.AlertSource) domain.AlertSource {
	if source.IsValid() {
		return source
	}
	return domain.AlertSourceCustom
}
