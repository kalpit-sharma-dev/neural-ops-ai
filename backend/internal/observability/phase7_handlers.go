package observability

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) registerGovernanceRoutes(v1 *gin.RouterGroup) {
	admin := v1.Group("/admin")
	{
		admin.GET("/abac-policies", h.GetABACPolicies)
		admin.PUT("/abac-policies", h.PutABACPolicies)
		admin.GET("/data-residency", h.GetDataResidency)
		admin.PUT("/data-residency", h.PutDataResidency)
		admin.GET("/branding", h.GetBranding)
		admin.PUT("/branding", h.PutBranding)
		admin.GET("/msp/tenants", h.ListMSPTenants)
		admin.POST("/msp/tenants", h.CreateMSPTenant)
	}

	exports := v1.Group("/exports")
	{
		exports.POST("/warehouse", h.ExportWarehouse)
		exports.POST("/bi", h.ExportBI)
		exports.POST("/events", h.ExportEvents)
	}
}

func (h *Handler) GetABACPolicies(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.GetABACPolicy())
}

func (h *Handler) PutABACPolicies(c *gin.Context) {
	var body ABACPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid abac policy body")
		return
	}
	p := h.deps.Mem.PutABACPolicy(body)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		_ = h.deps.SRS.SaveABAC(c.Request.Context(), tenantID(c), p)
	}
	writeSuccess(c, p)
}

func (h *Handler) GetDataResidency(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.GetDataResidency())
}

func (h *Handler) PutDataResidency(c *gin.Context) {
	var body DataResidencyPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid data residency body")
		return
	}
	p := h.deps.Mem.PutDataResidency(body)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		_ = h.deps.SRS.SaveResidency(c.Request.Context(), tenantID(c), p)
	}
	writeSuccess(c, p)
}

func (h *Handler) GetBranding(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.GetBranding())
}

func (h *Handler) PutBranding(c *gin.Context) {
	var body BrandingTheme
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid branding body")
		return
	}
	writeSuccess(c, h.deps.Mem.PutBranding(body))
}

func (h *Handler) ListMSPTenants(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListMSPTenants())
}

func (h *Handler) CreateMSPTenant(c *gin.Context) {
	var body MSPTenant
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		writeError(c, http.StatusBadRequest, "name is required")
		return
	}
	writeSuccess(c, h.deps.Mem.CreateMSPTenant(body))
}

type exportRequest struct {
	Destination string            `json:"destination"`
	Format      string            `json:"format,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

func (h *Handler) ExportWarehouse(c *gin.Context) {
	h.createExport(c, "warehouse")
}

func (h *Handler) ExportBI(c *gin.Context) {
	h.createExport(c, "bi")
}

func (h *Handler) ExportEvents(c *gin.Context) {
	h.createExport(c, "events")
}

func (h *Handler) createExport(c *gin.Context, exportType string) {
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Destination == "" {
		writeError(c, http.StatusBadRequest, "destination is required")
		return
	}
	writeSuccess(c, h.deps.Mem.CreateExportJob(exportType, req.Destination))
}
