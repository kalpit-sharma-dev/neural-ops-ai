package observability

import (
	"context"

	"github.com/neuralops/platform/internal/cloudinventory"
)

func (c *CloudCollectorService) listSDKAssets(ctx context.Context, provider string) ([]CloudAsset, bool) {
	if !c.liveEnabled() {
		return nil, false
	}
	reg := cloudinventory.NewRegistryFromEnv()
	if len(reg.Providers) == 0 {
		return nil, false
	}
	assets, err := reg.ListAll(ctx, provider)
	if err != nil || len(assets) == 0 {
		return nil, false
	}
	out := make([]CloudAsset, 0, len(assets))
	for _, a := range assets {
		out = append(out, CloudAsset{
			ID: a.ID, Provider: a.Provider, Type: a.Type, Name: a.Name,
			Region: a.Region, AccountID: a.AccountID, Status: a.Status,
			Tags: a.Tags, MonthlyUSD: a.MonthlyUSD, UpdatedAt: a.UpdatedAt,
		})
	}
	return out, true
}
