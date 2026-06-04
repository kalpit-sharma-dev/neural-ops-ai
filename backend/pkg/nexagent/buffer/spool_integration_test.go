package buffer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/neuralops/platform/pkg/nexagent/buffer"
)

func TestSpoolReplayDrain(t *testing.T) {
	dir := t.TempDir()
	spool := buffer.NewDiskSpool(dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	if err := spool.Enqueue(map[string]any{"level": "INFO", "message": "offline spool event"}); err != nil {
		t.Fatal(err)
	}
	n, err := spool.Replay(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 replayed, got %d", n)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() && e.Name() != "." {
			t.Fatalf("spool should be drained, leftover %s", e.Name())
		}
	}
}
