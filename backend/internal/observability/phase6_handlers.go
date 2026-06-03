package observability

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) registerExperienceRoutes(v1 *gin.RouterGroup) {
	v1.GET("/rum/funnels", h.ListRUMFunnels)
	v1.POST("/rum/funnels", h.CreateRUMFunnel)

	syn := v1.Group("/synthetic")
	{
		syn.GET("/browser-tests", h.ListSyntheticBrowserTests)
		syn.POST("/browser-tests", h.CreateSyntheticBrowserTest)
		syn.GET("/mobile-tests", h.ListSyntheticMobileTests)
		syn.POST("/mobile-tests", h.CreateSyntheticMobileTest)
		syn.GET("/private-locations", h.ListSyntheticPrivateLocations)
		syn.POST("/private-locations", h.RegisterSyntheticPrivateLocation)
	}

	biz := v1.Group("/business")
	{
		biz.GET("/kpi-packs", h.ListBusinessKPIPacks)
		biz.POST("/kpi-packs/:id/enable", h.EnableBusinessKPIPack)
	}
}

func (h *Handler) ListRUMFunnels(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListRUMFunnels())
}

type createRUMFunnelRequest struct {
	Name  string `json:"name"`
	Steps []struct {
		Name  string `json:"name"`
		Event string `json:"event"`
	} `json:"steps"`
}

func (h *Handler) CreateRUMFunnel(c *gin.Context) {
	var req createRUMFunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateFunnelName(req.Name); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Steps) == 0 {
		writeError(c, http.StatusBadRequest, "at least one funnel step is required")
		return
	}
	steps := make([]RUMFunnelStep, 0, len(req.Steps))
	for _, st := range req.Steps {
		steps = append(steps, RUMFunnelStep{Name: st.Name, Event: st.Event})
	}
	writeSuccess(c, h.deps.Mem.SaveRUMFunnel(RUMFunnel{Name: req.Name, Steps: steps}))
}

type createBrowserTestRequest struct {
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Script    string   `json:"script"`
	Locations []string `json:"locations"`
	Enabled   *bool    `json:"enabled"`
}

func (h *Handler) ListSyntheticBrowserTests(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListSyntheticBrowserTests())
}

func (h *Handler) ListSyntheticMobileTests(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListSyntheticMobileTests())
}

func (h *Handler) ListSyntheticPrivateLocations(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListSyntheticPrivateLocations())
}

func (h *Handler) CreateSyntheticBrowserTest(c *gin.Context) {
	var req createBrowserTestRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || req.URL == "" {
		writeError(c, http.StatusBadRequest, "name and url are required")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	writeSuccess(c, h.deps.Mem.SaveSyntheticBrowserTest(SyntheticBrowserTest{
		Name: req.Name, URL: req.URL, Script: req.Script, Locations: req.Locations, Enabled: enabled,
	}))
}

type createMobileTestRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
	BundleID string `json:"bundleId"`
	Script   string `json:"script"`
	Enabled  *bool  `json:"enabled"`
}

func (h *Handler) CreateSyntheticMobileTest(c *gin.Context) {
	var req createMobileTestRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		writeError(c, http.StatusBadRequest, "name is required")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	writeSuccess(c, h.deps.Mem.SaveSyntheticMobileTest(SyntheticMobileTest{
		Name: req.Name, Platform: req.Platform, BundleID: req.BundleID, Script: req.Script, Enabled: enabled,
	}))
}

type registerPrivateLocationRequest struct {
	Name         string `json:"name"`
	Region       string `json:"region"`
	AgentVersion string `json:"agentVersion"`
}

func (h *Handler) RegisterSyntheticPrivateLocation(c *gin.Context) {
	var req registerPrivateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		writeError(c, http.StatusBadRequest, "name is required")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveSyntheticPrivateLocation(SyntheticPrivateLocation{
		Name: req.Name, Region: req.Region, AgentVersion: req.AgentVersion,
	}))
}

func (h *Handler) ListBusinessKPIPacks(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListBusinessKPIPacks())
}

func (h *Handler) EnableBusinessKPIPack(c *gin.Context) {
	pack, err := h.deps.Mem.EnableBusinessKPIPack(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, pack)
}
