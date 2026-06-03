package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) registerCloudNetworkRoutes(v1 *gin.RouterGroup) {
	v1.GET("/cloud/assets", h.ListCloudAssets)
	v1.GET("/cloud/topology", h.GetCloudAssetTopology)

	finops := v1.Group("/finops")
	{
		finops.GET("/costs", h.GetFinOpsCosts)
		finops.GET("/anomalies", h.ListFinOpsAnomalies)
		finops.GET("/carbon", h.GetFinOpsCarbon)
		h.registerFinOpsRoutes(finops)
	}

	network := v1.Group("/network")
	{
		network.GET("/flows", h.ListNetworkFlows)
		network.GET("/devices", h.ListNetworkDevices)
		network.GET("/topology", h.GetNetworkTopology)
		network.GET("/anomalies", h.ListNetworkAnomalies)
	}
}

func (h *Handler) ListCloudAssets(c *gin.Context) {
	if h.deps.Cloud != nil {
		writeSuccess(c, h.deps.Cloud.ListCloudAssets(c.Request.Context(), c.Query("provider")))
		return
	}
	writeSuccess(c, h.deps.Mem.ListCloudAssets(c.Query("provider")))
}

func (h *Handler) GetCloudAssetTopology(c *gin.Context) {
	if c.Query("live") == "true" && h.deps.Cloud != nil {
		assets := h.deps.Cloud.ListCloudAssets(c.Request.Context(), c.Query("provider"))
		writeSuccess(c, buildTopologyFromAssets(assets))
		return
	}
	writeSuccess(c, h.deps.Mem.CloudAssetTopology())
}

func (h *Handler) ListNetworkFlows(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if h.deps.Cloud != nil {
		writeSuccess(c, h.deps.Cloud.ListNetworkFlows(c.Request.Context(), limit))
		return
	}
	writeSuccess(c, h.deps.Mem.ListNetworkFlows(limit))
}

func (h *Handler) ListNetworkDevices(c *gin.Context) {
	if h.deps.Cloud != nil {
		writeSuccess(c, h.deps.Cloud.ListNetworkDevices(c.Request.Context()))
		return
	}
	writeSuccess(c, h.deps.Mem.ListNetworkDevices())
}

func (h *Handler) GetNetworkTopology(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.NetworkTopology())
}

func (h *Handler) ListNetworkAnomalies(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.NetworkAnomalies())
}

func buildTopologyFromAssets(assets []CloudAsset) CloudAssetTopology {
	nodes := make([]CloudAssetNode, 0, len(assets))
	edges := make([]CloudAssetEdge, 0)
	byType := map[string][]CloudAsset{}
	for _, a := range assets {
		nodes = append(nodes, CloudAssetNode{ID: a.ID, Label: a.Name, Type: a.Type, Provider: a.Provider, Health: a.Status})
		byType[a.Type] = append(byType[a.Type], a)
	}
	for _, a := range assets {
		if a.Type == "ec2" || a.Type == "vm" {
			for _, net := range byType["vpc"] {
				if net.Region == a.Region {
					edges = append(edges, CloudAssetEdge{Source: net.ID, Target: a.ID, Relation: "contains"})
				}
			}
		}
	}
	return CloudAssetTopology{Nodes: nodes, Edges: edges, At: time.Now().UTC()}
}
