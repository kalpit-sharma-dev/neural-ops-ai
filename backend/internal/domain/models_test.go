package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateMessageLength(t *testing.T) {
	if err := ValidateMessageLength("ok"); err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}

	longMessage := make([]byte, maxLogMessageLength+1)
	if err := ValidateMessageLength(string(longMessage)); err == nil {
		t.Fatal("expected validation error for oversized message")
	}
}

func TestValidateConfidence(t *testing.T) {
	if err := ValidateConfidence(0.5); err != nil {
		t.Fatalf("expected valid confidence, got %v", err)
	}
	if err := ValidateConfidence(-0.1); err == nil {
		t.Fatal("expected error for negative confidence")
	}
	if err := ValidateConfidence(1.1); err == nil {
		t.Fatal("expected error for confidence above 1")
	}
}

func TestIncidentComputeMTTR(t *testing.T) {
	start := time.Now().Add(-2 * time.Hour)
	resolved := start.Add(30 * time.Minute)

	incident := Incident{
		StartTime:    start,
		ResolvedTime: &resolved,
		Status:       IncidentStatusResolved,
	}
	incident.ComputeMTTR()

	if incident.MTTR != 30*time.Minute {
		t.Fatalf("expected MTTR 30m, got %v", incident.MTTR)
	}
}

func TestIncidentIsOpen(t *testing.T) {
	open := Incident{Status: IncidentStatusOpen}
	if !open.IsOpen() {
		t.Fatal("expected OPEN incident to be open")
	}

	resolved := Incident{Status: IncidentStatusResolved}
	if resolved.IsOpen() {
		t.Fatal("expected RESOLVED incident not to be open")
	}
}

func TestLogEntryIsErrorSeverity(t *testing.T) {
	entry := LogEntry{Severity: LogSeverityError}
	if !entry.IsErrorSeverity() {
		t.Fatal("expected ERROR to be error severity")
	}

	info := LogEntry{Severity: LogSeverityInfo}
	if info.IsErrorSeverity() {
		t.Fatal("expected INFO not to be error severity")
	}
}

func TestTransactionFailedHop(t *testing.T) {
	txn := Transaction{
		Hops: []TxnHop{
			{ServiceName: "auth-service", Status: HopStatusSuccess},
			{ServiceName: "upi-service", Status: HopStatusFailed, ErrorMessage: "timeout"},
		},
	}

	failed := txn.FailedHop()
	if failed == nil || failed.ServiceName != "upi-service" {
		t.Fatalf("expected failed hop upi-service, got %+v", failed)
	}
}

func TestLogEntryHasStackTrace(t *testing.T) {
	withStack := LogEntry{
		ParsedStackTrace: &StackTrace{
			Language: StackTraceLanguageJava,
			Frames:   []StackFrame{{FileName: "Main.java", LineNumber: 10}},
		},
	}
	if !withStack.HasStackTrace() {
		t.Fatal("expected stack trace to be detected")
	}

	empty := LogEntry{}
	if empty.HasStackTrace() {
		t.Fatal("expected no stack trace")
	}
}

func TestIncidentIDGeneration(t *testing.T) {
	id := uuid.New()
	if id == uuid.Nil {
		t.Fatal("expected non-nil UUID")
	}
}
