package repository

import (
	"context"

	correlationrepo "github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/domain"
)

// TransactionStore reads transaction journeys from ClickHouse.
type TransactionStore struct {
	store *correlationrepo.ClickHouseStore
}

// NewTransactionStore creates a transaction store.
func NewTransactionStore(ctx context.Context, dsn, table string) (*TransactionStore, error) {
	store, err := correlationrepo.NewClickHouseStore(ctx, dsn, table)
	if err != nil {
		return nil, err
	}
	return &TransactionStore{store: store}, nil
}

// GetTransaction returns a transaction journey.
func (s *TransactionStore) GetTransaction(ctx context.Context, txnID string) (*domain.Transaction, error) {
	return s.store.GetTransaction(ctx, txnID)
}
