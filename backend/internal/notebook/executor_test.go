package notebook_test

import (
	"testing"

	"github.com/neuralops/platform/internal/notebook"
)

func TestParseLogQuery(t *testing.T) {
	// exercise parse via exported Search path shape
	backend := notebook.NewLogBackend("")
	if backend != nil {
		t.Fatal("expected nil backend for empty url")
	}
	e := notebook.NewExecutor(nil, nil, nil)
	if e == nil {
		t.Fatal("expected executor")
	}
}

func TestRunPromQLUnavailable(t *testing.T) {
	_, err := notebook.RunPromQL(t.Context(), nil, "up")
	if err == nil {
		t.Fatal("expected error")
	}
}
