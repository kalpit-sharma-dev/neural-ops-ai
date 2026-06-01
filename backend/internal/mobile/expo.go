package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

// PushMessage is one Expo push notification.
type PushMessage struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Sound string `json:"sound,omitempty"`
}

// ExpoClient sends push notifications via Expo.
type ExpoClient struct {
	pool       *pgxpool.Pool
	httpClient *http.Client
}

// NewExpoClient creates an Expo push client.
func NewExpoClient(pool *pgxpool.Pool) *ExpoClient {
	return &ExpoClient{
		pool:       pool,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// ListTokens returns Expo tokens for a tenant.
func (c *ExpoClient) ListTokens(ctx context.Context, tenantID string) ([]string, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	rows, err := c.pool.Query(ctx, `
SELECT expo_push_token FROM mobile_push_tokens WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err == nil && token != "" {
			out = append(out, token)
		}
	}
	return out, rows.Err()
}

// SendToTenant pushes a notification to all registered devices for a tenant.
func (c *ExpoClient) SendToTenant(ctx context.Context, tenantID, title, body string) error {
	tokens, err := c.ListTokens(ctx, tenantID)
	if err != nil || len(tokens) == 0 {
		return err
	}
	msgs := make([]PushMessage, 0, len(tokens))
	for _, t := range tokens {
		msgs = append(msgs, PushMessage{To: t, Title: title, Body: body, Sound: "default"})
	}
	raw, _ := json.Marshal(msgs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("expo push: %s", resp.Status)
	}
	return nil
}
