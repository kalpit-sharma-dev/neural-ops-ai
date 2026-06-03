// Package license validates offline NeuralOps license files (BANK-018).
package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Document is the signed offline license payload.
type Document struct {
	CustomerID string    `json:"customerId"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Features   []string  `json:"features,omitempty"`
	Signature  string    `json:"signature"`
}

// ValidateFile reads and validates a license JSON file.
func ValidateFile(path, secret string) (Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("read license: %w", err)
	}
	return ValidateJSON(raw, secret)
}

// ValidateJSON validates license bytes.
func ValidateJSON(raw []byte, secret string) (Document, error) {
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Document{}, fmt.Errorf("parse license: %w", err)
	}
	return Validate(doc, secret)
}

// Validate checks signature and expiry.
func Validate(doc Document, secret string) (Document, error) {
	if strings.TrimSpace(secret) == "" {
		return Document{}, fmt.Errorf("license secret not configured")
	}
	if doc.CustomerID == "" {
		return Document{}, fmt.Errorf("license customerId is required")
	}
	if doc.ExpiresAt.IsZero() {
		return Document{}, fmt.Errorf("license expiresAt is required")
	}
	if time.Now().UTC().After(doc.ExpiresAt) {
		return Document{}, fmt.Errorf("license expired at %s", doc.ExpiresAt.UTC().Format(time.RFC3339))
	}
	expected := sign(doc, secret)
	if !hmac.Equal([]byte(strings.ToLower(doc.Signature)), []byte(strings.ToLower(expected))) {
		return Document{}, fmt.Errorf("license signature invalid")
	}
	return doc, nil
}

func sign(doc Document, secret string) string {
	payload := struct {
		CustomerID string    `json:"customerId"`
		ExpiresAt  time.Time `json:"expiresAt"`
		Features   []string  `json:"features,omitempty"`
	}{
		CustomerID: doc.CustomerID,
		ExpiresAt:    doc.ExpiresAt.UTC(),
		Features:     doc.Features,
	}
	raw, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	return hex.EncodeToString(mac.Sum(nil))
}
