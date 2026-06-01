package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/incident/config"
	"github.com/neuralops/platform/internal/incident/repository"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service orchestrates incident creation and lifecycle.
type Service struct {
	log     *zap.Logger
	cfg     *config.Config
	repo    *repository.Store
	llm     ai.LLMClient
	redis   *redis.Client
	mu      sync.Mutex
	metrics map[string]*ServiceMetrics
}

// NewService creates an incident engine service.
func NewService(log *zap.Logger, cfg *config.Config, repo *repository.Store, llm ai.LLMClient, redisClient *redis.Client) *Service {
	return &Service{
		log:     log,
		cfg:     cfg,
		repo:    repo,
		llm:     llm,
		redis:   redisClient,
		metrics: make(map[string]*ServiceMetrics),
	}
}

// HandleLog processes enriched logs for incident detection.
func (s *Service) HandleLog(ctx context.Context, raw []byte) error {
	entry, err := decodeLogEntry(raw)
	if err != nil {
		return err
	}

	s.mu.Lock()
	metrics := s.metrics[entry.Service]
	if metrics == nil {
		metrics = &ServiceMetrics{}
		s.metrics[entry.Service] = metrics
	}
	metrics.TotalLogs++
	if entry.IsErrorSeverity() {
		metrics.ErrorLogs++
	}
	s.mu.Unlock()

	if entry.IsErrorSeverity() {
		return s.evaluateAndCreate(ctx, entry)
	}
	return nil
}

// HandleAnomaly processes anomaly events.
func (s *Service) HandleAnomaly(ctx context.Context, raw []byte) error {
	var anomaly domain.AnomalyDetection
	if err := json.Unmarshal(raw, &anomaly); err != nil {
		return err
	}

	s.mu.Lock()
	metrics := s.metrics[anomaly.ServiceName]
	if metrics == nil {
		metrics = &ServiceMetrics{}
		s.metrics[anomaly.ServiceName] = metrics
	}
	metrics.AnomalyScore = anomaly.Score
	s.mu.Unlock()

	if !anomaly.Detected {
		return nil
	}

	entry := domain.LogEntry{
		ID:        uuid.New(),
		Timestamp: anomaly.DetectedAt,
		Service:   anomaly.ServiceName,
		Severity:  domain.LogSeverityWarn,
		Message:   fmt.Sprintf("Anomaly detected on %s/%s score=%.2f", anomaly.ServiceName, anomaly.MetricType, anomaly.Score),
	}
	return s.evaluateAndCreate(ctx, entry)
}

func (s *Service) evaluateAndCreate(ctx context.Context, entry domain.LogEntry) error {
	s.refreshMetricsFromRedis(ctx, entry.Service)

	s.mu.Lock()
	metrics := s.metrics[entry.Service]
	if metrics == nil {
		metrics = &ServiceMetrics{}
	}
	s.mu.Unlock()

	severity := EvaluateSeverity(entry.Service, *metrics, s.cfg.Incident.PaymentServices)
	if !ShouldCreateIncident(severity) {
		return nil
	}

	category := domain.ErrorCategoryUnknown
	if entry.ClassifiedError != nil {
		category = entry.ClassifiedError.Category
	}

	tenantID := s.cfg.Incident.DefaultTenant
	if suppressed, _ := s.repo.IsSuppressed(ctx, tenantID, entry.Service, category); suppressed {
		return nil
	}

	fingerprint := BuildFingerprint(entry.Service, category, entry.Timestamp, s.cfg.Incident.DedupWindow)
	if existing, err := s.repo.FindActiveByFingerprint(ctx, fingerprint); err == nil && existing != nil {
		MergeIncidents(existing, entry.Service, entry)
		AppendTimelineEvent(existing, BuildTimelineEvent(domain.IncidentEventTypeError, "Additional correlated error", entry.Message, entry.Timestamp))
		return s.repo.UpdateIncident(ctx, existing)
	}

	incident := s.buildIncident(entry, severity, category)
	s.enrichWithAI(ctx, &incident, []domain.LogEntry{entry})
	incident.Timeline = []domain.IncidentEvent{
		BuildTimelineEvent(domain.IncidentEventTypeError, "Initial error detected", entry.Message, entry.Timestamp),
	}

	if err := s.repo.CreateIncident(ctx, &incident, fingerprint, entry.Service, string(category), tenantID); err != nil {
		return err
	}
	s.log.Info("incident created", zap.String("id", incident.ID.String()), zap.String("severity", string(severity)))
	return nil
}

func (s *Service) buildIncident(entry domain.LogEntry, severity domain.IncidentSeverity, category domain.ErrorCategory) domain.Incident {
	title := fmt.Sprintf("%s incident on %s", severity, entry.Service)
	if category != domain.ErrorCategoryUnknown {
		title = fmt.Sprintf("%s - %s on %s", severity, category, entry.Service)
	}
	return domain.Incident{
		ID:               uuid.New(),
		Title:            title,
		Summary:          entry.Message,
		Severity:         severity,
		Status:           domain.IncidentStatusOpen,
		AffectedServices: []string{entry.Service},
		CorrelatedLogIDs: []uuid.UUID{entry.ID},
		StartTime:        entry.Timestamp.UTC(),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func (s *Service) enrichWithAI(ctx context.Context, incident *domain.Incident, logs []domain.LogEntry) {
	if s.llm == nil {
		return
	}
	rca, err := s.llm.GenerateRCA(ctx, ai.IncidentContext{Incident: *incident, Logs: logs})
	if err != nil {
		s.log.Warn("llm incident summary failed", zap.Error(err))
		return
	}
	if rca == nil {
		return
	}
	incident.RootCauseAnalysis = rca
	if rca.RootCauseDescription != "" {
		incident.Summary = rca.RootCauseDescription
	}
	if rca.FirstFailingService != "" {
		incident.BlastRadius = []string{rca.FirstFailingService}
	}
}

func (s *Service) refreshMetricsFromRedis(ctx context.Context, service string) {
	if s.redis == nil {
		return
	}
	values, err := s.redis.HGetAll(ctx, "analysis:stats:"+service).Result()
	if err != nil {
		return
	}
	var total, errors float64
	_, _ = fmt.Sscan(values["total"], &total)
	_, _ = fmt.Sscan(values["errors"], &errors)

	s.mu.Lock()
	metrics := s.metrics[service]
	if metrics == nil {
		metrics = &ServiceMetrics{}
		s.metrics[service] = metrics
	}
	if total > 0 {
		metrics.TotalLogs = int64(total)
		metrics.ErrorLogs = int64(errors)
	}
	score, err := s.redis.Get(ctx, fmt.Sprintf("analysis:anomaly:%s:%s", service, "LATENCY")).Float64()
	if err == nil {
		metrics.AnomalyScore = score
	}
	metrics.PaymentDown = db.IsPaymentService(service) && metrics.ErrorRate() >= 0.99
	s.mu.Unlock()
}

func decodeLogEntry(raw []byte) (domain.LogEntry, error) {
	var enriched dto.EnrichedLog
	if err := json.Unmarshal(raw, &enriched); err == nil && enriched.Service != "" {
		return enriched.LogEntry, nil
	}
	var wrapper struct {
		Log domain.LogEntry `json:"log"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Log.Service != "" {
		return wrapper.Log, nil
	}
	var entry domain.LogEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return domain.LogEntry{}, err
	}
	return entry, nil
}

// Acknowledge acknowledges an incident.
func (s *Service) Acknowledge(ctx context.Context, tenantID string, id uuid.UUID) (*domain.Incident, error) {
	now := time.Now().UTC()
	if err := s.repo.AcknowledgeIncident(ctx, tenantID, id, now); err != nil {
		return nil, err
	}
	incident, err := s.repo.GetIncident(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	AppendTimelineEvent(incident, BuildTimelineEvent(domain.IncidentEventTypeStatusChange, "Incident acknowledged", "", now))
	_ = s.repo.UpdateIncident(ctx, incident)
	return incident, nil
}

// Resolve resolves an incident with notes.
func (s *Service) Resolve(ctx context.Context, tenantID string, id uuid.UUID, notes string) (*domain.Incident, error) {
	incident, err := s.repo.GetIncident(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	mttr := ComputeMTTR(incident, now)
	incident.Status = domain.IncidentStatusResolved
	incident.ResolvedTime = &now
	incident.MTTR = mttr
	AppendTimelineEvent(incident, BuildTimelineEvent(domain.IncidentEventTypeStatusChange, "Incident resolved", notes, now))

	if err := s.repo.ResolveIncident(ctx, tenantID, id, now, notes, mttr); err != nil {
		return nil, err
	}
	for _, service := range incident.AffectedServices {
		_ = s.repo.SaveMTTRHistory(ctx, id, service, "", mttr, incident.StartTime, now)
	}
	incident.Summary = BuildNarrative(*incident)
	return incident, nil
}

// GetRecommendations generates recommendations for an incident.
func (s *Service) GetRecommendations(ctx context.Context, tenantID string, id uuid.UUID) ([]domain.Recommendation, error) {
	incident, err := s.repo.GetIncident(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if len(incident.Recommendations) > 0 {
		return incident.Recommendations, nil
	}
	if s.llm == nil || incident.RootCauseAnalysis == nil {
		return incident.Recommendations, nil
	}
	recs, err := s.llm.GenerateRecommendation(ctx, *incident.RootCauseAnalysis)
	if err != nil {
		return nil, err
	}
	incident.Recommendations = recs
	_ = s.repo.UpdateIncident(ctx, incident)
	return recs, nil
}
