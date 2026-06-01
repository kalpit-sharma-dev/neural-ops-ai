package apm

import "time"

// Span represents a distributed trace span.
type Span struct {
	TraceID    string            `json:"traceId"`
	SpanID     string            `json:"spanId"`
	ParentID   string            `json:"parentId,omitempty"`
	Service    string            `json:"service"`
	Operation  string            `json:"operation"`
	StartTime  time.Time         `json:"startTime"`
	DurationMs int64             `json:"durationMs"`
	Status     string            `json:"status"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// TraceDetail is a full trace with span tree metadata.
type TraceDetail struct {
	TraceID   string `json:"traceId"`
	RootSpan  string `json:"rootSpanId"`
	Spans     []Span `json:"spans"`
	TotalMs   int64  `json:"totalMs"`
	Service   string `json:"service"`
	Status    string `json:"status"`
	SpanCount int    `json:"spanCount"`
}

// TraceSearchRequest filters trace search.
type TraceSearchRequest struct {
	Service   string    `json:"service"`
	Operation string    `json:"operation"`
	Status    string    `json:"status"`
	MinMs     int64     `json:"minDurationMs"`
	MaxMs     int64     `json:"maxDurationMs"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Limit     int       `json:"limit"`
}

// TraceSummary is a list item for trace search results.
type TraceSummary struct {
	TraceID    string    `json:"traceId"`
	Service    string    `json:"service"`
	Operation  string    `json:"operation"`
	DurationMs int64     `json:"durationMs"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"startTime"`
	SpanCount  int       `json:"spanCount"`
}

// FlowEdge is a service-flow aggregation edge.
type FlowEdge struct {
	Source    string  `json:"source"`
	Target    string  `json:"target"`
	CallCount int64   `json:"callCount"`
	ErrorRate float64 `json:"errorRate"`
	P50Ms     float64 `json:"p50Ms"`
	P95Ms     float64 `json:"p95Ms"`
}
