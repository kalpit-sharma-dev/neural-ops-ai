package mobile_test

import (
	"testing"

	"github.com/neuralops/platform/internal/mobile"
)

func TestNewExpoClientNilPool(t *testing.T) {
	c := mobile.NewExpoClient(nil)
	if c == nil {
		t.Fatal("expected client")
	}
	_, err := c.ListTokens(t.Context(), "default")
	if err == nil {
		t.Fatal("expected error without pool")
	}
}
