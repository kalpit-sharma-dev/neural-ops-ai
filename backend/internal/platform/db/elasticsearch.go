package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

// ApplyElasticsearchLogTemplate installs the canonical logs index template from migrations.
func ApplyElasticsearchLogTemplate(ctx context.Context, es *elasticsearch.Client, templatePath string) error {
	if templatePath == "" {
		templatePath = resolveMigrationAsset([]string{
			"migrations/elasticsearch/logs-template.json",
			"../migrations/elasticsearch/logs-template.json",
			"../../migrations/elasticsearch/logs-template.json",
		})
	}
	if templatePath == "" {
		return fmt.Errorf("elasticsearch logs template not found")
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	var payload struct {
		IndexPatterns []string       `json:"index_patterns"`
		Template      map[string]any `json:"template"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	body := map[string]any{
		"index_patterns": payload.IndexPatterns,
		"template":       payload.Template,
	}
	encoded, _ := json.Marshal(body)
	res, err := es.Indices.PutIndexTemplate("logs-template", bytes.NewReader(encoded), es.Indices.PutIndexTemplate.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		response, _ := io.ReadAll(res.Body)
		return fmt.Errorf("put logs template: %s", string(response))
	}
	return nil
}

func resolveMigrationAsset(candidates []string) string {
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
