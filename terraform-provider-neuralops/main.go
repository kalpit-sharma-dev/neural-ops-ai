package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/neuralops/terraform-provider-neuralops/neuralops"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: neuralops.New})
}
