package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// QdrantCollectionSpec describes a Qdrant collection migration asset.
type QdrantCollectionSpec struct {
	Collection     string         `json:"collection"`
	Vectors        map[string]any `json:"vectors"`
	PayloadSchema  map[string]any `json:"payload_schema"`
}

// EnsureQdrantCollectionFromMigration creates or updates a Qdrant collection from JSON spec.
func EnsureQdrantCollectionFromMigration(ctx context.Context, baseURL, specPath string) error {
	if specPath == "" {
		specPath = resolveMigrationAsset([]string{
			"migrations/qdrant/log_embeddings.json",
			"../migrations/qdrant/log_embeddings.json",
			"../../migrations/qdrant/log_embeddings.json",
		})
	}
	if specPath == "" {
		return fmt.Errorf("qdrant collection spec not found")
	}

	raw, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	var spec QdrantCollectionSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return err
	}
	if spec.Collection == "" {
		spec.Collection = "log_embeddings"
	}
	if spec.Vectors == nil {
		spec.Vectors = map[string]any{"size": 1536, "distance": "Cosine"}
	}

	body, _ := json.Marshal(map[string]any{"vectors": spec.Vectors})
	url := fmt.Sprintf("%s/collections/%s", trimRightSlash(baseURL), spec.Collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 && res.StatusCode != 409 {
		response, _ := io.ReadAll(res.Body)
		return fmt.Errorf("ensure qdrant collection: status=%d body=%s", res.StatusCode, string(response))
	}
	return nil
}

func trimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
