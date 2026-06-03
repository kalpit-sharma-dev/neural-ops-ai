package reporting

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Service manages FinOps report generation and schedules (REQ-FINOPS-070).
type Service struct {
	schedules map[string][]domain.ReportSchedule
}

// NewService creates reporting service.
func NewService() *Service {
	s := &Service{schedules: map[string][]domain.ReportSchedule{}}
	s.seed("default")
	return s
}

func (s *Service) seed(tenantID string) {
	now := time.Now().UTC()
	s.schedules[tenantID] = []domain.ReportSchedule{
		{ID: "rpt-1", Name: "Weekly cost summary", Scope: "all", Format: "csv", Cadence: "weekly", DeliveryChannel: "email", DeliveryTarget: "finops@neuralops.ai", Enabled: true, CreatedAt: now},
		{ID: "rpt-2", Name: "Monthly executive FinOps", Scope: "all", Format: "pdf", Cadence: "monthly", DeliveryChannel: "webhook", DeliveryTarget: "https://hooks.example.com/finops", Enabled: true, CreatedAt: now},
	}
}

// ListReports returns recent generated reports.
func (s *Service) ListReports(scope string) []domain.FinOpsReport {
	now := time.Now().UTC()
	return []domain.FinOpsReport{
		{ID: "gen-1", Name: "Cost breakdown export", Scope: scope, Format: "csv", URL: "/api/v1/finops/reports/gen-1/download", Generated: now.Add(-24 * time.Hour)},
		{ID: "gen-2", Name: "Optimization savings summary", Scope: scope, Format: "pdf", URL: "/api/v1/finops/reports/gen-2/download", Generated: now.Add(-72 * time.Hour)},
	}
}

// ListSchedules returns report schedules for tenant.
func (s *Service) ListSchedules(tenantID string) []domain.ReportSchedule {
	if list, ok := s.schedules[tenantID]; ok {
		return append([]domain.ReportSchedule(nil), list...)
	}
	return nil
}

// CreateSchedule adds a report schedule.
func (s *Service) CreateSchedule(tenantID string, sch domain.ReportSchedule) domain.ReportSchedule {
	if sch.ID == "" {
		sch.ID = "rpt-" + uuid.NewString()[:8]
	}
	sch.CreatedAt = time.Now().UTC()
	if sch.Format == "" {
		sch.Format = "csv"
	}
	if sch.Cadence == "" {
		sch.Cadence = "weekly"
	}
	s.schedules[tenantID] = append(s.schedules[tenantID], sch)
	return sch
}

// GenerateReportURL returns a download URL for a report id.
func (s *Service) GenerateReportURL(id, format string) string {
	return fmt.Sprintf("/api/v1/finops/reports/%s/download?format=%s", id, format)
}
