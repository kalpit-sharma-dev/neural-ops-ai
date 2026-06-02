package workflow

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// recordingRunner returns a StepRunner that records invocations and can fail
// steps whose label is in failLabels.
func recordingRunner(failLabels ...string) (StepRunner, func() []string, *int32) {
	var mu sync.Mutex
	var maxConcurrent int32
	var current int32
	calls := make([]string, 0)
	failSet := make(map[string]bool, len(failLabels))
	for _, l := range failLabels {
		failSet[l] = true
	}
	run := func(ctx context.Context, step Step) (string, error) {
		c := atomic.AddInt32(&current, 1)
		for {
			old := atomic.LoadInt32(&maxConcurrent)
			if c <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, c) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond) // widen the window so parallel steps overlap
		atomic.AddInt32(&current, -1)
		mu.Lock()
		calls = append(calls, step.Label)
		mu.Unlock()
		if failSet[step.Label] {
			return "ran " + step.Label, errContext("boom")
		}
		return "ran " + step.Label, nil
	}
	snapshot := func() []string {
		mu.Lock()
		defer mu.Unlock()
		out := append([]string(nil), calls...)
		sort.Strings(out)
		return out
	}
	return run, snapshot, &maxConcurrent
}

type errString string

func (e errString) Error() string { return string(e) }
func errContext(s string) error   { return errString(s) }

func TestRunGraph_ParallelBranchesAndJoin(t *testing.T) {
	g := graph{
		Nodes: []graphNode{
			{ID: "a", Label: "start"},
			{ID: "b", Label: "branch-1"},
			{ID: "c", Label: "branch-2"},
			{ID: "d", Label: "join"},
		},
		Edges: []graphEdge{
			{Source: "a", Target: "b"},
			{Source: "a", Target: "c"},
			{Source: "b", Target: "d"},
			{Source: "c", Target: "d"},
		},
	}
	run, snapshot, maxConcurrent := recordingRunner()
	status, _ := RunGraph(context.Background(), run, g, nil)
	if status != "completed" {
		t.Fatalf("expected completed, got %s", status)
	}
	got := snapshot()
	want := []string{"branch-1", "branch-2", "join", "start"}
	if len(got) != len(want) {
		t.Fatalf("expected all 4 steps to run, got %v", got)
	}
	if *maxConcurrent < 2 {
		t.Fatalf("expected branch-1 and branch-2 to run concurrently, max concurrency was %d", *maxConcurrent)
	}
}

func TestRunGraph_FailureEdgeBranches(t *testing.T) {
	g := graph{
		Nodes: []graphNode{
			{ID: "a", Label: "deploy"},
			{ID: "ok", Label: "notify-success"},
			{ID: "bad", Label: "rollback"},
		},
		Edges: []graphEdge{
			{Source: "a", Target: "ok", Condition: "success"},
			{Source: "a", Target: "bad", Condition: "failure"},
		},
	}
	run, snapshot, _ := recordingRunner("deploy")
	status, _ := RunGraph(context.Background(), run, g, nil)
	if status != "failed" {
		t.Fatalf("expected failed, got %s", status)
	}
	got := snapshot()
	// deploy failed -> rollback should run, notify-success should be skipped.
	hasRollback, hasNotify := false, false
	for _, c := range got {
		if c == "rollback" {
			hasRollback = true
		}
		if c == "notify-success" {
			hasNotify = true
		}
	}
	if !hasRollback {
		t.Fatalf("expected rollback to run on failure, got %v", got)
	}
	if hasNotify {
		t.Fatalf("expected notify-success to be skipped, got %v", got)
	}
}

func TestRunGraph_SkipPropagatesThroughJoin(t *testing.T) {
	// b is skipped (only reachable via a's failure edge), so the join d, which
	// depends on b succeeding, must also be skipped.
	g := graph{
		Nodes: []graphNode{
			{ID: "a", Label: "a"},
			{ID: "b", Label: "b"},
			{ID: "d", Label: "d"},
		},
		Edges: []graphEdge{
			{Source: "a", Target: "b", Condition: "failure"},
			{Source: "b", Target: "d", Condition: "success"},
		},
	}
	run, snapshot, _ := recordingRunner()
	status, _ := RunGraph(context.Background(), run, g, nil)
	if status != "completed" {
		t.Fatalf("expected completed, got %s", status)
	}
	for _, c := range snapshot() {
		if c == "b" || c == "d" {
			t.Fatalf("expected b and d to be skipped, got %v", snapshot())
		}
	}
}

func TestRunGraph_ContextCondition(t *testing.T) {
	g := graph{
		Nodes: []graphNode{
			{ID: "a", Label: "classify"},
			{ID: "p1", Label: "page-oncall"},
		},
		Edges: []graphEdge{
			{Source: "a", Target: "p1", Condition: "severity=critical"},
		},
	}
	run, snapshot, _ := recordingRunner()
	status, _ := RunGraph(context.Background(), run, g, map[string]string{"severity": "critical"})
	if status != "completed" {
		t.Fatalf("expected completed, got %s", status)
	}
	found := false
	for _, c := range snapshot() {
		if c == "page-oncall" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected page-oncall to run when severity=critical, got %v", snapshot())
	}

	run2, snapshot2, _ := recordingRunner()
	RunGraph(context.Background(), run2, g, map[string]string{"severity": "low"})
	for _, c := range snapshot2() {
		if c == "page-oncall" {
			t.Fatalf("expected page-oncall to be skipped when severity=low, got %v", snapshot2())
		}
	}
}

func TestRunGraph_CycleDoesNotDeadlock(t *testing.T) {
	g := graph{
		Nodes: []graphNode{
			{ID: "a", Label: "a"},
			{ID: "b", Label: "b"},
		},
		Edges: []graphEdge{
			{Source: "a", Target: "b"},
			{Source: "b", Target: "a"},
		},
	}
	run, _, _ := recordingRunner()
	done := make(chan string, 1)
	go func() {
		status, _ := RunGraph(context.Background(), run, g, nil)
		done <- status
	}()
	select {
	case <-done:
		// completed without hanging
	case <-time.After(2 * time.Second):
		t.Fatal("RunGraph deadlocked on a cyclic graph")
	}
}
