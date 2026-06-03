package neuralops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFinOpsBudget() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFinOpsBudgetCreate,
		ReadContext:   resourceFinOpsBudgetRead,
		UpdateContext: resourceFinOpsBudgetUpdate,
		DeleteContext: resourceFinOpsBudgetDelete,
		Schema: map[string]*schema.Schema{
			"name":        {Type: schema.TypeString, Required: true},
			"scope_type":  {Type: schema.TypeString, Required: true},
			"scope_value": {Type: schema.TypeString, Required: true},
			"period":      {Type: schema.TypeString, Optional: true, Default: "monthly"},
			"amount_usd":  {Type: schema.TypeFloat, Required: true},
			"thresholds":  {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeInt}, Default: []int{50, 80, 100}},
		},
	}
}

func finOpsBudgetBody(d *schema.ResourceData) map[string]any {
	thresholds := []int{50, 80, 100}
	if raw, ok := d.GetOk("thresholds"); ok {
		thresholds = make([]int, 0, len(raw.([]interface{})))
		for _, v := range raw.([]interface{}) {
			thresholds = append(thresholds, v.(int))
		}
	}
	return map[string]any{
		"name":       d.Get("name").(string),
		"scopeType":  d.Get("scope_type").(string),
		"scopeValue": d.Get("scope_value").(string),
		"period":     d.Get("period").(string),
		"amountUsd":  d.Get("amount_usd").(float64),
		"thresholds": thresholds,
	}
}

func resourceFinOpsBudgetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(finOpsBudgetBody(d))
	raw, code, err := client.do(ctx, http.MethodPost, "/finops/budgets", body)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("create finops budget HTTP %d: %s", code, string(raw))
	}
	id, err := parseDataID(raw)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return resourceFinOpsBudgetRead(ctx, d, meta)
}

func resourceFinOpsBudgetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodGet, "/finops/budgets", nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("read finops budgets HTTP %d: %s", code, string(raw))
	}
	var parsed struct {
		Data []struct {
			ID         string  `json:"id"`
			Name       string  `json:"name"`
			ScopeType  string  `json:"scopeType"`
			ScopeValue string  `json:"scopeValue"`
			Period     string  `json:"period"`
			AmountUSD  float64 `json:"amountUsd"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	for _, b := range parsed.Data {
		if b.ID == d.Id() {
			_ = d.Set("name", b.Name)
			_ = d.Set("scope_type", b.ScopeType)
			_ = d.Set("scope_value", b.ScopeValue)
			_ = d.Set("period", b.Period)
			_ = d.Set("amount_usd", b.AmountUSD)
			return nil
		}
	}
	return diag.Errorf("budget %s not found", d.Id())
}

func resourceFinOpsBudgetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body := finOpsBudgetBody(d)
	body["id"] = d.Id()
	raw, err := json.Marshal(body)
	if err != nil {
		return diag.FromErr(err)
	}
	respRaw, code, err := client.do(ctx, http.MethodPut, "/finops/budgets/"+d.Id(), raw)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("update finops budget HTTP %d: %s", code, string(respRaw))
	}
	return resourceFinOpsBudgetRead(ctx, d, meta)
}

func resourceFinOpsBudgetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodDelete, "/finops/budgets/"+d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("delete finops budget HTTP %d: %s", code, string(raw))
	}
	d.SetId("")
	return nil
}

func resourceFinOpsAllocationRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFinOpsAllocationRuleCreate,
		ReadContext:   resourceFinOpsAllocationRuleRead,
		UpdateContext: resourceFinOpsAllocationRuleUpdate,
		DeleteContext: resourceFinOpsAllocationRuleDelete,
		Schema: map[string]*schema.Schema{
			"name":      {Type: schema.TypeString, Required: true},
			"dimension": {Type: schema.TypeString, Required: true},
			"tag_key":   {Type: schema.TypeString, Required: true},
			"tag_value": {Type: schema.TypeString, Optional: true, Default: ""},
			"priority":  {Type: schema.TypeInt, Optional: true, Default: 100},
			"enabled":   {Type: schema.TypeBool, Optional: true, Default: true},
		},
	}
}

func finOpsRuleBody(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"name":      d.Get("name").(string),
		"dimension": d.Get("dimension").(string),
		"tagKey":    d.Get("tag_key").(string),
		"tagValue":  d.Get("tag_value").(string),
		"priority":  d.Get("priority").(int),
		"enabled":   d.Get("enabled").(bool),
	}
}

func resourceFinOpsAllocationRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(finOpsRuleBody(d))
	raw, code, err := client.do(ctx, http.MethodPost, "/finops/allocation/rules", body)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("create finops allocation rule HTTP %d: %s", code, string(raw))
	}
	id, err := parseDataID(raw)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return resourceFinOpsAllocationRuleRead(ctx, d, meta)
}

func resourceFinOpsAllocationRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodGet, "/finops/allocation/rules", nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("read finops rules HTTP %d: %s", code, string(raw))
	}
	var parsed struct {
		Data []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Dimension string `json:"dimension"`
			TagKey    string `json:"tagKey"`
			TagValue  string `json:"tagValue"`
			Priority  int    `json:"priority"`
			Enabled   bool   `json:"enabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	for _, r := range parsed.Data {
		if r.ID == d.Id() {
			_ = d.Set("name", r.Name)
			_ = d.Set("dimension", r.Dimension)
			_ = d.Set("tag_key", r.TagKey)
			_ = d.Set("tag_value", r.TagValue)
			_ = d.Set("priority", r.Priority)
			_ = d.Set("enabled", r.Enabled)
			return nil
		}
	}
	return diag.Errorf("allocation rule %s not found", d.Id())
}

func resourceFinOpsAllocationRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body := finOpsRuleBody(d)
	body["id"] = d.Id()
	raw, err := json.Marshal(body)
	if err != nil {
		return diag.FromErr(err)
	}
	respRaw, code, err := client.do(ctx, http.MethodPut, "/finops/allocation/rules/"+d.Id(), raw)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("update finops rule HTTP %d: %s", code, string(respRaw))
	}
	return resourceFinOpsAllocationRuleRead(ctx, d, meta)
}

func resourceFinOpsAllocationRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	raw, code, err := client.do(ctx, http.MethodDelete, "/finops/allocation/rules/"+d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("delete finops rule HTTP %d: %s", code, string(raw))
	}
	d.SetId("")
	return nil
}

func parseDataID(raw []byte) (string, error) {
	var parsed struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if parsed.Data.ID == "" {
		return "", fmt.Errorf("missing id in response")
	}
	return parsed.Data.ID, nil
}

func (c *apiClient) do(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}
