package neuralops

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMSPTenant() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMSPTenantCreate,
		ReadContext:   resourceMSPTenantRead,
		UpdateContext: resourceMSPTenantUpdate,
		DeleteContext: resourceMSPTenantDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the child tenant",
			},
			"slug": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL-safe tenant slug",
			},
			"plan": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "enterprise",
				Description: "Commercial plan tier",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data residency region",
			},
		},
	}
}

func mspTenantBody(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"name":   d.Get("name").(string),
		"slug":   d.Get("slug").(string),
		"plan":   d.Get("plan").(string),
		"region": d.Get("region").(string),
	}
}

func resourceMSPTenantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(mspTenantBody(d))
	raw, code, err := client.do(ctx, http.MethodPost, "/admin/msp/tenants", body)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("create msp tenant HTTP %d: %s", code, string(raw))
	}
	var parsed struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	if parsed.Data.ID == "" {
		return diag.Errorf("create msp tenant: missing id in response")
	}
	d.SetId(parsed.Data.ID)
	return resourceMSPTenantRead(ctx, d, meta)
}

func resourceMSPTenantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodGet, "/admin/msp/tenants", nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("read msp tenants HTTP %d: %s", code, string(raw))
	}
	var parsed struct {
		Data []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Slug   string `json:"slug"`
			Plan   string `json:"plan"`
			Region string `json:"region"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	for _, t := range parsed.Data {
		if t.ID == d.Id() {
			_ = d.Set("name", t.Name)
			_ = d.Set("slug", t.Slug)
			_ = d.Set("plan", t.Plan)
			_ = d.Set("region", t.Region)
			return nil
		}
	}
	return diag.Errorf("msp tenant %s not found", d.Id())
}

func resourceMSPTenantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceMSPTenantRead(ctx, d, meta)
}

func resourceMSPTenantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
