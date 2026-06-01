package classifier

import (
	"regexp"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

type rule struct {
	category   domain.ErrorCategory
	patterns   []*regexp.Regexp
	keywords   []string
	confidence float64
	reasoning  string
}

var defaultRules = []rule{
	{
		category:   domain.ErrorCategoryTimeout,
		keywords:   []string{"sockettimeoutexception", "read timed out", "context deadline exceeded", "timeout"},
		confidence: 0.95,
		reasoning:  "Matched timeout-related exception or keyword",
	},
	{
		category:   domain.ErrorCategoryDBDeadlock,
		keywords:   []string{"deadlock", "lock wait timeout", "serialization failure"},
		confidence: 0.93,
		reasoning:  "Matched database deadlock indicators",
	},
	{
		category:   domain.ErrorCategoryMemoryLeak,
		keywords:   []string{"oomkilled", "outofmemoryerror", "cannot allocate memory", "heap space"},
		confidence: 0.94,
		reasoning:  "Matched memory exhaustion indicators",
	},
	{
		category:   domain.ErrorCategoryConnectionExhaustion,
		keywords:   []string{"connection refused", "connection reset", "too many connections", "pool exhausted"},
		confidence: 0.92,
		reasoning:  "Matched connection exhaustion indicators",
	},
	{
		category:   domain.ErrorCategoryDNSIssue,
		keywords:   []string{"no such host", "dns", "name resolution"},
		confidence: 0.9,
		reasoning:  "Matched DNS resolution failure indicators",
	},
	{
		category:   domain.ErrorCategorySSLIssue,
		keywords:   []string{"ssl", "tls", "certificate", "x509"},
		confidence: 0.9,
		reasoning:  "Matched TLS/SSL failure indicators",
	},
	{
		category:   domain.ErrorCategoryRetryStorm,
		keywords:   []string{"retry storm", "retry limit exceeded", "max retries exceeded"},
		confidence: 0.88,
		reasoning:  "Matched retry storm indicators",
	},
	{
		category:   domain.ErrorCategoryKafkaLag,
		keywords:   []string{"kafka lag", "consumer lag", "offset lag"},
		confidence: 0.88,
		reasoning:  "Matched Kafka lag indicators",
	},
	{
		category:   domain.ErrorCategoryThreadStarvation,
		keywords:   []string{"thread pool exhausted", "thread starvation", "rejected execution"},
		confidence: 0.91,
		reasoning:  "Matched thread pool exhaustion indicators",
	},
	{
		category:   domain.ErrorCategoryCircuitBreakerOpen,
		keywords:   []string{"circuit breaker open", "breaker is open", "circuitbreakeropenexception"},
		confidence: 0.9,
		reasoning:  "Matched circuit breaker open indicators",
	},
	{
		category:   domain.ErrorCategoryRateLimiting,
		keywords:   []string{"rate limit", "429", "too many requests"},
		confidence: 0.89,
		reasoning:  "Matched rate limiting indicators",
	},
	{
		category:   domain.ErrorCategoryAuthFailure,
		keywords:   []string{"unauthorized", "authentication failed", "invalid token", "403", "401"},
		confidence: 0.87,
		reasoning:  "Matched authentication/authorization failure indicators",
	},
	{
		category:   domain.ErrorCategoryDependencyFailure,
		keywords:   []string{"dependency failed", "downstream service unavailable", "upstream error"},
		confidence: 0.86,
		reasoning:  "Matched dependency failure indicators",
	},
}

// ClassifyRuleBased performs high-confidence rule-based classification.
func ClassifyRuleBased(message string) *domain.ErrorClassification {
	lower := strings.ToLower(message)
	for _, item := range defaultRules {
		for _, keyword := range item.keywords {
			if strings.Contains(lower, keyword) {
				return &domain.ErrorClassification{
					Category:   item.category,
					Confidence: item.confidence,
					Reasoning:  item.reasoning,
				}
			}
		}
		for _, pattern := range item.patterns {
			if pattern.MatchString(message) {
				return &domain.ErrorClassification{
					Category:   item.category,
					Confidence: item.confidence,
					Reasoning:  item.reasoning,
				}
			}
		}
	}
	return nil
}
