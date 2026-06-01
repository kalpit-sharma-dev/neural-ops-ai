package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

// BuildFingerprint creates a deduplication fingerprint.
func BuildFingerprint(rootService string, category domain.ErrorCategory, at time.Time, window time.Duration) string {
	bucket := at.UTC().Truncate(window).Unix()
	raw := fmt.Sprintf("%s|%s|%d", rootService, category, bucket)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

// MergeIncidents appends data from a new signal into an existing incident.
func MergeIncidents(existing *domain.Incident, service string, logID domain.LogEntry) {
	for _, existingService := range existing.AffectedServices {
		if existingService == service {
			goto appendLog
		}
	}
	existing.AffectedServices = append(existing.AffectedServices, service)

appendLog:
	existing.CorrelatedLogIDs = append(existing.CorrelatedLogIDs, logID.ID)
	existing.UpdatedAt = time.Now().UTC()
}
