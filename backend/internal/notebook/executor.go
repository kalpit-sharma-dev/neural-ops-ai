package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CellResult is the output of executing one notebook cell.
type CellResult struct {
	CellID     string    `json:"cellId"`
	Type       string    `json:"type"`
	Output     string    `json:"output"`
	RanAt      time.Time `json:"ranAt"`
	DurationMs int64     `json:"durationMs"`
}

// Executor runs notebook cells.
type Executor struct {
	pool   *pgxpool.Pool
	logs   *LogBackend
	prom   PromQuerier
}

// NewExecutor creates a notebook executor.
func NewExecutor(pool *pgxpool.Pool, logs *LogBackend, prom PromQuerier) *Executor {
	return &Executor{pool: pool, logs: logs, prom: prom}
}

// ExecuteNotebook runs all cells in a notebook.
func (e *Executor) ExecuteNotebook(ctx context.Context, tenantID, notebookID string) ([]CellResult, error) {
	if e.pool == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	var cellsJSON []byte
	err := e.pool.QueryRow(ctx, `
SELECT cells FROM observability_notebooks WHERE tenant_id = $1 AND id = $2`, tenantID, notebookID).Scan(&cellsJSON)
	if err != nil {
		return nil, fmt.Errorf("notebook not found")
	}
	var cells []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(cellsJSON, &cells); err != nil {
		return nil, err
	}
	results := make([]CellResult, 0, len(cells))
	for _, cell := range cells {
		start := time.Now()
		out, err := e.runCell(ctx, tenantID, cell.Type, cell.Content)
		res := CellResult{
			CellID:     cell.ID,
			Type:       cell.Type,
			Output:     out,
			RanAt:      start.UTC(),
			DurationMs: time.Since(start).Milliseconds(),
		}
		if err != nil {
			res.Output = "ERROR: " + err.Error()
		}
		results = append(results, res)
	}
	return results, nil
}

func (e *Executor) runCell(ctx context.Context, tenantID, cellType, content string) (string, error) {
	switch strings.ToLower(cellType) {
	case "markdown":
		return content, nil
	case "query", "log":
		if e.logs == nil {
			return "", fmt.Errorf("log search backend unavailable")
		}
		return e.logs.Search(ctx, tenantID, content)
	case "promql", "metric":
		return RunPromQL(ctx, e.prom, content)
	default:
		return content, nil
	}
}
