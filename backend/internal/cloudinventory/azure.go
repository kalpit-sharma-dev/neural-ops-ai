package cloudinventory

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

// HasAzure reports whether Azure is configured.
func HasAzure() bool {
	return os.Getenv("AZURE_SUBSCRIPTION_ID") != ""
}

// AzureProvider lists VMs with SDK pager across resource groups.
type AzureProvider struct{}

func NewAzureProvider() *AzureProvider { return &AzureProvider{} }

func (p *AzureProvider) Name() string { return "azure" }

func (p *AzureProvider) ListAssets(ctx context.Context, opts ListOptions) ([]Asset, error) {
	subID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subID == "" {
		return nil, fmt.Errorf("AZURE_SUBSCRIPTION_ID required")
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client, err := armcompute.NewVirtualMachinesClient(subID, cred, nil)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var out []Asset
	pages := 0
	pager := client.NewListAllPager(nil)
	for pager.More() {
		if opts.MaxPages > 0 && pages >= opts.MaxPages {
			break
		}
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		pages++
		for _, vm := range page.Value {
			if vm == nil || vm.Name == nil {
				continue
			}
			region := ""
			if vm.Location != nil {
				region = *vm.Location
			}
			status := "unknown"
			if vm.Properties != nil && vm.Properties.ProvisioningState != nil {
				status = *vm.Properties.ProvisioningState
			}
			id := "azure-vm"
			if vm.ID != nil {
				id = *vm.ID
			}
			out = append(out, Asset{
				ID: id, Provider: "azure", Type: "vm",
				Name: *vm.Name, Region: region, Status: status,
				AccountID: subID, UpdatedAt: now,
				Tags: map[string]string{"source": "azure-sdk"},
			})
		}
	}
	return out, nil
}
