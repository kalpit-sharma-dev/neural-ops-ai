package escalation

import (
	"context"
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/alerting/config"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/alerting/notifier"
	"github.com/neuralops/platform/internal/alerting/oncall"
	"github.com/neuralops/platform/internal/alerting/repository"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

// Engine escalates unacknowledged P1 alerts using DB policies or config fallback.
type Engine struct {
	cfg      config.AlertingConfig
	log      *zap.Logger
	repo     *repository.Store
	notifier *notifier.Registry
	oncall   *oncall.Schedule
}

// NewEngine creates an escalation engine.
func NewEngine(cfg config.AlertingConfig, log *zap.Logger, repo *repository.Store, registry *notifier.Registry, schedule *oncall.Schedule) *Engine {
	return &Engine{cfg: cfg, log: log, repo: repo, notifier: registry, oncall: schedule}
}

// Run starts periodic escalation checks until context cancellation.
func (e *Engine) Run(ctx context.Context) {
	interval := e.cfg.EscalationCheckInterval
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.check(ctx); err != nil {
				e.log.Warn("escalation check failed", zap.Error(err))
			}
		}
	}
}

func (e *Engine) check(ctx context.Context) error {
	alerts, err := e.repo.ListUnacknowledgedP1(ctx, e.cfg.DefaultTenant)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, alert := range alerts {
		policy, _ := e.repo.MatchEscalationPolicy(ctx, alert.TenantID, alert.Service)
		targetLevel, message := e.resolveLevel(alert, policy, now)
		if targetLevel <= alert.EscalationLevel {
			continue
		}

		alert.EscalationLevel = targetLevel
		alert.Description = fmt.Sprintf("%s\n[%s]", alert.Description, message)
		if err := e.repo.UpdateAlert(ctx, &alert); err != nil {
			return err
		}

		oncallPerson := e.oncall.Current("platform-sre", now)
		incident := &domain.Incident{Title: fmt.Sprintf("Escalation L%d: %s", targetLevel, alert.Title)}
		escalated := alert
		escalated.Title = fmt.Sprintf("[Escalation L%d] %s", targetLevel, alert.Title)
		if oncallPerson != nil {
			escalated.Description = fmt.Sprintf("%s\nOn-call: %s (%s)", escalated.Description, oncallPerson.Name, oncallPerson.Email)
		}
		if e.cfg.DemoMode {
			e.log.Info("demo mode: skipping escalation notification",
				zap.String("alert", alert.Title),
				zap.Int("level", targetLevel),
			)
			continue
		}
		for _, n := range e.notifier.All() {
			if err := n.Send(ctx, escalated, incident); err != nil {
				e.log.Warn("escalation notification failed", zap.Error(err))
			}
		}
	}
	return nil
}

func (e *Engine) resolveLevel(alert model.AlertRecord, policy *model.EscalationPolicy, now time.Time) (int, string) {
	age := now.Sub(alert.FiredAt)
	if policy != nil && len(policy.Levels) > 0 {
		target := alert.EscalationLevel
		msg := ""
		for _, lvl := range policy.Levels {
			d, err := time.ParseDuration(lvl.After)
			if err != nil {
				continue
			}
			if age >= d && lvl.Level > target {
				target = lvl.Level
				msg = fmt.Sprintf("Escalated to %s after %s", lvl.Target, lvl.After)
			}
		}
		if target > alert.EscalationLevel {
			return target, msg
		}
	}
	// Config fallback
	switch {
	case age >= e.cfg.P1EscalateManager && alert.EscalationLevel < 2:
		return 2, "Escalated to manager on-call"
	case age >= e.cfg.P1EscalateL2 && alert.EscalationLevel < 1:
		return 1, "Escalated to L2 on-call"
	default:
		return alert.EscalationLevel, ""
	}
}
