package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
)

// PostgresStore persists analysis results in PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a PostgreSQL repository.
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	store := &PostgresStore{pool: pool}
	if err := store.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}

// Migrate applies analysis schema migrations.
func (s *PostgresStore) Migrate(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS error_classifications (
			id UUID PRIMARY KEY,
			log_id UUID NOT NULL,
			service VARCHAR(255) NOT NULL,
			category VARCHAR(64) NOT NULL,
			confidence DOUBLE PRECISION NOT NULL,
			reasoning TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_error_classifications_service ON error_classifications(service)`,
		`CREATE INDEX IF NOT EXISTS idx_error_classifications_log_id ON error_classifications(log_id)`,
		`CREATE TABLE IF NOT EXISTS log_explanations (
			id UUID PRIMARY KEY,
			log_id UUID NOT NULL,
			service VARCHAR(255) NOT NULL,
			plain_english TEXT NOT NULL,
			technical_summary TEXT NOT NULL,
			severity_assessment VARCHAR(32) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS recommendations (
			id UUID PRIMARY KEY,
			incident_id UUID,
			type VARCHAR(64) NOT NULL,
			description TEXT NOT NULL,
			code_snippet TEXT,
			priority VARCHAR(16) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}
	for _, query := range queries {
		if _, err := s.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// SaveClassification stores an error classification.
func (s *PostgresStore) SaveClassification(ctx context.Context, logID uuid.UUID, service string, classification *domain.ErrorClassification) error {
	if classification == nil {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO error_classifications (id, log_id, service, category, confidence, reasoning, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), logID, service, classification.Category, classification.Confidence, classification.Reasoning, time.Now().UTC(),
	)
	return err
}

// SaveExplanation stores a generated log explanation.
func (s *PostgresStore) SaveExplanation(ctx context.Context, logID uuid.UUID, service string, explanation *ai.LogExplanation) error {
	if explanation == nil {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO log_explanations (id, log_id, service, plain_english, technical_summary, severity_assessment, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), logID, service, explanation.PlainEnglish, explanation.TechnicalSummary, explanation.SeverityAssessment, time.Now().UTC(),
	)
	return err
}

// SaveRecommendations stores AI recommendations.
func (s *PostgresStore) SaveRecommendations(ctx context.Context, incidentID *uuid.UUID, recommendations []domain.Recommendation) error {
	for _, rec := range recommendations {
		_, err := s.pool.Exec(ctx, `
INSERT INTO recommendations (id, incident_id, type, description, code_snippet, priority, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), incidentID, rec.Type, rec.Description, rec.CodeSnippet, rec.Priority, time.Now().UTC(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
