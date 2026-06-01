package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/domain"
)

// BuildTimelineEvent creates an incident timeline event.
func BuildTimelineEvent(eventType domain.IncidentEventType, title, description string, at time.Time) domain.IncidentEvent {
	return domain.IncidentEvent{
		ID:          uuid.New(),
		Timestamp:   at.UTC(),
		EventType:   eventType,
		Title:       title,
		Description: description,
	}
}

// AppendTimelineEvent appends an event keeping chronological order.
func AppendTimelineEvent(incident *domain.Incident, event domain.IncidentEvent) {
	incident.Timeline = append(incident.Timeline, event)
}

// BuildNarrative generates a human-readable incident narrative.
func BuildNarrative(incident domain.Incident) string {
	if len(incident.Timeline) == 0 {
		return incident.Summary
	}

	parts := make([]string, 0, len(incident.Timeline)+1)
	parts = append(parts, incident.Summary)
	for _, event := range incident.Timeline {
		parts = append(parts, fmt.Sprintf("At %s, %s: %s",
			event.Timestamp.Format("15:04"),
			strings.ToLower(string(event.EventType)),
			event.Title,
		))
	}
	return strings.Join(parts, " ")
}

// ComputeMTTR calculates MTTR when resolved.
func ComputeMTTR(incident *domain.Incident, resolvedAt time.Time) time.Duration {
	if incident.StartTime.IsZero() || resolvedAt.Before(incident.StartTime) {
		return 0
	}
	return resolvedAt.Sub(incident.StartTime)
}
