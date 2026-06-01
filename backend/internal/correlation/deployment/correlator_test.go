package deployment

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestCorrelationLevelFromRatio(t *testing.T) {
	t.Parallel()
	deployedAt := time.Now().UTC().Add(-5 * time.Minute)
	evalWindow := 15 * time.Minute

	cases := []struct {
		before float64
		after  float64
		want   repository.CorrelationLevel
	}{
		{0.01, 0.03, repository.CorrelationHigh},
		{0.01, 0.05, repository.CorrelationHigh},
		{0.10, 0.11, repository.CorrelationLow},
	}

	for _, tc := range cases {
		level := classifyCorrelation(tc.before, tc.after, deployedAt, evalWindow)
		require.Equal(t, tc.want, level)
	}
}

func TestHandleLogIncrementsWatchCounters(t *testing.T) {
	c := &Correlator{
		watchWindow: time.Hour,
		evalWindow:  15 * time.Minute,
		watches: []watch{{
			deployment: domain.Deployment{
				ID:         uuid.New(),
				Service:    "upi-service",
				DeployedAt: time.Now().UTC(),
			},
			expiresAt: time.Now().UTC().Add(time.Hour),
		}},
	}

	c.HandleLog(domain.LogEntry{Service: "upi-service", Severity: domain.LogSeverityError})
	c.HandleLog(domain.LogEntry{Service: "upi-service", Severity: domain.LogSeverityInfo})

	require.Equal(t, int64(2), c.watches[0].totalCount)
	require.Equal(t, int64(1), c.watches[0].errorCount)
}

func classifyCorrelation(before, after float64, deployedAt time.Time, evalWindow time.Duration) repository.CorrelationLevel {
	if before <= 0 {
		before = 0.01
	}
	ratio := after / before
	switch {
	case ratio > 2.0 && time.Since(deployedAt) <= evalWindow:
		return repository.CorrelationHigh
	case ratio >= 1.5:
		return repository.CorrelationMedium
	default:
		return repository.CorrelationLow
	}
}
