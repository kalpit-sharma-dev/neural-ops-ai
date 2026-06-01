package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/neuralops/platform/internal/ai"
	correlationrepo "github.com/neuralops/platform/internal/correlation/repository"
	searchai "github.com/neuralops/platform/internal/search/ai"
	"github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/search/dto"
	esclient "github.com/neuralops/platform/internal/search/elasticsearch"
	qdrantclient "github.com/neuralops/platform/internal/search/qdrant"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

// Service orchestrates search operations across Elasticsearch, Qdrant, and ClickHouse.
type Service struct {
	cfg        *config.Config
	log        *zap.Logger
	es         *esclient.Client
	qdrant     *qdrantclient.Client
	txnStore   *correlationrepo.ClickHouseStore
	llm        ai.LLMClient
	parser     *searchai.Parser
}

// New creates a search service.
func New(
	cfg *config.Config,
	log *zap.Logger,
	es *esclient.Client,
	qdrant *qdrantclient.Client,
	txnStore *correlationrepo.ClickHouseStore,
	llm ai.LLMClient,
) *Service {
	return &Service{
		cfg:      cfg,
		log:      log,
		es:       es,
		qdrant:   qdrant,
		txnStore: txnStore,
		llm:      llm,
		parser:   searchai.NewParser(llm),
	}
}

// SearchLogs executes structured full-text search.
func (s *Service) SearchLogs(ctx context.Context, req dto.LogSearchRequest) (*dto.SearchResponse, error) {
	req = s.normalizeLogRequest(req)
	if s.es == nil {
		return nil, fmt.Errorf("elasticsearch unavailable")
	}
	return s.es.SearchLogs(ctx, req)
}

// SemanticSearch performs vector search with optional hybrid BM25 fusion.
func (s *Service) SemanticSearch(ctx context.Context, req dto.SemanticSearchRequest) (*dto.SearchResponse, error) {
	req = s.normalizeSemanticRequest(req)
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	if s.qdrant == nil {
		return nil, fmt.Errorf("qdrant unavailable")
	}

	vectors, err := s.llm.Embed(ctx, []string{req.Query})
	if err != nil || len(vectors) == 0 {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	vectorHits, err := s.qdrant.Search(ctx, vectors[0], req.Size, req.StartTime, req.EndTime, req.TenantID)
	if err != nil {
		return nil, err
	}

	if !req.Hybrid || s.es == nil {
		return &dto.SearchResponse{
			Hits:  vectorHits,
			Total: int64(len(vectorHits)),
		}, nil
	}

	bm25, err := s.es.SearchLogs(ctx, dto.LogSearchRequest{
		Query:     req.Query,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		TenantID:  req.TenantID,
		Size:      req.Size,
	})
	if err != nil {
		s.log.Warn("hybrid bm25 search failed, returning vector results only", zap.Error(err))
		return &dto.SearchResponse{Hits: vectorHits, Total: int64(len(vectorHits))}, nil
	}

	merged := mergeHybrid(vectorHits, bm25.Hits, s.cfg.Search.HybridVectorWeight, s.cfg.Search.HybridBM25Weight)
	if len(merged) > req.Size {
		merged = merged[:req.Size]
	}
	return &dto.SearchResponse{
		Hits:  merged,
		Total: int64(len(merged)),
	}, nil
}

// SearchTrace returns all logs for a distributed trace.
func (s *Service) SearchTrace(ctx context.Context, traceID, tenantID string, size int) (*dto.SearchResponse, error) {
	if strings.TrimSpace(traceID) == "" {
		return nil, fmt.Errorf("traceId is required")
	}
	if s.es == nil {
		return nil, fmt.Errorf("elasticsearch unavailable")
	}
	return s.es.SearchByTrace(ctx, traceID, tenantID, size)
}

// GetTransaction returns a transaction journey by ID.
func (s *Service) GetTransaction(ctx context.Context, txnID string) (*domain.Transaction, error) {
	if strings.TrimSpace(txnID) == "" {
		return nil, fmt.Errorf("txnId is required")
	}
	if s.txnStore == nil {
		return nil, fmt.Errorf("clickhouse unavailable")
	}
	return s.txnStore.GetTransaction(ctx, txnID)
}

// SearchTransactions searches transaction journeys in ClickHouse.
func (s *Service) SearchTransactions(ctx context.Context, req dto.TransactionSearchRequest) ([]domain.Transaction, error) {
	if s.txnStore == nil {
		return nil, fmt.Errorf("clickhouse unavailable")
	}
	limit := req.Limit
	if limit <= 0 {
		limit = s.cfg.Search.DefaultPageSize
	}
	return s.txnStore.SearchTransactions(ctx, correlationrepo.TransactionFilter{
		TxnID:         req.TxnID,
		TxnType:       req.TxnType,
		Status:        req.Status,
		FailedService: req.FailedService,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Limit:         limit,
	})
}

// AISearch parses natural language and executes log search.
func (s *Service) AISearch(ctx context.Context, req dto.AISearchRequest, tenantID string, allowedServices []string) (*dto.ParsedQuery, *dto.SearchResponse, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, nil, fmt.Errorf("question is required")
	}

	size := req.Size
	if size <= 0 {
		size = s.cfg.Search.DefaultPageSize
	}

	parsed, err := s.parser.Parse(ctx, question)
	if err != nil {
		return nil, nil, err
	}

	searchReq := searchai.ToLogSearchRequest(parsed, size)
	searchReq.TenantID = tenantID
	searchReq.AllowedServices = allowedServices
	if len(allowedServices) > 0 && searchReq.Service != "" {
		allowed := false
		for _, svc := range allowedServices {
			if strings.EqualFold(strings.TrimSpace(svc), strings.TrimSpace(searchReq.Service)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return parsed, nil, fmt.Errorf("service not in developer scope")
		}
	}
	response, err := s.SearchLogs(ctx, searchReq)
	if err != nil {
		return parsed, nil, err
	}
	return parsed, response, nil
}

func (s *Service) normalizeLogRequest(req dto.LogSearchRequest) dto.LogSearchRequest {
	if req.Size <= 0 {
		req.Size = s.cfg.Search.DefaultPageSize
	}
	if req.Size > s.cfg.Search.MaxPageSize {
		req.Size = s.cfg.Search.MaxPageSize
	}
	return req
}

func (s *Service) normalizeSemanticRequest(req dto.SemanticSearchRequest) dto.SemanticSearchRequest {
	if req.Size <= 0 {
		req.Size = s.cfg.Search.DefaultPageSize
	}
	if req.Size > s.cfg.Search.MaxPageSize {
		req.Size = s.cfg.Search.MaxPageSize
	}
	return req
}

func mergeHybrid(vectorHits, bm25Hits []dto.LogHit, vectorWeight, bm25Weight float64) []dto.LogHit {
	if vectorWeight <= 0 && bm25Weight <= 0 {
		vectorWeight, bm25Weight = 0.5, 0.5
	}

	maxVector := 0.0
	maxBM25 := 0.0
	for _, hit := range vectorHits {
		if hit.Score > maxVector {
			maxVector = hit.Score
		}
	}
	for _, hit := range bm25Hits {
		if hit.Score > maxBM25 {
			maxBM25 = hit.Score
		}
	}

	combined := make(map[string]dto.LogHit)
	scores := make(map[string]float64)

	for _, hit := range vectorHits {
		normalized := 0.0
		if maxVector > 0 {
			normalized = hit.Score / maxVector
		}
		scores[hit.ID] += normalized * vectorWeight
		hit.Source = "hybrid"
		combined[hit.ID] = hit
	}
	for _, hit := range bm25Hits {
		normalized := 0.0
		if maxBM25 > 0 {
			normalized = hit.Score / maxBM25
		}
		scores[hit.ID] += normalized * bm25Weight
		if existing, ok := combined[hit.ID]; ok {
			existing.Message = firstNonEmpty(existing.Message, hit.Message)
			existing.Service = firstNonEmpty(existing.Service, hit.Service)
			combined[hit.ID] = existing
		} else {
			hit.Source = "hybrid"
			combined[hit.ID] = hit
		}
	}

	results := make([]dto.LogHit, 0, len(combined))
	for id, hit := range combined {
		hit.Score = scores[id]
		results = append(results, hit)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
