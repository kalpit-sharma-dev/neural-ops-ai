package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
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
func (e *Executor) ExecuteWorkflow(ctx context.Context, tenantID, workflowID, triggerEvent string, ctxData map[string]string) (RunResult, error) {
	result := RunResult{
		RunID:     uuid.New().String(),
		Status:    "completed",
		StartedAt: time.Now().UTC(),
	}
	if e.pool == nil {
		return result, fmt.Errorf("postgres unavailable")
	}
	var stepsJSON []byte
	var graphJSON []byte
	var name string
	err := e.pool.QueryRow(ctx, `
SELECT name, steps, graph FROM observability_workflows WHERE tenant_id = $1 AND id = $2 AND enabled = true`,
		tenantID, workflowID).Scan(&name, &stepsJSON, &graphJSON)
	if err != nil {
		return result, fmt.Errorf("workflow not found")
	}

	runner := func(ctx context.Context, step Step) (string, error) {
		return e.runStep(ctx, tenantID, step, ctxData)
	}

	// Prefer the authored branching graph: run independent branches concurrently
	// with join + conditional-edge semantics. Fall back to the linear steps list
	// for legacy workflows that have no stored graph.
	if g, ok := parseGraph(graphJSON); ok {
		status, log := RunGraph(ctx, runner, g, ctxData)
		result.Status = status
		result.StepsLog = log
	} else {
		result.Status, result.StepsLog = runLinear(ctx, runner, parseSteps(stepsJSON))
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
	case "servicenow", "servicenow_incident", "itsm":
		return e.createServiceNowIncident(ctx, tenantID, step, ctxData)
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
		if strings.Contains(strings.ToLower(step.Label), "servicenow") || strings.Contains(strings.ToLower(step.Label), "snow") {
			return e.createServiceNowIncident(ctx, tenantID, step, ctxData)
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
	case strings.Contains(l, "servicenow"), strings.Contains(l, "snow"):
		return "servicenow"
	default:
		return "action"
	}
}

// StepRunner executes a single step and returns a human-readable log message.
// Injecting the runner keeps the orchestration logic pure and unit-testable.
type StepRunner func(ctx context.Context, step Step) (string, error)

// graphNode / graphEdge / graph mirror the persisted workflow graph JSON.
type graphNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type graphEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"`
}

type graph struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

// parseSteps decodes the steps JSONB column, supporting both the object form
// and the legacy string-array form.
func parseSteps(stepsJSON []byte) []Step {
	var steps []Step
	if len(stepsJSON) > 0 && stepsJSON[0] == '[' {
		if err := json.Unmarshal(stepsJSON, &steps); err != nil || hasEmptyStep(steps) {
			steps = nil
			var legacy []string
			_ = json.Unmarshal(stepsJSON, &legacy)
			for i, label := range legacy {
				steps = append(steps, Step{ID: fmt.Sprintf("s%d", i), Type: inferType(label), Label: label})
			}
		}
	}
	return steps
}

func hasEmptyStep(steps []Step) bool {
	for _, s := range steps {
		if s.Label == "" {
			return true
		}
	}
	return false
}

// parseGraph decodes the graph JSONB column, returning ok=false when no nodes
// are present so the caller can fall back to linear execution.
func parseGraph(graphJSON []byte) (graph, bool) {
	if len(graphJSON) == 0 {
		return graph{}, false
	}
	var g graph
	if err := json.Unmarshal(graphJSON, &g); err != nil {
		return graph{}, false
	}
	if len(g.Nodes) == 0 {
		return graph{}, false
	}
	return g, true
}

// runLinear executes steps sequentially, stopping on the first failure. This
// preserves the historical behavior for workflows without a branching graph.
func runLinear(ctx context.Context, run StepRunner, steps []Step) (string, []string) {
	status := "completed"
	log := make([]string, 0, len(steps))
	for _, step := range steps {
		msg, err := run(ctx, step)
		log = append(log, msg)
		if err != nil {
			status = "failed"
			log = append(log, err.Error())
			break
		}
	}
	return status, log
}

// node execution states for the DAG scheduler.
const (
	stPending = iota
	stDone
	stSkipped
)

type nodeState struct {
	node    graphNode
	state   int
	success bool
	message string
}

// RunGraph executes a workflow graph as a DAG: nodes whose predecessors have all
// resolved run concurrently (parallel branches), a node with multiple inbound
// edges joins (waits for every predecessor), and conditional edges gate whether
// a node runs or is skipped. Cyclic/unreachable nodes are skipped rather than
// deadlocking. The returned log is ordered by execution level for determinism.
func RunGraph(ctx context.Context, run StepRunner, g graph, ctxData map[string]string) (string, []string) {
	states := make(map[string]*nodeState, len(g.Nodes))
	order := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if states[n.ID] != nil {
			continue
		}
		states[n.ID] = &nodeState{node: n, state: stPending}
		order = append(order, n.ID)
	}

	inbound := make(map[string][]graphEdge, len(g.Nodes))
	for _, ed := range g.Edges {
		if states[ed.Source] == nil || states[ed.Target] == nil {
			continue
		}
		inbound[ed.Target] = append(inbound[ed.Target], ed)
	}

	resolved := func(id string) bool {
		s := states[id]
		return s != nil && (s.state == stDone || s.state == stSkipped)
	}

	status := "completed"
	logBuf := make([]string, 0, len(g.Nodes)*2)

	for {
		pending := 0
		ready := make([]*nodeState, 0)
		for _, id := range order {
			s := states[id]
			if s.state != stPending {
				continue
			}
			pending++
			allResolved := true
			for _, ed := range inbound[id] {
				if !resolved(ed.Source) {
					allResolved = false
					break
				}
			}
			if allResolved {
				ready = append(ready, s)
			}
		}
		if pending == 0 {
			break
		}
		if len(ready) == 0 {
			// No progress possible (cycle / orphaned join): skip the remainder.
			for _, id := range order {
				if states[id].state == stPending {
					states[id].state = stSkipped
					logBuf = append(logBuf, "Skipped (unreachable): "+states[id].node.Label)
				}
			}
			break
		}

		// Decide which ready nodes run vs. skip based on inbound edge conditions.
		toRun := make([]*nodeState, 0, len(ready))
		for _, s := range ready {
			edges := inbound[s.node.ID]
			if len(edges) == 0 || anyEdgeSatisfied(edges, states, ctxData) {
				toRun = append(toRun, s)
			} else {
				s.state = stSkipped
				logBuf = append(logBuf, "Skipped: "+s.node.Label)
			}
		}

		// Execute this level's runnable nodes concurrently, then join.
		var wg sync.WaitGroup
		for _, s := range toRun {
			wg.Add(1)
			go func(s *nodeState) {
				defer wg.Done()
				msg, err := run(ctx, Step{ID: s.node.ID, Type: s.node.Type, Label: s.node.Label})
				s.success = err == nil
				if err != nil {
					if msg == "" {
						msg = "error: " + err.Error()
					} else {
						msg = msg + " | error: " + err.Error()
					}
				}
				s.message = msg
			}(s)
		}
		wg.Wait()

		for _, s := range toRun {
			s.state = stDone
			if !s.success {
				status = "failed"
			}
			logBuf = append(logBuf, s.message)
		}
	}

	return status, logBuf
}

// anyEdgeSatisfied reports whether at least one inbound edge permits the target
// to run, given its predecessors' outcomes and the trigger context.
func anyEdgeSatisfied(edges []graphEdge, states map[string]*nodeState, ctxData map[string]string) bool {
	for _, ed := range edges {
		src := states[ed.Source]
		if src == nil || src.state != stDone {
			continue // a skipped/unrun predecessor never satisfies its edge
		}
		switch cond := strings.ToLower(strings.TrimSpace(ed.Condition)); cond {
		case "", "success":
			if src.success {
				return true
			}
		case "always":
			return true
		case "failure":
			if !src.success {
				return true
			}
		default:
			if !src.success {
				continue
			}
			key, val, ok := splitKV(cond)
			if !ok {
				return true // malformed condition is treated as success-gated
			}
			if strings.EqualFold(strings.TrimSpace(ctxData[key]), val) {
				return true
			}
		}
	}
	return false
}

// splitKV parses "key=value" or "key:value" conditions.
func splitKV(cond string) (string, string, bool) {
	for _, sep := range []string{"=", ":"} {
		if i := strings.Index(cond, sep); i > 0 {
			return strings.TrimSpace(cond[:i]), strings.TrimSpace(cond[i+1:]), true
		}
	}
	return "", "", false
}
