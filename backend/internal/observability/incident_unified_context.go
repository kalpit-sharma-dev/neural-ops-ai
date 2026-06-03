package observability

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// IncidentDeepLinks are cross-signal navigation targets for war-room triage.
type IncidentDeepLinks struct {
	Logs       string `json:"logs"`
	Traces     string `json:"traces"`
	Metrics    string `json:"metrics"`
	Security   string `json:"security"`
	ServiceMap string `json:"serviceMap"`
	Workflows  string `json:"workflows"`
	AIChat     string `json:"aiChat"`
}

// SuggestedQuery is a one-click unified or log query for incident investigation.
type SuggestedQuery struct {
	Label string `json:"label"`
	Kind  string `json:"kind"` // logs|traces|metrics|nexql
	Href  string `json:"href"`
}

// IncidentUnifiedContext bundles correlated signals and deep links (single-pane-of-glass triage).
type IncidentUnifiedContext struct {
	IncidentID       string            `json:"incidentId"`
	PrimaryService   string            `json:"primaryService"`
	AffectedServices []string          `json:"affectedServices"`
	DeepLinks        IncidentDeepLinks `json:"deepLinks"`
	SecurityFindings []SecurityFinding `json:"securityFindings"`
	SuggestedQueries []SuggestedQuery  `json:"suggestedQueries"`
	GeneratedAt      time.Time         `json:"generatedAt"`
}

func (h *Handler) registerUnifiedRoutes(v1 *gin.RouterGroup) {
	v1.GET("/unified/incidents/:id/context", h.GetIncidentUnifiedContext)
}

func (h *Handler) GetIncidentUnifiedContext(c *gin.Context) {
	incidentID := c.Param("id")
	service := strings.TrimSpace(c.Query("service"))
	services := splitCSV(c.Query("services"))
	if service == "" && len(services) > 0 {
		service = services[0]
	}
	if service == "" {
		service = "payment-service"
	}

	findings := make([]SecurityFinding, 0)
	for _, f := range h.deps.Mem.ListSecurityFindings() {
		if f.IncidentID == incidentID || (f.IncidentID == "" && f.Service == service) {
			findings = append(findings, f)
		}
	}
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if list, err := h.deps.SRS.ListSecurityFindings(c.Request.Context(), tenantID(c)); err == nil {
			for _, f := range list {
				if f.IncidentID == incidentID || (f.IncidentID == "" && f.Service == service) {
					findings = append(findings, f)
				}
			}
		}
	}

	links := buildIncidentDeepLinks(service, incidentID)
	queries := buildSuggestedQueries(service, incidentID)

	writeSuccess(c, IncidentUnifiedContext{
		IncidentID:       incidentID,
		PrimaryService:   service,
		AffectedServices: services,
		DeepLinks:        links,
		SecurityFindings: findings,
		SuggestedQueries: queries,
		GeneratedAt:      time.Now().UTC(),
	})
}

func buildIncidentDeepLinks(service, incidentID string) IncidentDeepLinks {
	svc := url.QueryEscape(service)
	return IncidentDeepLinks{
		Logs:       fmt.Sprintf("/logs?service=%s&mode=errors", svc),
		Traces:     fmt.Sprintf("/traces?service=%s", svc),
		Metrics:    fmt.Sprintf("/metrics?service=%s", svc),
		Security:   fmt.Sprintf("/security?service=%s", svc),
		ServiceMap: "/service-map",
		Workflows:  fmt.Sprintf("/workflows?context=incident:%s", url.QueryEscape(incidentID)),
		AIChat:     fmt.Sprintf("/ai-chat?context=incident:%s", url.QueryEscape(incidentID)),
	}
}

func buildSuggestedQueries(service, incidentID string) []SuggestedQuery {
	svc := url.QueryEscape(service)
	return []SuggestedQuery{
		{Label: "Error logs (1h)", Kind: "logs", Href: fmt.Sprintf("/logs?service=%s&q=severity:error", svc)},
		{Label: "Slow traces", Kind: "traces", Href: fmt.Sprintf("/traces?service=%s&minDuration=500ms", svc)},
		{Label: "Latency metrics", Kind: "metrics", Href: fmt.Sprintf("/metrics?service=%s&metric=http.server.duration", svc)},
		{Label: "NexQL cross-signal", Kind: "nexql", Href: fmt.Sprintf("/query-workbench?incident=%s&service=%s", url.QueryEscape(incidentID), svc)},
	}
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
