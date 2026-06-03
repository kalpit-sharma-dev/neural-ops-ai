package neuralops

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBranding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBrandingUpsert,
		UpdateContext: resourceBrandingUpsert,
		ReadContext:   resourceBrandingRead,
		Schema: map[string]*schema.Schema{
			"product_name": {Type: schema.TypeString, Required: true},
			"logo_url":     {Type: schema.TypeString, Optional: true},
			"primary_hex":  {Type: schema.TypeString, Optional: true},
		},
	}
}

func resourceBrandingUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(map[string]string{
		"productName": d.Get("product_name").(string),
		"logoUrl":     d.Get("logo_url").(string),
		"primaryHex":  d.Get("primary_hex").(string),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, client.baseURL+"/admin/branding", bytes.NewReader(body))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if client.token != "" {
		req.Header.Set("Authorization", "Bearer "+client.token)
	}
	resp, err := client.http.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return diag.Errorf("branding upsert HTTP %d: %s", resp.StatusCode, string(raw))
	}
	d.SetId("branding")
	return resourceBrandingRead(ctx, d, meta)
}

func resourceBrandingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		d.SetId("branding")
	}
	return nil
}
