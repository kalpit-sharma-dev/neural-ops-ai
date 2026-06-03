package cloudinventory

import (
	"context"
	"time"
)

// Asset is a normalized cloud resource for the observability API.
type Asset struct {
	ID         string
	Provider   string
	Type       string
	Name       string
	Region     string
	AccountID  string
	Status     string
	Tags       map[string]string
	MonthlyUSD float64
	UpdatedAt  time.Time
}

// ListOptions controls paginated inventory scans.
type ListOptions struct {
	Provider string
	MaxPages int
	PageSize int
}

// Provider lists cloud assets with SDK pagination.
type Provider interface {
	Name() string
	ListAssets(ctx context.Context, opts ListOptions) ([]Asset, error)
}
