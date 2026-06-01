package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/alerting/notifier"
	"github.com/neuralops/platform/internal/alerting/pipeline"
	"github.com/neuralops/platform/internal/alerting/repository"
)

// Service orchestrates alerting operations.
type Service struct {
	cfg       string
	processor *pipeline.Processor
	repo      *repository.Store
	registry  *notifier.Registry
}

// New creates an alerting service.
func New(tenantID string, processor *pipeline.Processor, repo *repository.Store, registry *notifier.Registry) *Service {
	if tenantID == "" {
		tenantID = "default"
	}
	return &Service{cfg: tenantID, processor: processor, repo: repo, registry: registry}
}

// Ingest processes incoming normalized alerts.
func (s *Service) Ingest(ctx context.Context, tenantID string, incoming []model.IncomingAlert) ([]model.AlertRecord, error) {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.processor.Process(ctx, tenantID, incoming)
}

// ListAlerts lists alerts with filters.
func (s *Service) ListAlerts(ctx context.Context, filter repository.AlertFilter) ([]model.AlertRecord, error) {
	if filter.TenantID == "" {
		filter.TenantID = s.cfg
	}
	return s.repo.ListAlerts(ctx, filter)
}

// GetAlert returns alert detail.
func (s *Service) GetAlert(ctx context.Context, id uuid.UUID) (*model.AlertRecord, error) {
	return s.repo.GetAlert(ctx, id)
}

// AcknowledgeAlert marks alert as acknowledged.
func (s *Service) AcknowledgeAlert(ctx context.Context, id uuid.UUID) (*model.AlertRecord, error) {
	alert, err := s.repo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	alert.Status = model.AlertStatusAcknowledged
	alert.AcknowledgedAt = &now
	if err := s.repo.UpdateAlert(ctx, alert); err != nil {
		return nil, err
	}
	return alert, nil
}

// SuppressAlert suppresses alert for a duration.
func (s *Service) SuppressAlert(ctx context.Context, id uuid.UUID, duration time.Duration, reason string) (*model.AlertRecord, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}
	alert, err := s.repo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	until := time.Now().UTC().Add(duration)
	alert.Status = model.AlertStatusSuppressed
	alert.SuppressedUntil = &until
	if err := s.repo.UpdateAlert(ctx, alert); err != nil {
		return nil, err
	}

	silence := &model.Silence{
		ID:               uuid.New(),
		TenantID:         alert.TenantID,
		ServicePattern:   alert.Service,
		AlertNamePattern: alert.AlertName,
		Reason:           reason,
		StartsAt:         time.Now().UTC(),
		EndsAt:           until,
	}
	_ = s.repo.CreateSilence(ctx, silence)
	return alert, nil
}

// UpdateRule updates an alert rule.
func (s *Service) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	if rule.TenantID == "" {
		rule.TenantID = s.cfg
	}
	return s.repo.UpdateRule(ctx, rule)
}

// DeleteRule deletes an alert rule.
func (s *Service) DeleteRule(ctx context.Context, tenantID string, id uuid.UUID) error {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.DeleteRule(ctx, tenantID, id)
}

// UpdateChannel updates a notification channel.
func (s *Service) UpdateChannel(ctx context.Context, channel *model.NotificationChannel) error {
	if channel.TenantID == "" {
		channel.TenantID = s.cfg
	}
	if err := s.repo.UpdateChannel(ctx, channel); err != nil {
		return err
	}
	if channel.Enabled {
		n, err := notifier.BuildFromChannel(*channel)
		if err != nil {
			return err
		}
		s.registry.Register(channel.ID.String(), n)
	}
	return nil
}

// DeleteChannel deletes a notification channel.
func (s *Service) DeleteChannel(ctx context.Context, tenantID string, id uuid.UUID) error {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.DeleteChannel(ctx, tenantID, id)
}

// ListSilences lists silences.
func (s *Service) ListSilences(ctx context.Context, tenantID string) ([]model.Silence, error) {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.ListSilences(ctx, tenantID)
}

// CreateSilence creates a maintenance silence window.
func (s *Service) CreateSilence(ctx context.Context, silence *model.Silence) error {
	if silence.ID == uuid.Nil {
		silence.ID = uuid.New()
	}
	if silence.TenantID == "" {
		silence.TenantID = s.cfg
	}
	if silence.StartsAt.IsZero() {
		silence.StartsAt = time.Now().UTC()
	}
	return s.repo.CreateSilence(ctx, silence)
}

// CreateRule creates an alert rule.
func (s *Service) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	if rule.TenantID == "" {
		rule.TenantID = s.cfg
	}
	return s.repo.CreateRule(ctx, rule)
}

// ListRules lists alert rules.
func (s *Service) ListRules(ctx context.Context, tenantID string) ([]model.AlertRule, error) {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.ListRules(ctx, tenantID)
}

// CreateChannel creates a notification channel and registers notifier.
func (s *Service) CreateChannel(ctx context.Context, channel *model.NotificationChannel) error {
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}
	if channel.TenantID == "" {
		channel.TenantID = s.cfg
	}
	if err := s.repo.CreateChannel(ctx, channel); err != nil {
		return err
	}
	if channel.Enabled {
		n, err := notifier.BuildFromChannel(*channel)
		if err != nil {
			return err
		}
		s.registry.Register(channel.ID.String(), n)
	}
	return nil
}

// ListChannels lists notification channels.
func (s *Service) ListChannels(ctx context.Context, tenantID string) ([]model.NotificationChannel, error) {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.ListChannels(ctx, tenantID)
}

// ReloadNotifiers rebuilds notifier registry from DB.
func (s *Service) ReloadNotifiers(ctx context.Context, tenantID string) error {
	channels, err := s.repo.ListEnabledChannels(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		n, err := notifier.BuildFromChannel(channel)
		if err != nil {
			continue
		}
		s.registry.Register(channel.ID.String(), n)
	}
	return nil
}

// ListEscalationPolicies lists escalation policies.
func (s *Service) ListEscalationPolicies(ctx context.Context, tenantID string) ([]model.EscalationPolicy, error) {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.ListEscalationPolicies(ctx, tenantID)
}

// CreateEscalationPolicy creates an escalation policy.
func (s *Service) CreateEscalationPolicy(ctx context.Context, p *model.EscalationPolicy) error {
	if p.TenantID == "" {
		p.TenantID = s.cfg
	}
	return s.repo.CreateEscalationPolicy(ctx, p)
}

// UpdateEscalationPolicy updates an escalation policy.
func (s *Service) UpdateEscalationPolicy(ctx context.Context, p *model.EscalationPolicy) error {
	if p.TenantID == "" {
		p.TenantID = s.cfg
	}
	return s.repo.UpdateEscalationPolicy(ctx, p)
}

// DeleteEscalationPolicy deletes an escalation policy.
func (s *Service) DeleteEscalationPolicy(ctx context.Context, tenantID string, id uuid.UUID) error {
	if tenantID == "" {
		tenantID = s.cfg
	}
	return s.repo.DeleteEscalationPolicy(ctx, tenantID, id)
}
