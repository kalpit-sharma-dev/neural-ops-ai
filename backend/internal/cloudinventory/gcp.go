package cloudinventory

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// HasGCP reports whether GCP is configured.
func HasGCP() bool {
	return os.Getenv("GCP_PROJECT_ID") != "" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != ""
}

// GCPProvider lists Compute Engine instances with aggregated list pagination.
type GCPProvider struct{}

func NewGCPProvider() *GCPProvider { return &GCPProvider{} }

func (p *GCPProvider) Name() string { return "gcp" }

func (p *GCPProvider) ListAssets(ctx context.Context, opts ListOptions) ([]Asset, error) {
	project := os.Getenv("GCP_PROJECT_ID")
	if project == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID required")
	}
	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	svc, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var out []Asset
	pages := 0
	req := svc.Instances.AggregatedList(project).Context(ctx)
	err = req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		if opts.MaxPages > 0 && pages >= opts.MaxPages {
			return nil
		}
		pages++
		for zone, scoped := range page.Items {
			for _, inst := range scoped.Instances {
				out = append(out, Asset{
					ID:       fmt.Sprintf("gcp-%d", inst.Id),
					Provider: "gcp",
					Type:     "gce",
					Name:     inst.Name,
					Region:   zone,
					Status:   inst.Status,
					UpdatedAt: now,
					Tags:     map[string]string{"source": "gcp-sdk", "project": project},
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
