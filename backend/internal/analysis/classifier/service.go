package classifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
)

// ClassificationCache stores cached classifications.
type ClassificationCache interface {
	GetClassification(ctx context.Context, key string) (*domain.ErrorClassification, error)
	SetClassification(ctx context.Context, key string, value *domain.ErrorClassification, ttlSeconds int) error
}

// Service classifies log messages using rules, cache, and optional LLM fallback.
type Service struct {
	llm        ai.LLMClient
	cache      ClassificationCache
	threshold  float64
	llmBackend string
	onLLMCall  func(method string)
}

// NewService creates a classifier service.
func NewService(llm ai.LLMClient, cache ClassificationCache, threshold float64, llmBackend string, onLLMCall func(method string)) *Service {
	if threshold <= 0 {
		threshold = 0.8
	}
	return &Service{
		llm:        llm,
		cache:      cache,
		threshold:  threshold,
		llmBackend: llmBackend,
		onLLMCall:  onLLMCall,
	}
}

// Classify classifies a log message.
func (s *Service) Classify(ctx context.Context, message string) (*domain.ErrorClassification, error) {
	cacheKey := hashMessage(message)
	if s.cache != nil {
		if cached, err := s.cache.GetClassification(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	if rule := ClassifyRuleBased(message); rule != nil && rule.Confidence >= s.threshold {
		s.storeCache(ctx, cacheKey, rule)
		return rule, nil
	}

	if s.llm == nil {
		if rule := ClassifyRuleBased(message); rule != nil {
			s.storeCache(ctx, cacheKey, rule)
			return rule, nil
		}
		return &domain.ErrorClassification{
			Category:   domain.ErrorCategoryUnknown,
			Confidence: 0.4,
			Reasoning:  "No rule match and LLM unavailable",
		}, nil
	}

	if s.onLLMCall != nil {
		s.onLLMCall("Classify")
	}

	llmResult, err := s.llm.Classify(ctx, message)
	if err != nil {
		if rule := ClassifyRuleBased(message); rule != nil {
			s.storeCache(ctx, cacheKey, rule)
			return rule, nil
		}
		return nil, fmt.Errorf("llm classify: %w", err)
	}

	if llmResult.Confidence < s.threshold && llmResult.Category == domain.ErrorCategoryUnknown {
		if rule := ClassifyRuleBased(message); rule != nil {
			s.storeCache(ctx, cacheKey, rule)
			return rule, nil
		}
	}

	s.storeCache(ctx, cacheKey, llmResult)
	return llmResult, nil
}

func (s *Service) storeCache(ctx context.Context, key string, value *domain.ErrorClassification) {
	if s.cache == nil || value == nil {
		return
	}
	_ = s.cache.SetClassification(ctx, key, value, 3600)
}

func hashMessage(message string) string {
	sum := sha256.Sum256([]byte(message))
	return hex.EncodeToString(sum[:])
}
