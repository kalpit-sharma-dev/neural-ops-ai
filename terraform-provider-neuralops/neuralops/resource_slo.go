package neuralops

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSLO() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSLOCreate,
		ReadContext:   resourceSLORead,
		DeleteContext: schema.NoopContext,
		Schema: map[string]*schema.Schema{
			"name":         {Type: schema.TypeString, Required: true, ForceNew: true},
			"service":      {Type: schema.TypeString, Required: true, ForceNew: true},
			"sli_query":    {Type: schema.TypeString, Required: true, ForceNew: true},
			"target":       {Type: schema.TypeFloat, Required: true, ForceNew: true},
			"window_days":  {Type: schema.TypeInt, Optional: true, Default: 30, ForceNew: true},
			"slo_id":       {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceSLOCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	body, _ := json.Marshal(map[string]any{
		"name":       d.Get("name").(string),
		"service":    d.Get("service").(string),
		"sliQuery":   d.Get("sli_query").(string),
		"target":     d.Get("target").(float64),
		"windowDays": d.Get("window_days").(int),
	})
	raw, code, err := client.do(ctx, http.MethodPost, "/slos", body)
	if err != nil {
		return diag.FromErr(err)
	}
	if code >= 300 {
		return diag.Errorf("create slo HTTP %d: %s", code, string(raw))
	}
	id, err := parseDataID(raw)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	_ = d.Set("slo_id", id)
	return resourceSLORead(ctx, d, meta)
}

func resourceSLORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	_ = d.Set("slo_id", d.Id())
	return nil
}
