package notifier

import (
	"context"

	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/domain"
)

// Notifier dispatches alert notifications.
type Notifier interface {
	Send(ctx context.Context, alert model.AlertRecord, incident *domain.Incident) error
	Type() string
}

// Registry holds configured notifiers keyed by channel ID.
type Registry struct {
	byChannel map[string]Notifier
}

// NewRegistry creates an empty notifier registry.
func NewRegistry() *Registry {
	return &Registry{byChannel: map[string]Notifier{}}
}

// Register adds a notifier for a channel ID.
func (r *Registry) Register(channelID string, notifier Notifier) {
	r.byChannel[channelID] = notifier
}

// All returns all registered notifiers.
func (r *Registry) All() []Notifier {
	out := make([]Notifier, 0, len(r.byChannel))
	for _, notifier := range r.byChannel {
		out = append(out, notifier)
	}
	return out
}

// BuildFromChannel creates a notifier from channel config.
func BuildFromChannel(channel model.NotificationChannel) (Notifier, error) {
	switch channel.ChannelType {
	case "email":
		return NewEmailNotifier(channel.Config)
	case "slack":
		return NewSlackNotifier(channel.Config)
	case "teams":
		return NewTeamsNotifier(channel.Config)
	case "pagerduty":
		return NewPagerDutyNotifier(channel.Config)
	case "sms":
		return NewSMSNotifier(channel.Config)
	case "webhook":
		return NewWebhookNotifier(channel.Config)
	default:
		return NewWebhookNotifier(channel.Config)
	}
}
