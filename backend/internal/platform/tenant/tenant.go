package tenant

import (
	"context"
	"strings"
)

const HeaderName = "X-Tenant-ID"

type contextKey string

const tenantContextKey contextKey = "neuralops_tenant_id"

// WithContext stores tenant ID on context.
func WithContext(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return context.WithValue(ctx, tenantContextKey, tenantID)
}

// FromContext returns tenant ID from context.
func FromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(tenantContextKey).(string)
	return value, ok && value != ""
}

// NormalizeID trims and defaults empty tenant identifiers.
func NormalizeID(tenantID, fallback string) string {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return strings.TrimSpace(fallback)
	}
	return tenantID
}

// SanitizeForIndex normalizes tenant IDs for Elasticsearch index names.
func SanitizeForIndex(tenantID string) string {
	tenantID = strings.ToLower(strings.TrimSpace(tenantID))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	return replacer.Replace(tenantID)
}

// LogsAlias returns the write/search alias for a tenant's logs.
func LogsAlias(tenantID string) string {
	return "tenant-" + SanitizeForIndex(tenantID) + "-logs"
}

// LogsIndexPrefix returns the index prefix pattern for a tenant.
func LogsIndexPrefix(tenantID string) string {
	return LogsAlias(tenantID) + "-*"
}

// LogsBootstrapIndex returns the initial rollover index for a tenant.
func LogsBootstrapIndex(tenantID string) string {
	return LogsAlias(tenantID) + "-000001"
}
