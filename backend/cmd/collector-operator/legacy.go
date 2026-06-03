package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type desiredAgent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	PolicyID    string `json:"policyId"`
}

type reconcileSpec struct {
	Agents []desiredAgent `json:"agents"`
}

func runLegacy(specPath string) {
	if specPath == "" {
		log.Fatal("COLLECTOR_RECONCILE_SPEC or -spec is required in legacy mode")
	}
	apiBase := flag.Lookup("api")
	base := "http://localhost:8080/api/v1"
	if apiBase != nil {
		// not registered in legacy path
	}
	if v := os.Getenv("NEURALOPS_API_URL"); v != "" {
		base = strings.TrimRight(v, "/")
	}
	interval := 30 * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	reconcile := func() {
		raw, err := os.ReadFile(specPath)
		if err != nil {
			log.Printf("read spec: %v", err)
			return
		}
		var spec reconcileSpec
		if err := json.Unmarshal(raw, &spec); err != nil {
			log.Printf("parse spec: %v", err)
			return
		}
		for _, agent := range spec.Agents {
			if err := legacyUpsert(ctx, base, agent); err != nil {
				log.Printf("reconcile %s: %v", agent.ID, err)
			}
		}
	}
	reconcile()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcile()
		}
	}
}

func legacyUpsert(ctx context.Context, base string, agent desiredAgent) error {
	body, _ := json.Marshal(map[string]any{
		"name": agent.Name, "environment": agent.Environment, "version": agent.Version,
		"status": "healthy", "lastHeartbeatAt": time.Now().UTC().Format(time.RFC3339),
		"policyId": agent.PolicyID,
	})
	url := strings.TrimRight(base, "/") + "/collectors/fleet/" + agent.ID
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "default")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}
