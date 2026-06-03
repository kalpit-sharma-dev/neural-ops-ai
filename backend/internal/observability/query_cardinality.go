package observability

import (
	"os"
	"strconv"
	"strings"
)

const (
	defaultMaxQueryHits       = 500
	defaultMaxCardinality     = 10000
	defaultMaxServiceCard     = 256
)

// CardinalityGuard reports query safety limits applied during NexQL execution.
type CardinalityGuard struct {
	MaxHits           int  `json:"maxHits"`
	MaxCardinality    int  `json:"maxCardinality"`
	EstimatedCard     int  `json:"estimatedCardinality"`
	Truncated         bool `json:"truncated"`
	CardinalityCapped bool `json:"cardinalityCapped"`
}

func queryLimits() (maxHits, maxCard int) {
	maxHits = envInt("NEXQL_MAX_HITS", defaultMaxQueryHits)
	maxCard = envInt("NEXQL_MAX_CARDINALITY", defaultMaxCardinality)
	return maxHits, maxCard
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func applyCardinalityGuard(hits []UnifiedQueryHit, req UnifiedQueryRequest) ([]UnifiedQueryHit, CardinalityGuard) {
	maxHits, maxCard := queryLimits()
	services := make(map[string]struct{})
	for _, h := range hits {
		if h.Service != "" {
			services[strings.ToLower(h.Service)] = struct{}{}
		}
	}
	guard := CardinalityGuard{
		MaxHits:        maxHits,
		MaxCardinality: maxCard,
		EstimatedCard:  len(services) * max(1, len(hits)/max(1, len(services))),
	}
	if len(services) > defaultMaxServiceCard {
		guard.CardinalityCapped = true
		filtered := make([]UnifiedQueryHit, 0, len(hits))
		allowed := 0
		for _, h := range hits {
			if allowed >= maxCard {
				break
			}
			filtered = append(filtered, h)
			allowed++
		}
		hits = filtered
		guard.Truncated = true
	}
	if req.Limit > 0 && req.Limit < maxHits {
		maxHits = req.Limit
	}
	if len(hits) > maxHits {
		hits = hits[:maxHits]
		guard.Truncated = true
	}
	guard.EstimatedCard = len(services)
	return hits, guard
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
