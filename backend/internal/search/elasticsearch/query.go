package elasticsearch

import (
	"strings"
	"time"

	"github.com/neuralops/platform/internal/search/dto"
)

// BuildQuery constructs an Elasticsearch query DSL body.
func BuildQuery(req dto.LogSearchRequest, includeAggs bool) map[string]any {
	size := req.Size
	if size <= 0 {
		size = 50
	}

	must := make([]map[string]any, 0)
	filter := make([]map[string]any, 0)

	if strings.TrimSpace(req.Query) != "" {
		must = append(must, map[string]any{
			"match": map[string]any{
				"message": map[string]any{
					"query":    req.Query,
					"operator": "and",
				},
			},
		})
	}

	if req.MessageRegex != "" {
		must = append(must, map[string]any{
			"regexp": map[string]any{
				"message": req.MessageRegex,
			},
		})
	}

	addTermFilter(&filter, "service", req.Service)
	if len(req.AllowedServices) > 0 && strings.TrimSpace(req.Service) == "" {
		filter = append(filter, map[string]any{
			"terms": map[string]any{"service": req.AllowedServices},
		})
	}
	addTermFilter(&filter, "severity", strings.ToUpper(req.Severity))
	addTermFilter(&filter, "traceId", req.TraceID)
	addTermFilter(&filter, "txnId", req.TxnID)
	addTermFilter(&filter, "host", req.Host)
	addTermFilter(&filter, "pod", req.Pod)
	addTermFilter(&filter, "classification", strings.ToUpper(req.Classification))
	addTermFilter(&filter, "tenantId", req.TenantID)

	for key, value := range req.Labels {
		filter = append(filter, map[string]any{
			"term": map[string]any{
				"labels." + key: value,
			},
		})
	}

	if req.StartTime != nil || req.EndTime != nil {
		rangeFilter := map[string]any{}
		if req.StartTime != nil {
			rangeFilter["gte"] = req.StartTime.UTC().Format(time.RFC3339Nano)
		}
		if req.EndTime != nil {
			rangeFilter["lte"] = req.EndTime.UTC().Format(time.RFC3339Nano)
		}
		filter = append(filter, map[string]any{
			"range": map[string]any{"timestamp": rangeFilter},
		})
	}

	boolQuery := map[string]any{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}
	if len(boolQuery) == 0 {
		boolQuery["must"] = []map[string]any{{"match_all": map[string]any{}}}
	}

	body := map[string]any{
		"size": size,
		"sort": []map[string]any{
			{"timestamp": map[string]string{"order": "desc"}},
			{"_id": map[string]string{"order": "desc"}},
		},
		"query": map[string]any{"bool": boolQuery},
	}

	if len(req.SearchAfter) > 0 {
		body["search_after"] = req.SearchAfter
	}

	if includeAggs {
		body["aggs"] = map[string]any{
			"by_service": map[string]any{
				"terms": map[string]any{"field": "service", "size": 20},
			},
			"by_severity": map[string]any{
				"terms": map[string]any{"field": "severity", "size": 10},
			},
			"by_hour": map[string]any{
				"date_histogram": map[string]any{
					"field":             "timestamp",
					"calendar_interval": "1h",
				},
			},
		}
	}

	return body
}

func addTermFilter(filter *[]map[string]any, field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	*filter = append(*filter, map[string]any{
		"term": map[string]any{field: value},
	})
}

// BuildTraceQuery builds a query for all logs in a trace.
func BuildTraceQuery(traceID, tenantID string, size int) map[string]any {
	if size <= 0 {
		size = 200
	}
	filter := []map[string]any{{"term": map[string]any{"traceId": traceID}}}
	if tenantID != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"tenantId": tenantID}})
	}
	return map[string]any{
		"size": size,
		"sort": []map[string]any{
			{"timestamp": map[string]string{"order": "asc"}},
		},
		"query": map[string]any{
			"bool": map[string]any{"filter": filter},
		},
	}
}
