package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// QdrantStore writes embeddings to Qdrant.
type QdrantStore struct {
	baseURL    string
	collection string
	vectorSize int
	client     *http.Client
}

// NewQdrantStore creates a Qdrant repository.
func NewQdrantStore(url, collection string, vectorSize int) *QdrantStore {
	return &QdrantStore{
		baseURL:    trimRightSlash(url),
		collection: collection,
		vectorSize: vectorSize,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// EnsureCollection creates the embeddings collection if missing.
func (s *QdrantStore) EnsureCollection(ctx context.Context) error {
	body, _ := json.Marshal(map[string]any{
		"vectors": map[string]any{
			"size":     s.vectorSize,
			"distance": "Cosine",
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/collections/%s", s.baseURL, s.collection), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 && res.StatusCode != 409 {
		return fmt.Errorf("create qdrant collection: status=%d", res.StatusCode)
	}
	return nil
}

// UpsertEmbeddings stores vectors in Qdrant and returns point IDs.
func (s *QdrantStore) UpsertEmbeddings(ctx context.Context, logIDs []uuid.UUID, vectors [][]float32, payloads []map[string]any) ([]string, error) {
	if len(logIDs) == 0 {
		return nil, nil
	}

	points := make([]map[string]any, 0, len(logIDs))
	ids := make([]string, 0, len(logIDs))
	for i, logID := range logIDs {
		pointID := logID.String()
		ids = append(ids, pointID)
		point := map[string]any{
			"id":      pointID,
			"vector":  vectors[i],
			"payload": payloads[i],
		}
		points = append(points, point)
	}

	body, _ := json.Marshal(map[string]any{"points": points})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/collections/%s/points", s.baseURL, s.collection), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("upsert qdrant points: status=%d", res.StatusCode)
	}
	return ids, nil
}

func trimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
