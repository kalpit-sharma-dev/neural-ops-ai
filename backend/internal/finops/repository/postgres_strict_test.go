package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/finops/repository"
)

func TestPostgresStoreStrictRequiresPool(t *testing.T) {
	store := repository.NewPostgresStore(nil, repository.NewMemoryStore(), true)
	_, err := store.ListLineItems(context.Background(), "default", time.Now().Add(-24*time.Hour), time.Now(), "", "")
	if err == nil {
		t.Fatal("expected error when strict mode without pool")
	}
}

func TestPostgresStoreStrictNoMemoryFallbackOnEmpty(t *testing.T) {
	mem := repository.NewMemoryStore()
	_ = mem.SaveLineItems(context.Background(), "default", []domain.CostLineItem{
		{ID: "sim-1", BillingPeriod: time.Now().UTC(), Provider: "aws", AmortizedCost: 99},
	})
	strict := repository.NewPostgresStore(nil, mem, true)
	_, err := strict.ListLineItems(context.Background(), "default", time.Now().Add(-48*time.Hour), time.Now(), "", "")
	if err == nil {
		t.Fatal("expected error without pool in strict mode")
	}
}
