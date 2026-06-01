package observability

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/notebook"
)

// promNotebookAdapter adapts PromQLClient to notebook.PromQuerier.
type promNotebookAdapter struct {
	client *PromQLClient
}

func (a promNotebookAdapter) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]notebook.MetricPoint, error) {
	pts, err := a.client.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, err
	}
	out := make([]notebook.MetricPoint, len(pts))
	for i, p := range pts {
		out[i] = notebook.MetricPoint{Timestamp: p.Timestamp, Value: p.Value}
	}
	return out, nil
}

func (a promNotebookAdapter) QueryInstant(ctx context.Context, query string) (float64, error) {
	return a.client.QueryInstant(ctx, query)
}

// NotebookPromQuerier returns a PromQuerier for notebook execution.
func NotebookPromQuerier(client *PromQLClient) notebook.PromQuerier {
	if client == nil {
		return nil
	}
	return promNotebookAdapter{client: client}
}
