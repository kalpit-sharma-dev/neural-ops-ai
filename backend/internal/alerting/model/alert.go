package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/domain"
)

// AlertStatus represents alert lifecycle state.
type AlertStatus string

const (
	AlertStatusFiring       AlertStatus = "FIRING"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusSuppressed   AlertStatus = "SUPPRESSED"
	AlertStatusResolved     AlertStatus = "RESOLVED"
)

// AlertRecord is the persisted alerting entity.
type AlertRecord struct {
	ID               uuid.UUID               `json:"id"`
	TenantID         string                  `json:"tenantId"`
	Source           domain.AlertSource      `json:"source"`
	AlertName        string                  `json:"alertName"`
	Service          string                  `json:"service"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	Severity         domain.IncidentSeverity `json:"severity"`
	Status           AlertStatus             `json:"status"`
	Fingerprint      string                  `json:"fingerprint"`
	GroupID          *uuid.UUID              `json:"groupId,omitempty"`
	OccurrenceCount  int                     `json:"occurrenceCount"`
	Labels           map[string]string       `json:"labels,omitempty"`
	LinkedIncidentID *uuid.UUID              `json:"linkedIncidentId,omitempty"`
	AIExplanation    string                  `json:"aiExplanation,omitempty"`
	FiredAt          time.Time               `json:"firedAt"`
	LastSeenAt       time.Time               `json:"lastSeenAt"`
	ResolvedAt       *time.Time              `json:"resolvedAt,omitempty"`
	AcknowledgedAt   *time.Time              `json:"acknowledgedAt,omitempty"`
	SuppressedUntil  *time.Time              `json:"suppressedUntil,omitempty"`
	Deduplicated     bool                    `json:"deduplicated"`
	EscalationLevel  int                     `json:"escalationLevel"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
}

// ToDomain converts to shared Alert model.
func (a AlertRecord) ToDomain() domain.Alert {
	return domain.Alert{
		ID:               a.ID,
		Source:           a.Source,
		Title:            a.Title,
		Description:      a.Description,
		Severity:         a.Severity,
		FiredAt:          a.FiredAt,
		ResolvedAt:       a.ResolvedAt,
		Labels:           a.Labels,
		LinkedIncidentID: a.LinkedIncidentID,
		Deduplicated:     a.Deduplicated,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

// AlertRule defines routing/classification for incoming alerts.
type AlertRule struct {
	ID             uuid.UUID          `json:"id"`
	TenantID       string             `json:"tenantId"`
	Name           string             `json:"name"`
	Source         domain.AlertSource `json:"source"`
	ServicePattern string             `json:"servicePattern,omitempty"`
	Severity       domain.IncidentSeverity `json:"severity"`
	Enabled        bool               `json:"enabled"`
	Labels         map[string]string  `json:"labels,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// NotificationChannel stores outbound notification configuration.
type NotificationChannel struct {
	ID          uuid.UUID         `json:"id"`
	TenantID    string            `json:"tenantId"`
	Name        string            `json:"name"`
	ChannelType string            `json:"channelType"`
	Config      map[string]any    `json:"config"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// Silence suppresses alerts for a period.
type Silence struct {
	ID               uuid.UUID `json:"id"`
	TenantID         string    `json:"tenantId"`
	ServicePattern   string    `json:"servicePattern,omitempty"`
	AlertNamePattern string    `json:"alertNamePattern,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	StartsAt         time.Time `json:"startsAt"`
	EndsAt           time.Time `json:"endsAt"`
	CreatedAt        time.Time `json:"createdAt"`
}

// EscalationPolicy defines escalation chain.
type EscalationPolicy struct {
	ID             uuid.UUID         `json:"id"`
	TenantID       string            `json:"tenantId"`
	Name           string            `json:"name"`
	ServicePattern string            `json:"servicePattern,omitempty"`
	Levels         []EscalationLevel `json:"levels"`
	Enabled        bool              `json:"enabled"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// EscalationLevel is one step in an escalation chain.
type EscalationLevel struct {
	Level       int    `json:"level"`
	After       string `json:"after"`
	Target      string `json:"target"`
	ChannelType string `json:"channelType"`
}

// IncomingAlert is a normalized alert before persistence.
type IncomingAlert struct {
	Source      domain.AlertSource
	AlertName   string
	Service     string
	Title       string
	Description string
	Severity    domain.IncidentSeverity
	FiredAt     time.Time
	Labels      map[string]string
	Resolved    bool
}
