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

func resourceAlertPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertPolicyCreate,
		ReadContext:   resourceAlertPolicyRead,
		DeleteContext: schema.NoopContext,
		Schema: map[string]*schema.Schema{
			"name":            {Type: schema.TypeString, Required: true, ForceNew: true},
			"service_pattern": {Type: schema.TypeString, Required: true, ForceNew: true},
			"severity":        {Type: schema.TypeString, Optional: true, Default: "P2", ForceNew: true},
			"enabled":         {Type: schema.TypeBool, Optional: true, Default: true, ForceNew: true},
			"policy_id":       {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceAlertPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(map[string]any{
		"name":           d.Get("name").(string),
		"servicePattern": d.Get("service_pattern").(string),
		"severity":       d.Get("severity").(string),
		"enabled":        d.Get("enabled").(bool),
		"routes":         []map[string]any{{"channel": "slack", "target": "#oncall", "after": "0m", "priority": 1}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/alerts/policies", bytes.NewReader(body))
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
		return diag.Errorf("create alert policy HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(parsed.Data.ID)
	_ = d.Set("policy_id", parsed.Data.ID)
	return resourceAlertPolicyRead(ctx, d, meta)
}

func resourceAlertPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	_ = d.Set("policy_id", d.Id())
	return nil
}
