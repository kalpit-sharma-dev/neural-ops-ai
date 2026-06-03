package observability

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) registerNexQLRoutes(v1 *gin.RouterGroup) {
	q := v1.Group("/query")
	{
		q.POST("/explain", h.ExplainUnifiedQuery)
		q.GET("/saved", h.ListSavedQueries)
		q.POST("/saved", h.CreateSavedQuery)
		q.DELETE("/saved/:id", h.DeleteSavedQuery)
		q.GET("/cardinality-alerts", h.GetCardinalityAlertPolicy)
		q.PUT("/cardinality-alerts", h.PutCardinalityAlertPolicy)
	}
}

// SavedQuery is a tenant-scoped NexQL workbench query.
type SavedQuery struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Query     string `json:"query"`
	From      string `json:"from,omitempty"`
	Service   string `json:"service,omitempty"`
	TraceID   string `json:"traceId,omitempty"`
	TxnID     string `json:"txnId,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	UpdatedAt string `json:"updatedAt"`
}

func (h *Handler) ListSavedQueries(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListSavedQueries(tenantID(c)))
}

func (h *Handler) CreateSavedQuery(c *gin.Context) {
	var body SavedQuery
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" || body.Query == "" {
		writeError(c, http.StatusBadRequest, "name and query are required")
		return
	}
	if body.ID == "" {
		body.ID = "sq-" + uuid.New().String()[:8]
	}
	writeSuccess(c, h.deps.Mem.SaveSavedQuery(tenantID(c), body))
}

func (h *Handler) DeleteSavedQuery(c *gin.Context) {
	if !h.deps.Mem.DeleteSavedQuery(tenantID(c), c.Param("id")) {
		writeError(c, http.StatusNotFound, "saved query not found")
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

// QueryExplainResponse describes planner output without executing the query (NexQL M1).
type QueryExplainResponse struct {
	Valid        bool               `json:"valid"`
	Message      string             `json:"message,omitempty"`
	Steps        []QueryPlannerStep `json:"steps"`
	Stores       []string           `json:"stores"`
	JoinKeys     []string           `json:"joinKeys,omitempty"`
	Cardinality  CardinalityGuard   `json:"cardinality"`
	EstimatedMs  int                `json:"estimatedMs"`
}

func (h *Handler) ExplainUnifiedQuery(c *gin.Context) {
	var req UnifiedQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	writeSuccess(c, BuildQueryExplain(req))
}

func BuildQueryExplain(req UnifiedQueryRequest) QueryExplainResponse {
	q := strings.TrimSpace(req.Query)
	if q == "" {
		return QueryExplainResponse{Valid: false, Message: "query cannot be empty"}
	}
	steps := buildPlannerSteps(req)
	stores := plannerStores(req)
	joins := parseJoinKeys(q)
	_, guard := applyCardinalityGuard(nil, req)
	cost := 0
	for _, s := range steps {
		cost += s.EstimatedCost
	}
	return QueryExplainResponse{
		Valid:       true,
		Message:     "plan accepted",
		Steps:       steps,
		Stores:      stores,
		JoinKeys:    joins,
		Cardinality: guard,
		EstimatedMs: cost * 120,
	}
}

func plannerStores(req UnifiedQueryRequest) []string {
	from := strings.ToLower(strings.TrimSpace(req.From))
	switch from {
	case "logs":
		return []string{"elasticsearch", "clickhouse"}
	case "metrics":
		return []string{"prometheus", "clickhouse"}
	case "traces":
		return []string{"clickhouse", "tempo"}
	case "events":
		return []string{"kafka", "clickhouse"}
	default:
		return []string{"elasticsearch", "prometheus", "clickhouse", "kafka"}
	}
}

func parseJoinKeys(q string) []string {
	lower := strings.ToLower(q)
	var keys []string
	for _, key := range []string{"trace_id", "traceid", "txn_id", "txnid"} {
		if strings.Contains(lower, key) {
			normalized := key
			if key == "traceid" {
				normalized = "trace_id"
			}
			if key == "txnid" {
				normalized = "txn_id"
			}
			keys = append(keys, normalized)
		}
	}
	return keys
}
