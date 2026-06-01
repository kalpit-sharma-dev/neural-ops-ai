package seed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateAllLogsCount(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	logs := GenerateAllLogs(now, DefaultLogCount)
	require.GreaterOrEqual(t, len(logs), DefaultLogCount)
	require.Contains(t, joinMessages(logs), "SocketTimeoutException")
}

func TestGenerateTransactionsDistribution(t *testing.T) {
	now := time.Now().UTC()
	txns := GenerateTransactions(now)
	require.Equal(t, DefaultTransactionCount+UPIFailedTxnCount, len(txns))

	success, failed, upiOutageFailed := 0, 0, 0
	for _, txn := range txns {
		switch txn.Status {
		case "SUCCESS":
			success++
		case "FAILED":
			failed++
		}
		if len(txn.TxnID) >= 8 && txn.TxnID[:8] == "UPI-FAIL" {
			upiOutageFailed++
		}
	}
	require.Equal(t, 400, success)
	require.Equal(t, UPIFailedTxnCount, upiOutageFailed)
	require.Equal(t, DefaultTransactionCount-400+UPIFailedTxnCount, failed)
}

func TestGenerateDeploymentsDistribution(t *testing.T) {
	deployments := GenerateDeployments(time.Now().UTC())
	require.Len(t, deployments, DefaultDeploymentCount)

	clean, minor, caused := 0, 0, 0
	for _, dep := range deployments {
		switch dep.Category {
		case "clean":
			clean++
		case "minor_issue":
			minor++
		case "caused_incident":
			caused++
		}
	}
	require.Equal(t, 60, clean)
	require.Equal(t, 30, minor)
	require.Equal(t, 10, caused)
}

func TestScenarioTimelineOrdering(t *testing.T) {
	timeline := NewScenarioTimeline(time.Now().UTC())
	require.True(t, timeline.ErrorsStart.After(timeline.DeployAt))
	require.True(t, timeline.ThreadPoolAt.After(timeline.ErrorsStart))
	require.True(t, timeline.CascadeAt.After(timeline.ThreadPoolAt))
	require.True(t, timeline.ResolvedAt.After(timeline.CascadeAt))
}

func joinMessages(logs []LogRecord) string {
	out := ""
	for _, log := range logs {
		out += log.Message
	}
	return out
}
