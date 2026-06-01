package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/neuralops/platform/internal/search/config"
)

// EnsureIndexTemplate installs the logs index template with strict mappings.
func (c *Client) EnsureIndexTemplate(ctx context.Context, cfg config.ElasticsearchConfig) error {
	templateName := cfg.IndexPrefix + "-template"
	aliasName := cfg.IndexAlias
	if aliasName == "" {
		aliasName = cfg.IndexPrefix
	}
	body := map[string]any{
		"index_patterns": []string{cfg.IndexPrefix + "-*"},
		"template": map[string]any{
			"settings": map[string]any{
				"number_of_shards":               1,
				"number_of_replicas":             0,
				"index.codec":                    "best_compression",
				"index.lifecycle.name":           cfg.ILMPolicy,
				"index.lifecycle.rollover_alias": aliasName,
			},
			"mappings": map[string]any{
				"dynamic": false,
				"properties": map[string]any{
					"timestamp":      map[string]string{"type": "date"},
					"service":        map[string]string{"type": "keyword"},
					"environment":    map[string]string{"type": "keyword"},
					"severity":       map[string]string{"type": "keyword"},
					"message":        map[string]string{"type": "text"},
					"traceId":        map[string]string{"type": "keyword"},
					"txnId":          map[string]string{"type": "keyword"},
					"host":           map[string]string{"type": "keyword"},
					"pod":            map[string]string{"type": "keyword"},
					"classification": map[string]string{"type": "keyword"},
					"labels":         map[string]string{"type": "object"},
					"plain_english":  map[string]string{"type": "text"},
					"embedding_id":   map[string]string{"type": "keyword"},
				},
			},
		},
	}
	payload, _ := json.Marshal(body)
	res, err := c.es.Indices.PutIndexTemplate(templateName, bytes.NewReader(payload), c.es.Indices.PutIndexTemplate.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("put index template: %s", string(raw))
	}
	return nil
}

// EnsureILMPolicy configures index lifecycle management.
func (c *Client) EnsureILMPolicy(ctx context.Context, cfg config.ElasticsearchConfig) error {
	policy := map[string]any{
		"policy": map[string]any{
			"phases": map[string]any{
				"hot": map[string]any{
					"min_age": "0ms",
					"actions": map[string]any{
						"rollover": map[string]any{
							"max_size": cfg.RolloverMaxSize,
							"max_age":  cfg.RolloverMaxAge,
						},
					},
				},
				"warm": map[string]any{
					"min_age": fmt.Sprintf("%dd", cfg.WarmDays),
					"actions": map[string]any{
						"allocate": map[string]any{"number_of_replicas": 0},
					},
				},
				"cold": map[string]any{
					"min_age": fmt.Sprintf("%dd", cfg.ColdDays),
					"actions": map[string]any{},
				},
				"delete": map[string]any{
					"min_age": fmt.Sprintf("%dd", cfg.DeleteDays),
					"actions": map[string]any{
						"delete": map[string]any{},
					},
				},
			},
		},
	}
	payload, _ := json.Marshal(policy)
	res, err := c.es.ILM.PutLifecycle(cfg.ILMPolicy,
		c.es.ILM.PutLifecycle.WithBody(bytes.NewReader(payload)),
		c.es.ILM.PutLifecycle.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("put ilm policy: %s", string(raw))
	}
	return nil
}
