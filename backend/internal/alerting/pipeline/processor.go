package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/alerting/config"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/alerting/normalizer"
	"github.com/neuralops/platform/internal/alerting/notifier"
	"github.com/neuralops/platform/internal/alerting/repository"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

// Processor runs dedup, grouping, enrichment, suppression, and dispatch.
type Processor struct {
	cfg        config.AlertingConfig
	log        *zap.Logger
	repo       *repository.Store
	llm        ai.LLMClient
	notifier   *notifier.Registry
	mobilePush notifier.Notifier
}

// NewProcessor creates an alert processor.
func NewProcessor(cfg config.AlertingConfig, log *zap.Logger, repo *repository.Store, llm ai.LLMClient, registry *notifier.Registry, mobilePush notifier.Notifier) *Processor {
	return &Processor{cfg: cfg, log: log, repo: repo, llm: llm, notifier: registry, mobilePush: mobilePush}
}

// Process ingests normalized alerts.
func (p *Processor) Process(ctx context.Context, tenantID string, incoming []model.IncomingAlert) ([]model.AlertRecord, error) {
	if tenantID == "" {
		tenantID = p.cfg.DefaultTenant
	}
	silences, _ := p.repo.ListActiveSilences(ctx, tenantID, time.Now().UTC())
	results := make([]model.AlertRecord, 0, len(incoming))

	for _, item := range incoming {
		if suppressed(item, silences) {
			p.log.Info("alert suppressed", zap.String("alertName", item.AlertName), zap.String("service", item.Service))
			continue
		}

		fingerprint := normalizer.Fingerprint(item.Source, item.AlertName, item.Service, item.Labels)
		existing, err := p.repo.FindByFingerprint(ctx, tenantID, fingerprint)
		if err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		if existing != nil {
			existing.OccurrenceCount++
			existing.LastSeenAt = now
			existing.Deduplicated = true
			if item.Resolved {
				existing.Status = model.AlertStatusResolved
				existing.ResolvedAt = &now
			}
			if err := p.repo.UpdateAlert(ctx, existing); err != nil {
				return nil, err
			}
			results = append(results, *existing)
			continue
		}

		groupSince := now.Add(-p.cfg.GroupWindow)
		groupID, _ := p.repo.FindGroupCandidate(ctx, tenantID, item.Service, groupSince)

		alert := &model.AlertRecord{
			ID:              uuid.New(),
			TenantID:        tenantID,
			Source:          item.Source,
			AlertName:       item.AlertName,
			Service:         item.Service,
			Title:           item.Title,
			Description:     item.Description,
			Severity:        item.Severity,
			Status:          model.AlertStatusFiring,
			Fingerprint:     fingerprint,
			GroupID:         groupID,
			OccurrenceCount: 1,
			Labels:          item.Labels,
			FiredAt:         item.FiredAt,
			LastSeenAt:      now,
		}
		if item.Resolved {
			alert.Status = model.AlertStatusResolved
			alert.ResolvedAt = &now
		}

		incidentID, _ := p.repo.FindOpenIncidentByService(ctx, item.Service)
		alert.LinkedIncidentID = incidentID
		alert.AIExplanation = p.enrich(ctx, item)

		if err := p.repo.CreateAlert(ctx, alert); err != nil {
			return nil, err
		}
		if alert.Status == model.AlertStatusFiring {
			p.dispatch(ctx, *alert, incidentID)
		}
		results = append(results, *alert)
	}
	return results, nil
}

func (p *Processor) enrich(ctx context.Context, incoming model.IncomingAlert) string {
	if p.llm == nil {
		return incoming.Description
	}
	answer, err := p.llm.AnswerQuery(ctx, fmt.Sprintf("Explain this alert briefly for on-call engineers: %s - %s", incoming.Title, incoming.Description), nil)
	if err != nil || answer == "" {
		return incoming.Description
	}
	return answer
}

func (p *Processor) dispatch(ctx context.Context, alert model.AlertRecord, incidentID *uuid.UUID) {
	if p.mobilePush != nil && alert.Status == model.AlertStatusFiring {
		if err := p.mobilePush.Send(ctx, alert, nil); err != nil {
			p.log.Warn("mobile push failed", zap.Error(err))
		}
	}
	if p.cfg.DemoMode {
		p.log.Info("demo mode: skipping alert notification dispatch",
			zap.String("alert", alert.Title),
			zap.String("service", alert.Service),
		)
		return
	}

	var incident *domain.Incident
	if incidentID != nil {
		incident = &domain.Incident{ID: *incidentID, Title: alert.Title}
	}
	for _, n := range p.notifier.All() {
		if err := n.Send(ctx, alert, incident); err != nil {
			p.log.Warn("notification dispatch failed", zap.String("type", n.Type()), zap.Error(err))
		}
	}
}

func suppressed(alert model.IncomingAlert, silences []model.Silence) bool {
	for _, silence := range silences {
		if repository.MatchSilence(alert, silence) {
			return true
		}
	}
	return false
}
