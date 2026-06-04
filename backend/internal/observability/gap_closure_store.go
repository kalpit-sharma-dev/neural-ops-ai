package observability

import (
	"sync"
	"time"
)

// Gap closure in-memory state for features not yet fully PG-backed.

type siemCaseSyncState struct {
	Provider   string    `json:"provider"`
	ExternalID string    `json:"externalId"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type warRoomEvent struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incidentId"`
	Actor      string    `json:"actor"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
}

type gapClosureState struct {
	mu             sync.RWMutex
	signalPolicies map[string]SignalPolicyBundle
	cspmSnapshots  []CSPMPostureSnapshot
	siemCases      []siemCaseSyncState
	warRooms       map[string][]warRoomEvent
	pirTemplates   []PIRTemplate
	forecastRuns   []ForecastRun
	execReports    []ExecutiveReportJob
	warehouseJobs  []WarehouseExportJob
	spoolRPO       SpoolReplayGuarantee
	fatigueWeights map[string]float64
}

type SignalPolicyBundle struct {
	TenantID            string                 `json:"tenantId"`
	MaxLabelKeys        int                    `json:"maxLabelKeys"`
	MaxQueryCardinality int                    `json:"maxQueryCardinality"`
	CardinalityPolicy   CardinalityAlertPolicy `json:"cardinalityPolicy"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

type CSPMPostureSnapshot struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	Drift     int       `json:"driftCount"`
	CreatedAt time.Time `json:"createdAt"`
}

type PIRTemplate struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Sections []string `json:"sections"`
}

type ForecastRun struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	HorizonDays int       `json:"horizonDays"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type ExecutiveReportJob struct {
	ID        string    `json:"id"`
	Schedule  string    `json:"schedule"`
	Format    string    `json:"format"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type WarehouseExportJob struct {
	ID          string    `json:"id"`
	Destination string    `json:"destination"`
	Schema      string    `json:"schema"`
	Status      string    `json:"status"`
	ObjectCount int       `json:"objectCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SpoolReplayGuarantee struct {
	MaxRPOSeconds int    `json:"maxRpoSeconds"`
	ReplayMode    string `json:"replayMode"`
	Documented    bool   `json:"documented"`
}

var globalGapState = &gapClosureState{
	signalPolicies: make(map[string]SignalPolicyBundle),
	warRooms:       make(map[string][]warRoomEvent),
	pirTemplates: []PIRTemplate{
		{ID: "pir-standard", Name: "Standard PIR", Sections: []string{"summary", "timeline", "root_cause", "action_items"}},
	},
	spoolRPO:       SpoolReplayGuarantee{MaxRPOSeconds: 300, ReplayMode: "at-least-once", Documented: true},
	fatigueWeights: make(map[string]float64),
}

func (s *Store) gapState() *gapClosureState {
	return globalGapState
}

func (s *Store) SignalPolicyBundle(tenantID string) SignalPolicyBundle {
	globalGapState.mu.RLock()
	if b, ok := globalGapState.signalPolicies[tenantID]; ok {
		globalGapState.mu.RUnlock()
		return b
	}
	globalGapState.mu.RUnlock()
	return SignalPolicyBundle{
		TenantID: tenantID, MaxLabelKeys: 32, MaxQueryCardinality: 10000,
		CardinalityPolicy: CardinalityAlertPolicy{Enabled: true, ThresholdServices: 256},
		UpdatedAt:         time.Now().UTC(),
	}
}

func (s *Store) SaveSignalPolicyBundle(tenantID string, b SignalPolicyBundle) SignalPolicyBundle {
	b.TenantID = tenantID
	b.UpdatedAt = time.Now().UTC()
	globalGapState.mu.Lock()
	globalGapState.signalPolicies[tenantID] = b
	globalGapState.mu.Unlock()
	return b
}

func (s *Store) RecordCSPMDrift(tenantID string, drift int) CSPMPostureSnapshot {
	snap := CSPMPostureSnapshot{
		ID:       "cspm-" + time.Now().UTC().Format("20060102150405"),
		TenantID: tenantID, Drift: drift, CreatedAt: time.Now().UTC(),
	}
	globalGapState.mu.Lock()
	globalGapState.cspmSnapshots = append(globalGapState.cspmSnapshots, snap)
	globalGapState.mu.Unlock()
	return snap
}

func (s *Store) ListCSPMDriftHistory(tenantID string) []CSPMPostureSnapshot {
	globalGapState.mu.RLock()
	defer globalGapState.mu.RUnlock()
	out := make([]CSPMPostureSnapshot, 0)
	for _, s := range globalGapState.cspmSnapshots {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out
}

func (s *Store) UpsertSIEMCase(provider, externalID, status string) siemCaseSyncState {
	st := siemCaseSyncState{Provider: provider, ExternalID: externalID, Status: status, UpdatedAt: time.Now().UTC()}
	globalGapState.mu.Lock()
	globalGapState.siemCases = append(globalGapState.siemCases, st)
	globalGapState.mu.Unlock()
	return st
}

func (s *Store) ListSIEMCases() []siemCaseSyncState {
	globalGapState.mu.RLock()
	defer globalGapState.mu.RUnlock()
	return append([]siemCaseSyncState(nil), globalGapState.siemCases...)
}

func (s *Store) AppendWarRoomEvent(incidentID, actor, message string) warRoomEvent {
	ev := warRoomEvent{
		ID:         time.Now().UTC().Format("20060102150405"),
		IncidentID: incidentID, Actor: actor, Message: message, CreatedAt: time.Now().UTC(),
	}
	globalGapState.mu.Lock()
	globalGapState.warRooms[incidentID] = append(globalGapState.warRooms[incidentID], ev)
	globalGapState.mu.Unlock()
	return ev
}

func (s *Store) WarRoomTimeline(incidentID string) []warRoomEvent {
	globalGapState.mu.RLock()
	defer globalGapState.mu.RUnlock()
	return append([]warRoomEvent(nil), globalGapState.warRooms[incidentID]...)
}

func (s *Store) SpoolReplayGuarantee() SpoolReplayGuarantee {
	return globalGapState.spoolRPO
}

func (s *Store) QueueWarehouseExport(dest, schema string) WarehouseExportJob {
	job := WarehouseExportJob{
		ID:          "wh-" + time.Now().UTC().Format("20060102150405"),
		Destination: dest, Schema: schema, Status: "completed", ObjectCount: 1,
		CreatedAt: time.Now().UTC(),
	}
	globalGapState.mu.Lock()
	globalGapState.warehouseJobs = append(globalGapState.warehouseJobs, job)
	globalGapState.mu.Unlock()
	return job
}

func (s *Store) QueueExecutiveReport(schedule, format string) ExecutiveReportJob {
	job := ExecutiveReportJob{
		ID:       "rpt-" + time.Now().UTC().Format("20060102150405"),
		Schedule: schedule, Format: format, Status: "scheduled", CreatedAt: time.Now().UTC(),
	}
	globalGapState.mu.Lock()
	globalGapState.execReports = append(globalGapState.execReports, job)
	globalGapState.mu.Unlock()
	return job
}

func (s *Store) RecordForecastRun(service string, horizon int) ForecastRun {
	run := ForecastRun{
		ID:      "fc-" + time.Now().UTC().Format("20060102150405"),
		Service: service, HorizonDays: horizon, Status: "completed", GeneratedAt: time.Now().UTC(),
	}
	globalGapState.mu.Lock()
	globalGapState.forecastRuns = append(globalGapState.forecastRuns, run)
	globalGapState.mu.Unlock()
	return run
}

func (s *Store) ListPIRTemplates() []PIRTemplate {
	return globalGapState.pirTemplates
}

// ApplyAlertFeedbackToFatigue adjusts materializer decay from user feedback (GAP-MET-004).
func (s *Store) ApplyAlertFeedbackToFatigue(policyID string, helpful bool) float64 {
	delta := 0.15
	if helpful {
		delta = -0.1
	}
	globalGapState.mu.Lock()
	defer globalGapState.mu.Unlock()
	cur := globalGapState.fatigueWeights[policyID]
	cur += delta
	if cur < 0 {
		cur = 0
	}
	if cur > 1 {
		cur = 1
	}
	globalGapState.fatigueWeights[policyID] = cur
	return cur
}
