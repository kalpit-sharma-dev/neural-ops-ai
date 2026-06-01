package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/neuralops/platform/internal/search/config"
)

// EnsureRolloverAlias creates the write alias and bootstrap index when missing.
func (c *Client) EnsureRolloverAlias(ctx context.Context, cfg config.ElasticsearchConfig) error {
	aliasName := cfg.IndexAlias
	if aliasName == "" {
		aliasName = cfg.IndexPrefix
	}
	if aliasName == "" {
		return fmt.Errorf("index alias is required for rollover")
	}

	initialIndex := cfg.InitialIndex
	if initialIndex == "" {
		initialIndex = cfg.IndexPrefix + "-000001"
	}

	res, err := c.es.Indices.GetAlias(c.es.Indices.GetAlias.WithName(aliasName), c.es.Indices.GetAlias.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	createBody := map[string]any{
		"aliases": map[string]any{
			aliasName: map[string]any{
				"is_write_index": true,
			},
		},
	}
	payload, _ := json.Marshal(createBody)
	createRes, err := c.es.Indices.Create(initialIndex,
		c.es.Indices.Create.WithBody(bytes.NewReader(payload)),
		c.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		raw, _ := io.ReadAll(createRes.Body)
		body := string(raw)
		if strings.Contains(body, "resource_already_exists") {
			return c.attachWriteAlias(ctx, initialIndex, aliasName)
		}
		return fmt.Errorf("create bootstrap index: %s", body)
	}
	return nil
}

func (c *Client) attachWriteAlias(ctx context.Context, indexName, aliasName string) error {
	body := map[string]any{
		"actions": []map[string]any{
			{
				"add": map[string]any{
					"index":          indexName,
					"alias":          aliasName,
					"is_write_index": true,
				},
			},
		},
	}
	payload, _ := json.Marshal(body)
	res, err := c.es.Indices.UpdateAliases(bytes.NewReader(payload), c.es.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("attach write alias: %s", string(raw))
	}
	return nil
}
