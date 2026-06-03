package neuralops

import (
	"context"
	"net/http"
	"os"
	"strings"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEURALOPS_API_URL", "http://localhost:8080/api/v1"),
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NEURALOPS_API_TOKEN", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"neuralops_export_job":             resourceExportJob(),
			"neuralops_alert_policy":           resourceAlertPolicy(),
			"neuralops_branding":               resourceBranding(),
			"neuralops_finops_budget":          resourceFinOpsBudget(),
			"neuralops_finops_allocation_rule": resourceFinOpsAllocationRule(),
			"neuralops_slo":                    resourceSLO(),
			"neuralops_abac_policy":            resourceABACPolicy(),
			"neuralops_msp_tenant":             resourceMSPTenant(),
		},
		ConfigureContextFunc: configure,
	}
}

type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func configure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	base := strings.TrimRight(d.Get("base_url").(string), "/")
	token := d.Get("token").(string)
	if token == "" {
		token = os.Getenv("NEURALOPS_API_TOKEN")
	}
	return &apiClient{baseURL: base, token: token, http: http.DefaultClient}, nil
}
