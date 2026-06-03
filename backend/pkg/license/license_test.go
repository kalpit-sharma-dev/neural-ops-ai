package license_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/neuralops/platform/pkg/license"
)

func TestValidateLicense(t *testing.T) {
	secret := "test-secret-key"
	expires := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	unsigned := struct {
		CustomerID string    `json:"customerId"`
		ExpiresAt  time.Time `json:"expiresAt"`
		Features   []string  `json:"features,omitempty"`
	}{
		CustomerID: "bank-demo",
		ExpiresAt:  expires,
		Features:   []string{"observability", "finops"},
	}
	uRaw, _ := json.Marshal(unsigned)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(uRaw)
	doc := license.Document{
		CustomerID: unsigned.CustomerID,
		ExpiresAt:  expires,
		Features:   unsigned.Features,
		Signature:  hex.EncodeToString(mac.Sum(nil)),
	}
	signed, _ := json.Marshal(doc)

	got, err := license.ValidateJSON(signed, secret)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got.CustomerID != "bank-demo" {
		t.Fatalf("customerId=%q", got.CustomerID)
	}
}

func TestValidateLicenseExpired(t *testing.T) {
	secret := "test-secret-key"
	expires := time.Now().UTC().Add(-time.Hour)
	unsigned := struct {
		CustomerID string    `json:"customerId"`
		ExpiresAt  time.Time `json:"expiresAt"`
	}{CustomerID: "bank-demo", ExpiresAt: expires}
	uRaw, _ := json.Marshal(unsigned)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(uRaw)
	doc := license.Document{
		CustomerID: "bank-demo",
		ExpiresAt:  expires,
		Signature:  hex.EncodeToString(mac.Sum(nil)),
	}
	signed, _ := json.Marshal(doc)
	if _, err := license.ValidateJSON(signed, secret); err == nil {
		t.Fatal("expected expired license error")
	}
}
