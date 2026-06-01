package notifier

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/domain"
)

type emailNotifier struct {
	host     string
	port     string
	username string
	password string
	from     string
	to       []string
}

// NewEmailNotifier creates an SMTP email notifier.
func NewEmailNotifier(config map[string]any) (Notifier, error) {
	toRaw, _ := config["to"].([]any)
	to := make([]string, 0, len(toRaw))
	for _, item := range toRaw {
		if value, ok := item.(string); ok {
			to = append(to, value)
		}
	}
	return &emailNotifier{
		host:     stringValue(config, "host", "localhost"),
		port:     stringValue(config, "port", "587"),
		username: stringValue(config, "username", ""),
		password: stringValue(config, "password", ""),
		from:     stringValue(config, "from", "alerts@neuralops.io"),
		to:       to,
	}, nil
}

func (n *emailNotifier) Type() string { return "email" }

func (n *emailNotifier) Send(_ context.Context, alert model.AlertRecord, incident *domain.Incident) error {
	if len(n.to) == 0 {
		return fmt.Errorf("email notifier missing recipients")
	}
	subject := fmt.Sprintf("[%s] %s", alert.Severity, alert.Title)
	body := renderHTML(alert, incident)
	msg := bytes.NewBufferString(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		n.from, strings.Join(n.to, ","), subject, body))
	addr := net.JoinHostPort(n.host, n.port)
	auth := smtp.PlainAuth("", n.username, n.password, n.host)
	return smtp.SendMail(addr, auth, n.from, n.to, msg.Bytes())
}

type slackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a Slack incoming webhook notifier.
func NewSlackNotifier(config map[string]any) (Notifier, error) {
	url := stringValue(config, "webhook_url", "")
	if url == "" {
		return nil, fmt.Errorf("slack webhook_url required")
	}
	return &slackNotifier{
		webhookURL: url,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (n *slackNotifier) Type() string { return "slack" }

func (n *slackNotifier) Send(ctx context.Context, alert model.AlertRecord, incident *domain.Incident) error {
	payload := map[string]any{
		"text": fmt.Sprintf("*[%s]* %s", alert.Severity, alert.Title),
		"blocks": []map[string]any{
			{"type": "section", "text": map[string]string{"type": "mrkdwn", "text": fmt.Sprintf("*%s*\n%s\nService: `%s`", alert.Title, alert.Description, alert.Service)}},
		},
	}
	if incident != nil {
		payload["text"] = fmt.Sprintf("%s (incident: %s)", payload["text"], incident.Title)
	}
	return postJSON(ctx, n.client, n.webhookURL, payload)
}

type teamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsNotifier creates a Microsoft Teams notifier.
func NewTeamsNotifier(config map[string]any) (Notifier, error) {
	url := stringValue(config, "webhook_url", "")
	if url == "" {
		return nil, fmt.Errorf("teams webhook_url required")
	}
	return &teamsNotifier{webhookURL: url, client: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (n *teamsNotifier) Type() string { return "teams" }

func (n *teamsNotifier) Send(ctx context.Context, alert model.AlertRecord, incident *domain.Incident) error {
	payload := map[string]any{
		"type": "message",
		"attachments": []map[string]any{{
			"contentType": "application/vnd.microsoft.card.adaptive",
			"content": map[string]any{
				"type":    "AdaptiveCard",
				"version": "1.4",
				"body": []map[string]any{
					{"type": "TextBlock", "size": "Medium", "weight": "Bolder", "text": alert.Title},
					{"type": "TextBlock", "text": alert.Description, "wrap": true},
					{"type": "TextBlock", "text": fmt.Sprintf("Severity: %s | Service: %s", alert.Severity, alert.Service)},
				},
			},
		}},
	}
	if incident != nil {
		payload["summary"] = incident.Title
	}
	return postJSON(ctx, n.client, n.webhookURL, payload)
}

type pagerDutyNotifier struct {
	routingKey string
	client     *http.Client
}

// NewPagerDutyNotifier creates a PagerDuty Events API v2 notifier.
func NewPagerDutyNotifier(config map[string]any) (Notifier, error) {
	key := stringValue(config, "routing_key", "")
	if key == "" {
		return nil, fmt.Errorf("pagerduty routing_key required")
	}
	return &pagerDutyNotifier{
		routingKey: key,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (n *pagerDutyNotifier) Type() string { return "pagerduty" }

func (n *pagerDutyNotifier) Send(ctx context.Context, alert model.AlertRecord, _ *domain.Incident) error {
	payload := map[string]any{
		"routing_key":  n.routingKey,
		"event_action": "trigger",
		"dedup_key":    alert.Fingerprint,
		"payload": map[string]any{
			"summary":  alert.Title,
			"severity": strings.ToLower(string(alert.Severity)),
			"source":   alert.Service,
			"custom_details": map[string]any{
				"description": alert.Description,
				"alertName":   alert.AlertName,
			},
		},
	}
	return postJSON(ctx, n.client, "https://events.pagerduty.com/v2/enqueue", payload)
}

type smsNotifier struct {
	accountSID string
	authToken  string
	from       string
	to         []string
	client     *http.Client
}

// NewSMSNotifier creates a Twilio SMS notifier.
func NewSMSNotifier(config map[string]any) (Notifier, error) {
	toRaw, _ := config["to"].([]any)
	to := make([]string, 0, len(toRaw))
	for _, item := range toRaw {
		if value, ok := item.(string); ok {
			to = append(to, value)
		}
	}
	return &smsNotifier{
		accountSID: stringValue(config, "account_sid", ""),
		authToken:  stringValue(config, "auth_token", ""),
		from:       stringValue(config, "from", ""),
		to:         to,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (n *smsNotifier) Type() string { return "sms" }

func (n *smsNotifier) Send(ctx context.Context, alert model.AlertRecord, _ *domain.Incident) error {
	if n.accountSID == "" || n.authToken == "" || n.from == "" || len(n.to) == 0 {
		return fmt.Errorf("twilio sms notifier not fully configured")
	}
	message := fmt.Sprintf("[%s] %s - %s", alert.Severity, alert.Service, alert.Title)
	for _, recipient := range n.to {
		form := fmt.Sprintf("To=%s&From=%s&Body=%s", recipient, n.from, message)
		url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", n.accountSID)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(form))
		if err != nil {
			return err
		}
		req.SetBasicAuth(n.accountSID, n.authToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := n.client.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
		if res.StatusCode >= 400 {
			return fmt.Errorf("twilio sms failed: status=%d", res.StatusCode)
		}
	}
	return nil
}

type webhookNotifier struct {
	url    string
	client *http.Client
}

// NewWebhookNotifier creates a generic HTTP webhook notifier.
func NewWebhookNotifier(config map[string]any) (Notifier, error) {
	url := stringValue(config, "url", "")
	if url == "" {
		return nil, fmt.Errorf("webhook url required")
	}
	return &webhookNotifier{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}},
	}, nil
}

func (n *webhookNotifier) Type() string { return "webhook" }

func (n *webhookNotifier) Send(ctx context.Context, alert model.AlertRecord, incident *domain.Incident) error {
	payload := map[string]any{
		"alert":    alert,
		"incident": incident,
	}
	return postJSON(ctx, n.client, n.url, payload)
}

func renderHTML(alert model.AlertRecord, incident *domain.Incident) string {
	incidentLine := ""
	if incident != nil {
		incidentLine = fmt.Sprintf("<p><strong>Incident:</strong> %s</p>", incident.Title)
	}
	return fmt.Sprintf(`<html><body><h2>%s</h2><p>%s</p><p><strong>Service:</strong> %s</p><p><strong>Severity:</strong> %s</p>%s</body></html>`,
		alert.Title, alert.Description, alert.Service, alert.Severity, incidentLine)
}

func stringValue(config map[string]any, key, fallback string) string {
	if value, ok := config[key].(string); ok && value != "" {
		return value
	}
	return fallback
}
