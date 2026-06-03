package commitments

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

const commitmentAlertPolicyID = "finops-commitment-alert"

// AlertThresholds for commitment monitoring (REQ-FINOPS-042).
const (
	ExpiringDaysThreshold  = 90
	UnderUtilizedThreshold = 60.0
)

// DetectAlerts returns alerts for expiring or under-utilized commitments.
func (s *Service) DetectAlerts(commitments []domain.Commitment) []domain.CommitmentAlert {
	now := time.Now().UTC()
	var out []domain.CommitmentAlert
	for _, c := range commitments {
		daysToExpiry := c.ExpiresAt.Sub(now).Hours() / 24
		if daysToExpiry <= ExpiringDaysThreshold && daysToExpiry > 0 {
			severity := "medium"
			if daysToExpiry <= 30 {
				severity = "high"
			}
			out = append(out, domain.CommitmentAlert{
				ID: uuid.NewString(), CommitmentID: c.ID, Provider: c.Provider,
				AlertType: "expiring", Severity: severity,
				Message: fmt.Sprintf("%s %s in %s expires in %.0f days (%.0f%% utilized)",
					c.Provider, c.CommitmentType, c.Region, daysToExpiry, c.UtilizationPct),
				UtilizationPct: c.UtilizationPct, ExpiresAt: c.ExpiresAt,
				AlertPolicyID: commitmentAlertPolicyID,
			})
		}
		if c.UtilizationPct < UnderUtilizedThreshold && c.Status == "active" {
			out = append(out, domain.CommitmentAlert{
				ID: uuid.NewString(), CommitmentID: c.ID, Provider: c.Provider,
				AlertType: "under_utilized", Severity: "medium",
				Message: fmt.Sprintf("%s %s utilization %.0f%% below %.0f%% threshold — review rightsizing or exchange",
					c.Provider, c.CommitmentType, c.UtilizationPct, UnderUtilizedThreshold),
				UtilizationPct: c.UtilizationPct, ExpiresAt: c.ExpiresAt,
				AlertPolicyID: commitmentAlertPolicyID,
			})
		}
	}
	return out
}
