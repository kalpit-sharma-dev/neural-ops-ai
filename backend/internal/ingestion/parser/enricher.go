package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/dto"
)

var thirdPartyPrefixes = []string{
	"java.", "javax.", "sun.", "org.springframework.", "com.google.",
	"node_modules/", "github.com/", "golang.org/", "google.golang.org/",
}

// Enricher parses and enriches log entries before publishing.
type Enricher struct {
	tracePatterns []*regexp.Regexp
	txnPatterns   []*regexp.Regexp
	languages     map[domain.StackTraceLanguage]bool
	ingestorID    string
	datacenter    string
	geoIP         GeoIPResolver
}

// GeoIPResolver optionally resolves geographic metadata from host IPs.
type GeoIPResolver interface {
	Lookup(ip string) (country, city string)
}

type noopGeoIP struct{}

func (noopGeoIP) Lookup(string) (string, string) { return "", "" }

// NewEnricher creates a log enricher from parser configuration.
func NewEnricher(cfg config.ParserConfig, ingestorID, datacenter string, geoIP GeoIPResolver) (*Enricher, error) {
	tracePatterns, err := compilePatterns(cfg.TraceIDPatterns)
	if err != nil {
		return nil, err
	}
	txnPatterns, err := compilePatterns(cfg.TxnIDPatterns)
	if err != nil {
		return nil, err
	}

	languages := make(map[domain.StackTraceLanguage]bool)
	for _, lang := range cfg.StackTraceLanguages {
		languages[domain.StackTraceLanguage(lang)] = true
	}
	if len(languages) == 0 {
		languages[domain.StackTraceLanguageJava] = true
		languages[domain.StackTraceLanguageGolang] = true
		languages[domain.StackTraceLanguagePython] = true
		languages[domain.StackTraceLanguageNodeJS] = true
	}

	if geoIP == nil {
		geoIP = noopGeoIP{}
	}

	return &Enricher{
		tracePatterns: tracePatterns,
		txnPatterns:   txnPatterns,
		languages:     languages,
		ingestorID:    ingestorID,
		datacenter:    datacenter,
		geoIP:         geoIP,
	}, nil
}

// Enrich converts an ingest request into a domain log entry with metadata.
func (e *Enricher) Enrich(req *dto.LogIngestRequest, source string) dto.EnrichedLog {
	now := time.Now().UTC()

	env := domain.EnvironmentDev
	if req.Environment != "" {
		env = domain.Environment(req.Environment)
	}

	severity := detectSeverity(req.Message, req.Severity)
	traceID := req.TraceID
	if traceID == "" {
		traceID = extractFirst(e.tracePatterns, req.Message)
	}
	txnID := req.TxnID
	if txnID == "" {
		txnID = extractFirst(e.txnPatterns, req.Message)
	}

	entry := domain.LogEntry{
		ID:          uuid.New(),
		Timestamp:   req.Timestamp.UTC(),
		Service:     req.Service,
		Environment: env,
		Severity:    severity,
		Message:     req.Message,
		TraceID:     traceID,
		TxnID:       txnID,
		Thread:      req.Thread,
		Host:        req.Host,
		Pod:         req.Pod,
		Namespace:   req.Namespace,
		Labels:      req.Labels,
		RawJSON:     req.RawJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if stack := e.parseStackTrace(req.Message); stack != nil {
		entry.ParsedStackTrace = stack
		if entry.Severity == domain.LogSeverityInfo || entry.Severity == domain.LogSeverityDebug {
			entry.Severity = domain.LogSeverityError
		}
	}

	meta := dto.IngestionMetadata{
		ReceivedAt: now,
		IngestorID: e.ingestorID,
		Datacenter: e.datacenter,
		TenantID:   req.TenantID,
		Source:     source,
	}

	if req.Host != "" && looksLikeIP(req.Host) {
		country, city := e.geoIP.Lookup(req.Host)
		meta.GeoCountry = country
		meta.GeoCity = city
	}

	return dto.EnrichedLog{
		LogEntry:          entry,
		IngestionMetadata: meta,
	}
}

func (e *Enricher) parseStackTrace(message string) *domain.StackTrace {
	for lang, enabled := range e.languages {
		if !enabled {
			continue
		}
		var stack *domain.StackTrace
		switch lang {
		case domain.StackTraceLanguageJava:
			stack = parseJavaStackTrace(message)
		case domain.StackTraceLanguageGolang:
			stack = parseGolangStackTrace(message)
		case domain.StackTraceLanguagePython:
			stack = parsePythonStackTrace(message)
		case domain.StackTraceLanguageNodeJS:
			stack = parseNodeStackTrace(message)
		}
		if stack != nil && len(stack.Frames) > 0 {
			stack.Hotspot = identifyHotspot(stack.Frames)
			return stack
		}
	}
	return nil
}

func detectSeverity(message, provided string) domain.LogSeverity {
	if provided != "" {
		sev := domain.LogSeverity(strings.ToUpper(provided))
		if sev.IsValid() {
			return sev
		}
	}

	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "fatal") || strings.Contains(lower, "panic"):
		return domain.LogSeverityFatal
	case strings.Contains(lower, "critical"):
		return domain.LogSeverityCritical
	case strings.Contains(lower, "error") || strings.Contains(lower, "exception"):
		return domain.LogSeverityError
	case strings.Contains(lower, "warn") || strings.Contains(lower, "warning"):
		return domain.LogSeverityWarn
	case strings.Contains(lower, "debug"):
		return domain.LogSeverityDebug
	default:
		return domain.LogSeverityInfo
	}
}

func extractFirst(patterns []*regexp.Regexp, message string) string {
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func compilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

func identifyHotspot(frames []domain.StackFrame) string {
	for _, frame := range frames {
		if !frame.IsThirdParty {
			if frame.MethodName != "" {
				return frame.MethodName
			}
			if frame.FunctionName != "" {
				return frame.FunctionName
			}
		}
	}
	if len(frames) > 0 {
		frame := frames[0]
		if frame.MethodName != "" {
			return frame.MethodName
		}
		return frame.FunctionName
	}
	return ""
}

func isThirdParty(path string) bool {
	for _, prefix := range thirdPartyPrefixes {
		if strings.Contains(path, prefix) {
			return true
		}
	}
	return false
}

func looksLikeIP(host string) bool {
	return regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`).MatchString(host)
}
