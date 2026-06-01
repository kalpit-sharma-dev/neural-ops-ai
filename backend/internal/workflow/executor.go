package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Step defines one workflow action.
type Step struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Label  string            `json:"label"`
	Config map[string]string `json:"config,omitempty"`
}

// RunResult is the outcome of a workflow execution.
type RunResult struct {
	RunID     string   `json:"runId"`
	Status    string   `json:"status"`
	StepsLog  []string `json:"stepsLog"`
	StartedAt time.Time `json:"startedAt"`
}

// Executor runs workflow steps against external integrations.
type Executor struct {
	pool *pgxpool.Pool
}

// NewExecutor creates a workflow executor.
func NewExecutor(pool *pgxpool.Pool) *Executor {
	return &Executor{pool: pool}
}

// ExecuteWorkflow runs all steps for a workflow ID.
func (e *Executor) ExecuteWorkflow(ctx context.Context, tenantID, workflowID, triggerEvent string, context map[string]string) (RunResult, error) {
	result := RunResult{
		RunID:     uuid.New().String(),
		Status:    "completed",
		StartedAt: time.Now().UTC(),
	}
	if e.pool == nil {
		return result, fmt.Errorf("postgres unavailable")
	}
	var stepsJSON []byte
	var name string
	err := e.pool.QueryRow(ctx, `
SELECT name, steps FROM observability_workflows WHERE tenant_id = $1 AND id = $2 AND enabled = true`,
		tenantID, workflowID).Scan(&name, &stepsJSON)
	if err != nil {
		return result, fmt.Errorf("workflow not found")
	}

	var steps []Step
	if len(stepsJSON) > 0 && stepsJSON[0] == '[' {
		// New format: array of Step objects
		if err := json.Unmarshal(stepsJSON, &steps); err != nil {
			// Legacy: string array
			var legacy []string
			_ = json.Unmarshal(stepsJSON, &legacy)
			for i, label := range legacy {
				steps = append(steps, Step{ID: fmt.Sprintf("s%d", i), Type: inferType(label), Label: label})
			}
		}
	}

	for _, step := range steps {
		msg, stepErr := e.runStep(ctx, tenantID, step, context)
		result.StepsLog = append(result.StepsLog, msg)
		if stepErr != nil {
			result.Status = "failed"
			result.StepsLog = append(result.StepsLog, stepErr.Error())
			break
		}
	}

	logJSON, _ := json.Marshal(result.StepsLog)
	finished := time.Now().UTC()
	_, _ = e.pool.Exec(ctx, `
INSERT INTO observability_workflow_runs (id, tenant_id, workflow_id, trigger_event, status, steps_log, started_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		result.RunID, tenantID, workflowID, triggerEvent, result.Status, logJSON, result.StartedAt, finished)

	return result, nil
}

// ExecuteByTrigger finds workflows matching trigger and runs them.
func (e *Executor) ExecuteByTrigger(ctx context.Context, tenantID, trigger string, context map[string]string) ([]RunResult, error) {
	if e.pool == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	rows, err := e.pool.Query(ctx, `
SELECT id::text FROM observability_workflows WHERE tenant_id = $1 AND trigger_name = $2 AND enabled = true`,
		tenantID, trigger)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []RunResult
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return results, err
		}
		r, err := e.ExecuteWorkflow(ctx, tenantID, id, trigger, context)
		if err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (e *Executor) runStep(ctx context.Context, tenantID string, step Step, ctxData map[string]string) (string, error) {
	switch strings.ToLower(step.Type) {
	case "jira", "jira_ticket":
		return e.createJiraTicketImpl(ctx, tenantID, step, ctxData)
	case "slack", "notify_slack":
		return e.notifySlack(ctx, tenantID, step, ctxData)
	case "pagerduty", "page":
		return e.notifyPagerDuty(ctx, tenantID, step, ctxData)
	default:
		if strings.Contains(strings.ToLower(step.Label), "jira") {
			return e.createJiraTicketImpl(ctx, tenantID, step, ctxData)
		}
		if strings.Contains(strings.ToLower(step.Label), "slack") {
			return e.notifySlack(ctx, tenantID, step, ctxData)
		}
		if strings.Contains(strings.ToLower(step.Label), "pager") {
			return e.notifyPagerDuty(ctx, tenantID, step, ctxData)
		}
		return fmt.Sprintf("Executed: %s", step.Label), nil
	}
}

func inferType(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "jira"):
		return "jira"
	case strings.Contains(l, "slack"):
		return "slack"
	case strings.Contains(l, "pager"):
		return "pagerduty"
	default:
		return "action"
	}
}
