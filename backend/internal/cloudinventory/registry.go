package cloudinventory

import (
	"context"
	"os"
	"strconv"
)

// Registry runs all configured cloud providers.
type Registry struct {
	Providers []Provider
}

// NewRegistryFromEnv registers AWS, GCP, and Azure providers when credentials exist.
func NewRegistryFromEnv() *Registry {
	r := &Registry{}
	if HasAWS() {
		r.Providers = append(r.Providers, NewAWSProvider())
	}
	if HasGCP() {
		r.Providers = append(r.Providers, NewGCPProvider())
	}
	if HasAzure() {
		r.Providers = append(r.Providers, NewAzureProvider())
	}
	return r
}

// ListAll aggregates assets from every configured provider with pagination limits.
func (r *Registry) ListAll(ctx context.Context, providerFilter string) ([]Asset, error) {
	opts := ListOptions{
		Provider: providerFilter,
		MaxPages: envInt("NEURALOPS_CLOUD_MAX_PAGES", 500),
		PageSize: envInt("NEURALOPS_CLOUD_PAGE_SIZE", 100),
	}
	var all []Asset
	for _, p := range r.Providers {
		if providerFilter != "" && p.Name() != providerFilter {
			continue
		}
		assets, err := p.ListAssets(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, assets...)
	}
	return all, nil
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
