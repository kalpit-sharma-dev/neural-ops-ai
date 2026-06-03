package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// FleetAgent is the gateway collector fleet representation.
type FleetAgent struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Environment     string `json:"environment"`
	Version         string `json:"version"`
	Status          string `json:"status"`
	LastHeartbeatAt string `json:"lastHeartbeatAt"`
	PolicyID        string `json:"policyId"`
}

// Client syncs collector agents to the NeuralOps gateway API.
type Client struct {
	baseURL string
	token   string
	tenant  string
	http    *http.Client
}

// NewClientFromEnv builds a client from NEURALOPS_API_URL and NEURALOPS_API_TOKEN.
func NewClientFromEnv() *Client {
	base := strings.TrimRight(os.Getenv("NEURALOPS_API_URL"), "/")
	if base == "" {
		base = "http://localhost:8080/api/v1"
	}
	tenant := os.Getenv("NEURALOPS_TENANT")
	if tenant == "" {
		tenant = "default"
	}
	return &Client{
		baseURL: base,
		token:   os.Getenv("NEURALOPS_API_TOKEN"),
		tenant:  tenant,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// SyncAgent upserts a collector agent; uses canary version when set.
func (c *Client) SyncAgent(ctx context.Context, agentID string, name, env, version, policyID, canary string) error {
	targetVersion := version
	if canary != "" {
		targetVersion = canary
	}
	body, _ := json.Marshal(map[string]any{
		"name":              name,
		"environment":       env,
		"version":           targetVersion,
		"status":            "healthy",
		"lastHeartbeatAt":   time.Now().UTC().Format(time.RFC3339),
		"policyId":          policyID,
	})
	putURL := fmt.Sprintf("%s/collectors/fleet/%s", c.baseURL, agentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("put agent %s: HTTP %d", agentID, resp.StatusCode)
	}
	postBody, _ := json.Marshal(map[string]any{
		"id": agentID, "name": name, "environment": env, "version": targetVersion,
		"status": "healthy", "lastHeartbeatAt": time.Now().UTC().Format(time.RFC3339),
		"policyId": policyID,
	})
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/collectors/fleet", bytes.NewReader(postBody))
	if err != nil {
		return err
	}
	c.applyHeaders(postReq)
	postResp, err := c.http.Do(postReq)
	if err != nil {
		return err
	}
	defer postResp.Body.Close()
	if postResp.StatusCode >= 300 {
		return fmt.Errorf("post agent %s: HTTP %d", agentID, postResp.StatusCode)
	}
	return nil
}

// UpgradeAgent triggers a rolling upgrade via gateway.
func (c *Client) UpgradeAgent(ctx context.Context, agentID, targetVersion string) error {
	body, _ := json.Marshal(map[string]string{"targetVersion": targetVersion})
	url := fmt.Sprintf("%s/collectors/fleet/%s/upgrade", c.baseURL, agentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upgrade agent %s: HTTP %d", agentID, resp.StatusCode)
	}
	return nil
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", c.tenant)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}
