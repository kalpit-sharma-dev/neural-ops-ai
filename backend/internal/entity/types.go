// Package entity defines the unified observability entity model.
package entity

// Type identifies an observability entity kind.
type Type string

const (
	TypeService   Type = "service"
	TypeHost      Type = "host"
	TypePod       Type = "pod"
	TypeDatabase  Type = "database"
	TypeContainer Type = "container"
)

// Ref is a lightweight entity reference used across traces, metrics, and topology APIs.
type Ref struct {
	Type        Type              `json:"entityType"`
	ID          string            `json:"entityId"`
	TenantID    string            `json:"tenantId"`
	DisplayName string            `json:"displayName"`
	Labels      map[string]string `json:"labels,omitempty"`
	Health      string            `json:"health,omitempty"`
}
