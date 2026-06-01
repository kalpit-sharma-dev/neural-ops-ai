package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Domain error codes follow the platform convention: DOM### .
const (
	ErrCodeNotFound           ErrorCode = "DOM001"
	ErrCodeValidation         ErrorCode = "DOM002"
	ErrCodeConflict           ErrorCode = "DOM003"
	ErrCodeInvalidState       ErrorCode = "DOM004"
	ErrCodeInvalidTransition  ErrorCode = "DOM005"
	ErrCodeDuplicate          ErrorCode = "DOM006"
	ErrCodeUnauthorized       ErrorCode = "DOM007"
	ErrCodeForbidden          ErrorCode = "DOM008"
	ErrCodeInvariantViolation ErrorCode = "DOM009"
	ErrCodeExternalDependency ErrorCode = "DOM010"
)

// ErrorCode is a stable identifier for domain errors.
type ErrorCode string

func (c ErrorCode) String() string {
	return string(c)
}

// DomainError represents a typed domain-layer failure.
type DomainError struct {
	Code    ErrorCode `json:"errorCode"`
	Message string    `json:"message"`
	Field   string    `json:"field,omitempty"`
	Err     error     `json:"-"`
}

func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a domain error with the given code and message.
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a field-level validation error.
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeValidation,
		Message: message,
		Field:   field,
	}
}

// NewNotFoundError creates a not-found error for the given entity.
func NewNotFoundError(entity, id string) *DomainError {
	return &DomainError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found: %s", entity, id),
	}
}

// NewConflictError creates a conflict error for duplicate or concurrent updates.
func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// NewInvalidStateError creates an error when an entity is in an invalid state for the operation.
func NewInvalidStateError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidState,
		Message: message,
	}
}

// NewInvalidTransitionError creates an error for invalid state machine transitions.
func NewInvalidTransitionError(from, to string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidTransition,
		Message: fmt.Sprintf("invalid transition from %s to %s", from, to),
	}
}

// NewDuplicateError creates an error when a unique entity already exists.
func NewDuplicateError(entity, key string) *DomainError {
	return &DomainError{
		Code:    ErrCodeDuplicate,
		Message: fmt.Sprintf("%s already exists: %s", entity, key),
	}
}

// NewInvariantViolationError creates an error when domain invariants are violated.
func NewInvariantViolationError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvariantViolation,
		Message: message,
	}
}

// WrapExternalDependencyError wraps an upstream dependency failure.
func WrapExternalDependencyError(service string, err error) *DomainError {
	return &DomainError{
		Code:    ErrCodeExternalDependency,
		Message: fmt.Sprintf("external dependency failed: %s", service),
		Err:     err,
	}
}

// IsDomainError reports whether err is a DomainError.
func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}

// AsDomainError extracts a DomainError from err if present.
func AsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

// IsNotFound reports whether err is a not-found domain error.
func IsNotFound(err error) bool {
	domainErr, ok := AsDomainError(err)
	return ok && domainErr.Code == ErrCodeNotFound
}

// IsValidation reports whether err is a validation domain error.
func IsValidation(err error) bool {
	domainErr, ok := AsDomainError(err)
	return ok && domainErr.Code == ErrCodeValidation
}

// ValidateIncidentTransition validates allowed incident status transitions.
func ValidateIncidentTransition(from, to IncidentStatus) error {
	if from == to {
		return nil
	}

	allowed := map[IncidentStatus][]IncidentStatus{
		IncidentStatusOpen: {
			IncidentStatusInvestigating,
			IncidentStatusSuppressed,
			IncidentStatusResolved,
		},
		IncidentStatusInvestigating: {
			IncidentStatusResolved,
			IncidentStatusSuppressed,
			IncidentStatusOpen,
		},
		IncidentStatusSuppressed: {
			IncidentStatusOpen,
			IncidentStatusInvestigating,
		},
		IncidentStatusResolved: {
			IncidentStatusOpen,
		},
	}

	nextStates, ok := allowed[from]
	if !ok {
		return NewInvalidStateError(fmt.Sprintf("unknown incident status %q", from))
	}

	for _, state := range nextStates {
		if state == to {
			return nil
		}
	}

	return NewInvalidTransitionError(string(from), string(to))
}

// ValidateLogEntry performs basic domain validation on a log entry.
func ValidateLogEntry(entry *LogEntry) error {
	if entry == nil {
		return NewValidationError("logEntry", "must not be nil")
	}
	if entry.ID == uuid.Nil {
		return NewValidationError("id", "is required")
	}
	if entry.Service == "" {
		return NewValidationError("service", "is required")
	}
	if entry.Message == "" {
		return NewValidationError("message", "is required")
	}
	if entry.Timestamp.IsZero() {
		return NewValidationError("timestamp", "is required")
	}
	if !entry.Environment.IsValid() {
		return NewValidationError("environment", "must be prod, staging, or dev")
	}
	if !entry.Severity.IsValid() {
		return NewValidationError("severity", "is invalid")
	}
	if err := ValidateMessageLength(entry.Message); err != nil {
		return err
	}
	if entry.ClassifiedError != nil {
		if err := ValidateConfidence(entry.ClassifiedError.Confidence); err != nil {
			return err
		}
		if !entry.ClassifiedError.Category.IsValid() {
			return NewValidationError("classifiedError.category", "is invalid")
		}
	}
	return nil
}

// ValidateIncident performs basic domain validation on an incident.
func ValidateIncident(incident *Incident) error {
	if incident == nil {
		return NewValidationError("incident", "must not be nil")
	}
	if incident.ID == uuid.Nil {
		return NewValidationError("id", "is required")
	}
	if incident.Title == "" {
		return NewValidationError("title", "is required")
	}
	if incident.Summary == "" {
		return NewValidationError("summary", "is required")
	}
	if !incident.Severity.IsValid() {
		return NewValidationError("severity", "is invalid")
	}
	if !incident.Status.IsValid() {
		return NewValidationError("status", "is invalid")
	}
	if len(incident.AffectedServices) == 0 {
		return NewValidationError("affectedServices", "must contain at least one service")
	}
	if incident.StartTime.IsZero() {
		return NewValidationError("startTime", "is required")
	}
	if incident.RootCauseAnalysis != nil {
		if err := ValidateConfidence(incident.RootCauseAnalysis.Confidence); err != nil {
			return err
		}
	}
	return nil
}
