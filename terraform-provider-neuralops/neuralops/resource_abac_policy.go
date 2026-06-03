package neuralops

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceABACPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceABACPolicyCreate,
		ReadContext:   resourceABACPolicyRead,
		UpdateContext: resourceABACPolicyUpdate,
		Schema: map[string]*schema.Schema{
			"enabled": {Type: schema.TypeBool, Required: true},
			"rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":         {Type: schema.TypeString, Required: true},
						"effect":     {Type: schema.TypeString, Required: true},
						"action":     {Type: schema.TypeString, Required: true},
						"resource":   {Type: schema.TypeString, Required: true},
						"condition":  {Type: schema.TypeString, Optional: true},
					},
				},
			},
		},
	}
}

func abacBody(d *schema.ResourceData) map[string]any {
	rules := make([]map[string]any, 0)
	for _, raw := range d.Get("rules").([]interface{}) {
		m := raw.(map[string]interface{})
		cond := ""
		if v, ok := m["condition"]; ok && v != nil {
			cond = v.(string)
		}
		rules = append(rules, map[string]any{
			"id":        m["id"].(string),
			"effect":    m["effect"].(string),
			"action":    m["action"].(string),
			"resource":  m["resource"].(string),
			"condition": cond,
		})
	}
	return map[string]any{"enabled": d.Get("enabled").(bool), "rules": rules}
}

func resourceABACPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("tenant-abac-policy")
	return resourceABACPolicyUpdate(ctx, d, meta)
}

func resourceABACPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodGet, "/admin/abac-policies", nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("read abac HTTP %d: %s", code, string(raw))
	}
	var parsed struct {
		Data struct {
			Enabled bool `json:"enabled"`
			Rules   []struct {
				ID        string `json:"id"`
				Effect    string `json:"effect"`
				Action    string `json:"action"`
				Resource  string `json:"resource"`
				Condition string `json:"condition"`
			} `json:"rules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("enabled", parsed.Data.Enabled)
	return nil
}

func resourceABACPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(abacBody(d))
	raw, code, err := client.do(ctx, http.MethodPut, "/admin/abac-policies", body)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("update abac HTTP %d: %s", code, string(raw))
	}
	_ = raw
	return resourceABACPolicyRead(ctx, d, meta)
}
