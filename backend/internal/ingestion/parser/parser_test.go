package parser

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/dto"
)

func testEnricher(t *testing.T) *Enricher {
	t.Helper()
	enricher, err := NewEnricher(config.ParserConfig{
		StackTraceLanguages: []string{"Java", "Golang", "Python", "NodeJS"},
		TraceIDPatterns:     []string{`(?i)traceId=([a-f0-9]{16,32})`},
		TxnIDPatterns:         []string{`(?i)(UPI[0-9]{12,20})`},
	}, "test-ingestor", "dc1", nil)
	if err != nil {
		t.Fatalf("NewEnricher: %v", err)
	}
	return enricher
}

func TestParseJavaStackTrace(t *testing.T) {
	message := `java.net.SocketTimeoutException: Read timed out
	at com.neuralops.upi.PaymentService.process(PaymentService.java:142)
	at com.neuralops.upi.Controller.handle(Controller.java:58)
	at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1043)`

	stack := parseJavaStackTrace(message)
	if stack == nil {
		t.Fatal("expected java stack trace")
	}
	if stack.ExceptionType != "java.net.SocketTimeoutException" {
		t.Fatalf("unexpected exception type: %s", stack.ExceptionType)
	}
	if len(stack.Frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(stack.Frames))
	}
	if stack.Frames[0].MethodName != "process" {
		t.Fatalf("unexpected method: %s", stack.Frames[0].MethodName)
	}
	if stack.Frames[0].IsThirdParty {
		t.Fatal("expected first party frame")
	}
	if !stack.Frames[2].IsThirdParty {
		t.Fatal("expected third party spring frame")
	}
}

func TestParseGolangStackTrace(t *testing.T) {
	message := `panic: runtime error: index out of range [5] with length 3
main.processPayment(0xc0000b4000)
	/app/internal/payment/service.go:88 +0x245
main.main()
	/app/cmd/api/main.go:21 +0x98`

	stack := parseGolangStackTrace(message)
	if stack == nil {
		t.Fatal("expected golang stack trace")
	}
	if stack.ExceptionType != "panic" {
		t.Fatalf("unexpected exception type: %s", stack.ExceptionType)
	}
	if len(stack.Frames) < 1 {
		t.Fatal("expected at least one frame")
	}
}

func TestParsePythonStackTrace(t *testing.T) {
	message := `Traceback (most recent call last):
  File "/app/services/ledger.py", line 42, in post_entry
    result = client.submit(payload)
ValueError: invalid account number`

	stack := parsePythonStackTrace(message)
	if stack == nil {
		t.Fatal("expected python stack trace")
	}
	if stack.ExceptionType != "ValueError" {
		t.Fatalf("unexpected exception type: %s", stack.ExceptionType)
	}
	if len(stack.Frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(stack.Frames))
	}
	if stack.Frames[0].FunctionName != "post_entry" {
		t.Fatalf("unexpected function: %s", stack.Frames[0].FunctionName)
	}
}

func TestParseNodeStackTrace(t *testing.T) {
	message := `TypeError: Cannot read properties of undefined (reading 'id')
    at PaymentController.process (/app/src/controllers/payment.js:88:15)
    at Layer.handle [as handle_request] (/app/node_modules/express/lib/router/layer.js:95:5)`

	stack := parseNodeStackTrace(message)
	if stack == nil {
		t.Fatal("expected node stack trace")
	}
	if stack.ExceptionType != "TypeError" {
		t.Fatalf("unexpected exception type: %s", stack.ExceptionType)
	}
	if len(stack.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(stack.Frames))
	}
	if !stack.Frames[1].IsThirdParty {
		t.Fatal("expected node_modules frame to be third party")
	}
}

func TestDetectSeverity(t *testing.T) {
	tests := []struct {
		message  string
		provided string
		want     domain.LogSeverity
	}{
		{"plain info message", "", domain.LogSeverityInfo},
		{"warn: slow query detected", "", domain.LogSeverityWarn},
		{"error while connecting to database", "", domain.LogSeverityError},
		{"debug payload", "DEBUG", domain.LogSeverityDebug},
	}

	for _, tc := range tests {
		got := detectSeverity(tc.message, tc.provided)
		if got != tc.want {
			t.Fatalf("detectSeverity(%q, %q) = %s, want %s", tc.message, tc.provided, got, tc.want)
		}
	}
}

func TestEnricherExtractsTraceAndTxnIDs(t *testing.T) {
	enricher := testEnricher(t)
	now := time.Now().UTC()

	enriched := enricher.Enrich(&dto.LogIngestRequest{
		Timestamp: now,
		Service:   "upi-service",
		Message:   "payment failed traceId=abc123def4567890 for UPI123456789012345",
	}, "http")

	if enriched.TraceID != "abc123def4567890" {
		t.Fatalf("unexpected trace id: %s", enriched.TraceID)
	}
	if enriched.TxnID != "UPI123456789012345" {
		t.Fatalf("unexpected txn id: %s", enriched.TxnID)
	}
	if enriched.IngestionMetadata.IngestorID != "test-ingestor" {
		t.Fatalf("unexpected ingestor id: %s", enriched.IngestionMetadata.IngestorID)
	}
}

func TestEnricherPromotesSeverityWhenStackTracePresent(t *testing.T) {
	enricher := testEnricher(t)
	now := time.Now().UTC()

	enriched := enricher.Enrich(&dto.LogIngestRequest{
		Timestamp: now,
		Service:   "upi-service",
		Severity:  "INFO",
		Message: `java.net.SocketTimeoutException: Read timed out
	at com.neuralops.upi.PaymentService.process(PaymentService.java:142)`,
	}, "http")

	if enriched.Severity != domain.LogSeverityError {
		t.Fatalf("expected severity ERROR, got %s", enriched.Severity)
	}
	if enriched.ParsedStackTrace == nil {
		t.Fatal("expected parsed stack trace")
	}
	if enriched.ParsedStackTrace.Hotspot == "" {
		t.Fatal("expected hotspot")
	}
}

func TestIdentifyHotspot(t *testing.T) {
	hotspot := identifyHotspot([]domain.StackFrame{
		{MethodName: "process", IsThirdParty: false},
		{MethodName: "dispatch", IsThirdParty: true},
	})
	if hotspot != "process" {
		t.Fatalf("unexpected hotspot: %s", hotspot)
	}
}
