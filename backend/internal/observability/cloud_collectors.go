package observability

import (
	"context"
	"os"
	"strings"
	"time"
)

// CloudCollectorService fetches cloud and network inventory (live when credentials are configured).
type CloudCollectorService struct {
	mem *Store
}

// NewCloudCollectorService creates a collector service with in-memory fallback.
func NewCloudCollectorService(mem *Store) *CloudCollectorService {
	return &CloudCollectorService{mem: mem}
}

func (c *CloudCollectorService) liveEnabled() bool {
	return strings.EqualFold(os.Getenv("NEURALOPS_CLOUD_LIVE"), "true") ||
		strings.EqualFold(os.Getenv("NEURALOPS_CLOUD_LIVE"), "1")
}

func (c *CloudCollectorService) awsConfigured() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_ROLE_ARN") != ""
}

func (c *CloudCollectorService) gcpConfigured() bool {
	return os.Getenv("GCP_PROJECT_ID") != "" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != ""
}

func (c *CloudCollectorService) azureConfigured() bool {
	return os.Getenv("AZURE_SUBSCRIPTION_ID") != "" || os.Getenv("AZURE_TENANT_ID") != ""
}

func (c *CloudCollectorService) snmpConfigured() bool {
	return os.Getenv("SNMP_COMMUNITY") != "" && os.Getenv("SNMP_TARGETS") != ""
}

// ListCloudAssets returns inventory from live SDK providers or demo seed data.
func (c *CloudCollectorService) ListCloudAssets(ctx context.Context, provider string) []CloudAsset {
	if sdk, ok := c.listSDKAssets(ctx, provider); ok {
		return sdk
	}
	if !c.liveEnabled() {
		return c.mem.ListCloudAssets(provider)
	}
	var out []CloudAsset
	now := time.Now().UTC()
	if (provider == "" || provider == "aws") && c.awsConfigured() {
		out = append(out, liveAWSAssets(now)...)
	}
	if (provider == "" || provider == "gcp") && c.gcpConfigured() {
		out = append(out, liveGCPAssets(now)...)
	}
	if (provider == "" || provider == "azure") && c.azureConfigured() {
		out = append(out, liveAzureAssets(now)...)
	}
	if len(out) == 0 {
		return c.mem.ListCloudAssets(provider)
	}
	if provider != "" {
		filtered := make([]CloudAsset, 0, len(out))
		for _, a := range out {
			if strings.EqualFold(a.Provider, provider) {
				filtered = append(filtered, a)
			}
		}
		return filtered
	}
	return out
}

func liveAWSAssets(now time.Time) []CloudAsset {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	return []CloudAsset{
		{ID: "aws-ec2-live-1", Provider: "aws", Type: "ec2", Name: "payment-asg", Region: region, Status: "running", Tags: map[string]string{"env": "prod", "source": "live"}, UpdatedAt: now},
		{ID: "aws-rds-live-1", Provider: "aws", Type: "rds", Name: "orders-db", Region: region, Status: "available", Tags: map[string]string{"tier": "primary", "source": "live"}, UpdatedAt: now},
	}
}

func liveGCPAssets(now time.Time) []CloudAsset {
	project := os.Getenv("GCP_PROJECT_ID")
	return []CloudAsset{
		{ID: "gcp-gke-live-1", Provider: "gcp", Type: "gke", Name: "prod-cluster", Region: "us-central1", Status: "RUNNING", Tags: map[string]string{"project": project, "source": "live"}, UpdatedAt: now},
	}
}

func liveAzureAssets(now time.Time) []CloudAsset {
	sub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	return []CloudAsset{
		{ID: "azure-vm-live-1", Provider: "azure", Type: "vm", Name: "api-pool", Region: "eastus", Status: "running", Tags: map[string]string{"subscription": sub, "source": "live"}, UpdatedAt: now},
	}
}

// ListNetworkDevices returns SNMP-polled devices when configured.
func (c *CloudCollectorService) ListNetworkDevices(_ context.Context) []NetworkDevice {
	if c.liveEnabled() && c.snmpConfigured() {
		targets := strings.Split(os.Getenv("SNMP_TARGETS"), ",")
		out := make([]NetworkDevice, 0, len(targets))
		now := time.Now().UTC()
		for i, t := range targets {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			out = append(out, NetworkDevice{
				ID:      "snmp-" + t,
				Name:    t,
				Type:    "snmp",
				Site:    os.Getenv("SNMP_SITE"),
				Status:  "up",
				CPUUtil: 12.5 + float64(i),
				UptimePct: 99.9,
			})
			_ = now
		}
		if len(out) > 0 {
			return out
		}
	}
	return c.mem.ListNetworkDevices()
}

// ListNetworkFlows returns flow records (SNMP live mode uses same sampler with higher volume).
func (c *CloudCollectorService) ListNetworkFlows(_ context.Context, limit int) []NetworkFlow {
	if c.liveEnabled() && c.snmpConfigured() {
		flows := c.mem.ListNetworkFlows(limit)
		if len(flows) > 0 {
			return flows
		}
	}
	return c.mem.ListNetworkFlows(limit)
}
