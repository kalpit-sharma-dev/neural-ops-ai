package temporal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

type windowState struct {
	start    time.Time
	end      time.Time
	services map[string]int
	infra    []string
}

// Analyzer performs sliding-window temporal correlation.
type Analyzer struct {
	log    *zap.Logger
	repo   *repository.Store
	window time.Duration
	mu     sync.Mutex
	current windowState
}

// NewAnalyzer creates a temporal analyzer.
func NewAnalyzer(log *zap.Logger, repo *repository.Store, window time.Duration) *Analyzer {
	now := time.Now().UTC()
	return &Analyzer{
		log:    log,
		repo:   repo,
		window: window,
		current: windowState{
			start:    now,
			end:      now.Add(window),
			services: make(map[string]int),
		},
	}
}

// HandleLog records service failures in the current window.
func (a *Analyzer) HandleLog(entry domain.LogEntry) {
	if !entry.IsErrorSeverity() {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rotateIfNeeded(time.Now().UTC())
	a.current.services[entry.Service]++
}

// HandleInfraEvent records infrastructure events in the current window.
func (a *Analyzer) HandleInfraEvent(event string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rotateIfNeeded(time.Now().UTC())
	a.current.infra = append(a.current.infra, event)
}

func (a *Analyzer) rotateIfNeeded(now time.Time) {
	if now.Before(a.current.end) {
		return
	}
	a.flushLocked(context.Background())
	a.current = windowState{
		start:    now,
		end:      now.Add(a.window),
		services: make(map[string]int),
	}
}

func (a *Analyzer) flushLocked(ctx context.Context) {
	if len(a.current.services) < 2 {
		return
	}
	services := make([]string, 0, len(a.current.services))
	for service := range a.current.services {
		services = append(services, service)
	}
	hint := fmt.Sprintf("%d services failing in the same %s window", len(services), a.window)
	if err := a.repo.SaveTemporalCorrelation(ctx, a.current.start, a.current.end, services, hint, a.current.infra); err != nil {
		a.log.Warn("save temporal correlation failed", zap.Error(err))
	} else {
		a.log.Info("temporal correlation detected", zap.Strings("services", services))
	}
}

// StartFlusher rotates windows periodically.
func (a *Analyzer) StartFlusher(ctx context.Context) {
	ticker := time.NewTicker(a.window)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				a.mu.Lock()
				a.flushLocked(context.Background())
				a.mu.Unlock()
				return
			case <-ticker.C:
				a.mu.Lock()
				a.rotateIfNeeded(time.Now().UTC())
				a.mu.Unlock()
			}
		}
	}()
}
