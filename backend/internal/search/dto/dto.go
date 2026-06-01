package dto

import "time"

// LogSearchRequest is a structured log search request.
type LogSearchRequest struct {
	Query        string            `json:"query"`
	MessageRegex string            `json:"messageRegex"`
	Service      string            `json:"service"`
	Severity     string            `json:"severity"`
	TraceID      string            `json:"traceId"`
	TxnID        string            `json:"txnId"`
	Host         string            `json:"host"`
	Pod          string            `json:"pod"`
	Classification string          `json:"classification"`
	StartTime      *time.Time        `json:"startTime"`
	EndTime        *time.Time        `json:"endTime"`
	Labels         map[string]string `json:"labels"`
	TenantID       string            `json:"tenantId"`
	Size           int               `json:"size"`
	SearchAfter  []any             `json:"searchAfter"`
	AllowedServices []string       `json:"-"`
}

// SemanticSearchRequest is a natural language / vector search request.
type SemanticSearchRequest struct {
	Query       string     `json:"query"`
	Hybrid      bool       `json:"hybrid"`
	Size        int        `json:"size"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	TenantID    string     `json:"tenantId"`
	SearchAfter []any      `json:"searchAfter"`
}

// AISearchRequest is a natural language search request.
type AISearchRequest struct {
	Question string `json:"question"`
	Size     int    `json:"size"`
}

// TransactionSearchRequest filters transaction journeys.
type TransactionSearchRequest struct {
	TxnID        string     `json:"txnId"`
	TxnType      string     `json:"txnType"`
	Status       string     `json:"status"`
	FailedService string    `json:"failedService"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	Limit        int        `json:"limit"`
}

// LogHit represents a single log search result.
type LogHit struct {
	ID             string            `json:"id"`
	Score          float64           `json:"score"`
	Timestamp      time.Time         `json:"timestamp"`
	Service        string            `json:"service"`
	Severity       string            `json:"severity"`
	Message        string            `json:"message"`
	TraceID        string            `json:"traceId,omitempty"`
	TxnID          string            `json:"txnId,omitempty"`
	Host           string            `json:"host,omitempty"`
	Pod            string            `json:"pod,omitempty"`
	Classification string            `json:"classification,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	PlainEnglish   string            `json:"plainEnglish,omitempty"`
	Source         string            `json:"source,omitempty"`
}

// SearchResponse is a paginated search response.
type SearchResponse struct {
	Hits        []LogHit       `json:"hits"`
	Total       int64          `json:"total"`
	SearchAfter []any          `json:"searchAfter,omitempty"`
	Aggregations *Aggregations `json:"aggregations,omitempty"`
	TookMs      int64          `json:"tookMs"`
}

// Aggregations holds search aggregation buckets.
type Aggregations struct {
	ByService  []Bucket `json:"byService"`
	BySeverity []Bucket `json:"bySeverity"`
	ByHour     []Bucket `json:"byHour"`
}

// Bucket is an aggregation bucket.
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// ParsedQuery is extracted from natural language by the AI parser.
type ParsedQuery struct {
	Query        string     `json:"query"`
	Service      string     `json:"service"`
	Severity     string     `json:"severity"`
	Classification string   `json:"classification"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	Explanation  string     `json:"explanation"`
}
