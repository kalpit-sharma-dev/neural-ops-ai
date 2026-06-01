package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// Limiter provides per-tenant token bucket rate limiting.
type Limiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rates    map[string]float64
}

// New creates a limiter with plan-tier rates (events per second).
// A rate of 0 means unlimited for that plan.
func New(planRates map[string]float64) *Limiter {
	if planRates == nil {
		planRates = map[string]float64{}
	}
	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rates:    planRates,
	}
}

// Allow reports whether the tenant is allowed to proceed under their plan.
func (l *Limiter) Allow(tenantID, plan string) bool {
	if tenantID == "" {
		tenantID = "default"
	}
	rps := l.rateForPlan(plan)
	if rps <= 0 {
		return true
	}

	limiter := l.getLimiter(tenantID, rps)
	return limiter.Allow()
}

func (l *Limiter) rateForPlan(plan string) float64 {
	if plan == "" {
		plan = "startup"
	}
	if rate, ok := l.rates[plan]; ok {
		return rate
	}
	if rate, ok := l.rates["startup"]; ok {
		return rate
	}
	return 10000
}

func (l *Limiter) getLimiter(tenantID string, rps float64) *rate.Limiter {
	l.mu.RLock()
	limiter, ok := l.limiters[tenantID]
	l.mu.RUnlock()
	if ok {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, ok = l.limiters[tenantID]; ok {
		return limiter
	}

	burst := int(rps)
	if burst < 1 {
		burst = 1
	}
	if burst > 100000 {
		burst = 100000
	}

	limiter = rate.NewLimiter(rate.Limit(rps), burst)
	l.limiters[tenantID] = limiter
	return limiter
}
