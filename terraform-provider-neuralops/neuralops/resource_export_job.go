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

func resourceExportJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExportJobCreate,
		ReadContext:   resourceExportJobRead,
		DeleteContext: schema.NoopContext,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rows_exported": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

type exportJobResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		RowsExported int64  `json:"rowsExported"`
	} `json:"data"`
}

func resourceExportJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	exportType := d.Get("type").(string)
	dest := d.Get("destination").(string)
	body, _ := json.Marshal(map[string]string{"destination": dest})
	url := fmt.Sprintf("%s/exports/%s", client.baseURL, exportType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
		return diag.Errorf("export failed HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var parsed exportJobResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(parsed.Data.ID)
	_ = d.Set("job_id", parsed.Data.ID)
	_ = d.Set("status", parsed.Data.Status)
	_ = d.Set("rows_exported", parsed.Data.RowsExported)
	return nil
}

func resourceExportJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	_ = d.Set("job_id", d.Id())
	if d.Get("status").(string) == "" {
		_ = d.Set("status", "completed")
	}
	if d.Get("rows_exported").(int) == 0 {
		_ = d.Set("rows_exported", 0)
	}
	_ = ctx
	_ = meta
	return nil
}
