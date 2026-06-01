package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// Environment represents the deployment environment for logs and deployments.
type Environment string

const (
	EnvironmentProd    Environment = "prod"
	EnvironmentStaging Environment = "staging"
	EnvironmentDev     Environment = "dev"
)

func (e Environment) String() string { return string(e) }

func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentProd, EnvironmentStaging, EnvironmentDev:
		return true
	default:
		return false
	}
}

func (e *Environment) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*e = ""
		return nil
	case string:
		*e = Environment(v)
	case []byte:
		*e = Environment(string(v))
	default:
		return fmt.Errorf("domain: cannot scan Environment from %T", value)
	}
	if *e != "" && !e.IsValid() {
		return fmt.Errorf("domain: invalid environment %q", *e)
	}
	return nil
}

func (e Environment) Value() (driver.Value, error) {
	if e == "" {
		return nil, nil
	}
	if !e.IsValid() {
		return nil, fmt.Errorf("domain: invalid environment %q", e)
	}
	return string(e), nil
}

// LogSeverity represents log line severity levels.
type LogSeverity string

const (
	LogSeverityDebug    LogSeverity = "DEBUG"
	LogSeverityInfo     LogSeverity = "INFO"
	LogSeverityWarn     LogSeverity = "WARN"
	LogSeverityError    LogSeverity = "ERROR"
	LogSeverityFatal    LogSeverity = "FATAL"
	LogSeverityCritical LogSeverity = "CRITICAL"
)

func (s LogSeverity) String() string { return string(s) }

func (s LogSeverity) IsValid() bool {
	switch s {
	case LogSeverityDebug, LogSeverityInfo, LogSeverityWarn, LogSeverityError, LogSeverityFatal, LogSeverityCritical:
		return true
	default:
		return false
	}
}

func (s *LogSeverity) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = LogSeverity(strings.ToUpper(v))
	case []byte:
		*s = LogSeverity(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan LogSeverity from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid log severity %q", *s)
	}
	return nil
}

func (s LogSeverity) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid log severity %q", s)
	}
	return string(s), nil
}

// StackTraceLanguage identifies the runtime language of a parsed stack trace.
type StackTraceLanguage string

const (
	StackTraceLanguageJava   StackTraceLanguage = "Java"
	StackTraceLanguageGolang StackTraceLanguage = "Golang"
	StackTraceLanguagePython StackTraceLanguage = "Python"
	StackTraceLanguageNodeJS StackTraceLanguage = "NodeJS"
	StackTraceLanguageDotNet StackTraceLanguage = "DotNet"
)

func (l StackTraceLanguage) String() string { return string(l) }

func (l StackTraceLanguage) IsValid() bool {
	switch l {
	case StackTraceLanguageJava, StackTraceLanguageGolang, StackTraceLanguagePython,
		StackTraceLanguageNodeJS, StackTraceLanguageDotNet:
		return true
	default:
		return false
	}
}

func (l *StackTraceLanguage) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*l = ""
		return nil
	case string:
		*l = StackTraceLanguage(v)
	case []byte:
		*l = StackTraceLanguage(string(v))
	default:
		return fmt.Errorf("domain: cannot scan StackTraceLanguage from %T", value)
	}
	if *l != "" && !l.IsValid() {
		return fmt.Errorf("domain: invalid stack trace language %q", *l)
	}
	return nil
}

func (l StackTraceLanguage) Value() (driver.Value, error) {
	if l == "" {
		return nil, nil
	}
	if !l.IsValid() {
		return nil, fmt.Errorf("domain: invalid stack trace language %q", l)
	}
	return string(l), nil
}

// ErrorCategory classifies the root error pattern detected in logs.
type ErrorCategory string

const (
	ErrorCategoryTimeout               ErrorCategory = "TIMEOUT"
	ErrorCategoryDBDeadlock            ErrorCategory = "DB_DEADLOCK"
	ErrorCategoryMemoryLeak            ErrorCategory = "MEMORY_LEAK"
	ErrorCategoryConnectionExhaustion  ErrorCategory = "CONNECTION_EXHAUSTION"
	ErrorCategoryDNSIssue              ErrorCategory = "DNS_ISSUE"
	ErrorCategorySSLIssue              ErrorCategory = "SSL_ISSUE"
	ErrorCategoryRetryStorm            ErrorCategory = "RETRY_STORM"
	ErrorCategoryKafkaLag              ErrorCategory = "KAFKA_LAG"
	ErrorCategoryDeploymentIssue       ErrorCategory = "DEPLOYMENT_ISSUE"
	ErrorCategoryDependencyFailure     ErrorCategory = "DEPENDENCY_FAILURE"
	ErrorCategoryThreadStarvation      ErrorCategory = "THREAD_STARVATION"
	ErrorCategoryCircuitBreakerOpen    ErrorCategory = "CIRCUIT_BREAKER_OPEN"
	ErrorCategoryRateLimiting          ErrorCategory = "RATE_LIMITING"
	ErrorCategoryAuthFailure           ErrorCategory = "AUTH_FAILURE"
	ErrorCategoryUnknown               ErrorCategory = "UNKNOWN"
)

func (c ErrorCategory) String() string { return string(c) }

func (c ErrorCategory) IsValid() bool {
	switch c {
	case ErrorCategoryTimeout, ErrorCategoryDBDeadlock, ErrorCategoryMemoryLeak,
		ErrorCategoryConnectionExhaustion, ErrorCategoryDNSIssue, ErrorCategorySSLIssue,
		ErrorCategoryRetryStorm, ErrorCategoryKafkaLag, ErrorCategoryDeploymentIssue,
		ErrorCategoryDependencyFailure, ErrorCategoryThreadStarvation, ErrorCategoryCircuitBreakerOpen,
		ErrorCategoryRateLimiting, ErrorCategoryAuthFailure, ErrorCategoryUnknown:
		return true
	default:
		return false
	}
}

func (c *ErrorCategory) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*c = ""
		return nil
	case string:
		*c = ErrorCategory(strings.ToUpper(v))
	case []byte:
		*c = ErrorCategory(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan ErrorCategory from %T", value)
	}
	if *c != "" && !c.IsValid() {
		return fmt.Errorf("domain: invalid error category %q", *c)
	}
	return nil
}

func (c ErrorCategory) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	if !c.IsValid() {
		return nil, fmt.Errorf("domain: invalid error category %q", c)
	}
	return string(c), nil
}

// IncidentSeverity represents incident priority levels.
type IncidentSeverity string

const (
	IncidentSeverityP1 IncidentSeverity = "P1"
	IncidentSeverityP2 IncidentSeverity = "P2"
	IncidentSeverityP3 IncidentSeverity = "P3"
	IncidentSeverityP4 IncidentSeverity = "P4"
)

func (s IncidentSeverity) String() string { return string(s) }

func (s IncidentSeverity) IsValid() bool {
	switch s {
	case IncidentSeverityP1, IncidentSeverityP2, IncidentSeverityP3, IncidentSeverityP4:
		return true
	default:
		return false
	}
}

func (s *IncidentSeverity) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = IncidentSeverity(strings.ToUpper(v))
	case []byte:
		*s = IncidentSeverity(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan IncidentSeverity from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid incident severity %q", *s)
	}
	return nil
}

func (s IncidentSeverity) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid incident severity %q", s)
	}
	return string(s), nil
}

// IncidentStatus represents the lifecycle state of an incident.
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "OPEN"
	IncidentStatusInvestigating IncidentStatus = "INVESTIGATING"
	IncidentStatusResolved      IncidentStatus = "RESOLVED"
	IncidentStatusSuppressed    IncidentStatus = "SUPPRESSED"
)

func (s IncidentStatus) String() string { return string(s) }

func (s IncidentStatus) IsValid() bool {
	switch s {
	case IncidentStatusOpen, IncidentStatusInvestigating, IncidentStatusResolved, IncidentStatusSuppressed:
		return true
	default:
		return false
	}
}

func (s *IncidentStatus) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = IncidentStatus(strings.ToUpper(v))
	case []byte:
		*s = IncidentStatus(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan IncidentStatus from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid incident status %q", *s)
	}
	return nil
}

func (s IncidentStatus) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid incident status %q", s)
	}
	return string(s), nil
}

// RecommendationType categorizes remediation recommendations.
type RecommendationType string

const (
	RecommendationTypeConfig         RecommendationType = "CONFIG"
	RecommendationTypeTimeout      RecommendationType = "TIMEOUT"
	RecommendationTypeRetry          RecommendationType = "RETRY"
	RecommendationTypeAutoscaling    RecommendationType = "AUTOSCALING"
	RecommendationTypeDBTuning       RecommendationType = "DB_TUNING"
	RecommendationTypeJVM            RecommendationType = "JVM"
	RecommendationTypeCircuitBreaker RecommendationType = "CIRCUIT_BREAKER"
)

func (t RecommendationType) String() string { return string(t) }

func (t RecommendationType) IsValid() bool {
	switch t {
	case RecommendationTypeConfig, RecommendationTypeTimeout, RecommendationTypeRetry,
		RecommendationTypeAutoscaling, RecommendationTypeDBTuning, RecommendationTypeJVM,
		RecommendationTypeCircuitBreaker:
		return true
	default:
		return false
	}
}

func (t *RecommendationType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = RecommendationType(strings.ToUpper(v))
	case []byte:
		*t = RecommendationType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan RecommendationType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid recommendation type %q", *t)
	}
	return nil
}

func (t RecommendationType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid recommendation type %q", t)
	}
	return string(t), nil
}

// RecommendationPriority represents recommendation urgency.
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "HIGH"
	RecommendationPriorityMedium RecommendationPriority = "MEDIUM"
	RecommendationPriorityLow    RecommendationPriority = "LOW"
)

func (p RecommendationPriority) String() string { return string(p) }

func (p RecommendationPriority) IsValid() bool {
	switch p {
	case RecommendationPriorityHigh, RecommendationPriorityMedium, RecommendationPriorityLow:
		return true
	default:
		return false
	}
}

func (p *RecommendationPriority) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*p = ""
		return nil
	case string:
		*p = RecommendationPriority(strings.ToUpper(v))
	case []byte:
		*p = RecommendationPriority(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan RecommendationPriority from %T", value)
	}
	if *p != "" && !p.IsValid() {
		return fmt.Errorf("domain: invalid recommendation priority %q", *p)
	}
	return nil
}

func (p RecommendationPriority) Value() (driver.Value, error) {
	if p == "" {
		return nil, nil
	}
	if !p.IsValid() {
		return nil, fmt.Errorf("domain: invalid recommendation priority %q", p)
	}
	return string(p), nil
}

// ChangeType classifies deployment and configuration changes.
type ChangeType string

const (
	ChangeTypeDeployment    ChangeType = "DEPLOYMENT"
	ChangeTypeConfigChange  ChangeType = "CONFIG_CHANGE"
	ChangeTypeScaling       ChangeType = "SCALING"
	ChangeTypeFeatureToggle ChangeType = "FEATURE_TOGGLE"
)

func (c ChangeType) String() string { return string(c) }

func (c ChangeType) IsValid() bool {
	switch c {
	case ChangeTypeDeployment, ChangeTypeConfigChange, ChangeTypeScaling, ChangeTypeFeatureToggle:
		return true
	default:
		return false
	}
}

func (c *ChangeType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*c = ""
		return nil
	case string:
		*c = ChangeType(strings.ToUpper(v))
	case []byte:
		*c = ChangeType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan ChangeType from %T", value)
	}
	if *c != "" && !c.IsValid() {
		return fmt.Errorf("domain: invalid change type %q", *c)
	}
	return nil
}

func (c ChangeType) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	if !c.IsValid() {
		return nil, fmt.Errorf("domain: invalid change type %q", c)
	}
	return string(c), nil
}

// AlertSource identifies the origin system of an alert.
type AlertSource string

const (
	AlertSourcePrometheus AlertSource = "PROMETHEUS"
	AlertSourceDynatrace  AlertSource = "DYNATRACE"
	AlertSourceCloud      AlertSource = "CLOUD"
	AlertSourceCustom     AlertSource = "CUSTOM"
)

func (s AlertSource) String() string { return string(s) }

func (s AlertSource) IsValid() bool {
	switch s {
	case AlertSourcePrometheus, AlertSourceDynatrace, AlertSourceCloud, AlertSourceCustom:
		return true
	default:
		return false
	}
}

func (s *AlertSource) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = AlertSource(strings.ToUpper(v))
	case []byte:
		*s = AlertSource(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan AlertSource from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid alert source %q", *s)
	}
	return nil
}

func (s AlertSource) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid alert source %q", s)
	}
	return string(s), nil
}

// TxnType represents banking transaction rails.
type TxnType string

const (
	TxnTypeUPI  TxnType = "UPI"
	TxnTypeNEFT TxnType = "NEFT"
	TxnTypeRTGS TxnType = "RTGS"
	TxnTypeIMPS TxnType = "IMPS"
)

func (t TxnType) String() string { return string(t) }

func (t TxnType) IsValid() bool {
	switch t {
	case TxnTypeUPI, TxnTypeNEFT, TxnTypeRTGS, TxnTypeIMPS:
		return true
	default:
		return false
	}
}

func (t *TxnType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = TxnType(strings.ToUpper(v))
	case []byte:
		*t = TxnType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan TxnType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid transaction type %q", *t)
	}
	return nil
}

func (t TxnType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid transaction type %q", t)
	}
	return string(t), nil
}

// TxnStatus represents the outcome of a banking transaction.
type TxnStatus string

const (
	TxnStatusSuccess TxnStatus = "SUCCESS"
	TxnStatusFailed  TxnStatus = "FAILED"
	TxnStatusPartial TxnStatus = "PARTIAL"
	TxnStatusPending TxnStatus = "PENDING"
)

func (s TxnStatus) String() string { return string(s) }

func (s TxnStatus) IsValid() bool {
	switch s {
	case TxnStatusSuccess, TxnStatusFailed, TxnStatusPartial, TxnStatusPending:
		return true
	default:
		return false
	}
}

func (s *TxnStatus) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = TxnStatus(strings.ToUpper(v))
	case []byte:
		*s = TxnStatus(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan TxnStatus from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid transaction status %q", *s)
	}
	return nil
}

func (s TxnStatus) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid transaction status %q", s)
	}
	return string(s), nil
}

// HopStatus represents the outcome of an individual transaction hop.
type HopStatus string

const (
	HopStatusSuccess HopStatus = "SUCCESS"
	HopStatusFailed  HopStatus = "FAILED"
	HopStatusPartial HopStatus = "PARTIAL"
	HopStatusPending HopStatus = "PENDING"
)

func (s HopStatus) String() string { return string(s) }

func (s HopStatus) IsValid() bool {
	switch s {
	case HopStatusSuccess, HopStatusFailed, HopStatusPartial, HopStatusPending:
		return true
	default:
		return false
	}
}

func (s *HopStatus) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = HopStatus(strings.ToUpper(v))
	case []byte:
		*s = HopStatus(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan HopStatus from %T", value)
	}
	if *s != "" && !s.IsValid() {
		return fmt.Errorf("domain: invalid hop status %q", *s)
	}
	return nil
}

func (s HopStatus) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	if !s.IsValid() {
		return nil, fmt.Errorf("domain: invalid hop status %q", s)
	}
	return string(s), nil
}

// MetricType classifies infrastructure and application metrics.
type MetricType string

const (
	MetricTypeCPU           MetricType = "CPU"
	MetricTypeMemory        MetricType = "MEMORY"
	MetricTypeLatency       MetricType = "LATENCY"
	MetricTypeThroughput    MetricType = "THROUGHPUT"
	MetricTypeDBConnections MetricType = "DB_CONNECTIONS"
	MetricTypeThreadPool    MetricType = "THREAD_POOL"
	MetricTypeGCPause       MetricType = "GC_PAUSE"
)

func (t MetricType) String() string { return string(t) }

func (t MetricType) IsValid() bool {
	switch t {
	case MetricTypeCPU, MetricTypeMemory, MetricTypeLatency, MetricTypeThroughput,
		MetricTypeDBConnections, MetricTypeThreadPool, MetricTypeGCPause:
		return true
	default:
		return false
	}
}

func (t *MetricType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = MetricType(strings.ToUpper(v))
	case []byte:
		*t = MetricType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan MetricType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid metric type %q", *t)
	}
	return nil
}

func (t MetricType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid metric type %q", t)
	}
	return string(t), nil
}

// AnomalyType classifies detected metric anomaly patterns.
type AnomalyType string

const (
	AnomalyTypeSpike    AnomalyType = "SPIKE"
	AnomalyTypeDrop     AnomalyType = "DROP"
	AnomalyTypeTrend    AnomalyType = "TREND"
	AnomalyTypeSeasonal AnomalyType = "SEASONAL"
)

func (t AnomalyType) String() string { return string(t) }

func (t AnomalyType) IsValid() bool {
	switch t {
	case AnomalyTypeSpike, AnomalyTypeDrop, AnomalyTypeTrend, AnomalyTypeSeasonal:
		return true
	default:
		return false
	}
}

func (t *AnomalyType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = AnomalyType(strings.ToUpper(v))
	case []byte:
		*t = AnomalyType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan AnomalyType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid anomaly type %q", *t)
	}
	return nil
}

func (t AnomalyType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid anomaly type %q", t)
	}
	return string(t), nil
}

// EvidenceType categorizes RCA evidence artifacts.
type EvidenceType string

const (
	EvidenceTypeLog        EvidenceType = "LOG"
	EvidenceTypeMetric     EvidenceType = "METRIC"
	EvidenceTypeDeployment EvidenceType = "DEPLOYMENT"
	EvidenceTypeAlert      EvidenceType = "ALERT"
	EvidenceTypeTrace      EvidenceType = "TRACE"
)

func (t EvidenceType) String() string { return string(t) }

func (t EvidenceType) IsValid() bool {
	switch t {
	case EvidenceTypeLog, EvidenceTypeMetric, EvidenceTypeDeployment, EvidenceTypeAlert, EvidenceTypeTrace:
		return true
	default:
		return false
	}
}

func (t *EvidenceType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = EvidenceType(strings.ToUpper(v))
	case []byte:
		*t = EvidenceType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan EvidenceType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid evidence type %q", *t)
	}
	return nil
}

func (t EvidenceType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid evidence type %q", t)
	}
	return string(t), nil
}

// IncidentEventType categorizes timeline events on an incident.
type IncidentEventType string

const (
	IncidentEventTypeDeployment   IncidentEventType = "DEPLOYMENT"
	IncidentEventTypeError        IncidentEventType = "ERROR"
	IncidentEventTypeMetric       IncidentEventType = "METRIC"
	IncidentEventTypeAlert        IncidentEventType = "ALERT"
	IncidentEventTypeStatusChange IncidentEventType = "STATUS_CHANGE"
	IncidentEventTypeComment      IncidentEventType = "COMMENT"
)

func (t IncidentEventType) String() string { return string(t) }

func (t IncidentEventType) IsValid() bool {
	switch t {
	case IncidentEventTypeDeployment, IncidentEventTypeError, IncidentEventTypeMetric,
		IncidentEventTypeAlert, IncidentEventTypeStatusChange, IncidentEventTypeComment:
		return true
	default:
		return false
	}
}

func (t *IncidentEventType) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ""
		return nil
	case string:
		*t = IncidentEventType(strings.ToUpper(v))
	case []byte:
		*t = IncidentEventType(strings.ToUpper(string(v)))
	default:
		return fmt.Errorf("domain: cannot scan IncidentEventType from %T", value)
	}
	if *t != "" && !t.IsValid() {
		return fmt.Errorf("domain: invalid incident event type %q", *t)
	}
	return nil
}

func (t IncidentEventType) Value() (driver.Value, error) {
	if t == "" {
		return nil, nil
	}
	if !t.IsValid() {
		return nil, fmt.Errorf("domain: invalid incident event type %q", t)
	}
	return string(t), nil
}
