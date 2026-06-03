package observability

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts all observability UI API routes under /api/v1.
func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	h.registerCoreRoutes(v1)
	h.registerDomainRoutes(v1)
}

func (h *Handler) registerCoreRoutes(v1 *gin.RouterGroup) {
	apm := v1.Group("/apm")
	{
		apm.GET("/traces/:traceId", h.GetTrace)
		apm.POST("/traces/search", h.SearchTraces)
		apm.GET("/flow", h.ServiceFlow)
		apm.GET("/services/:service/operations", h.ListOperations)
	}

	v1.POST("/query/unified", h.UnifiedQuery)
	v1.POST("/query/validate", h.ValidateUnifiedQuery)
	v1.GET("/query/functions", h.ListUnifiedQueryFunctions)
	h.registerNexQLRoutes(v1)

	collectors := v1.Group("/collectors")
	{
		collectors.GET("/fleet", h.ListCollectorFleet)
		collectors.POST("/fleet", h.CreateCollectorAgent)
		collectors.PUT("/fleet/:id", h.UpdateCollectorAgent)
		collectors.GET("/pipelines", h.ListCollectorPipelines)
		collectors.POST("/pipelines", h.CreateCollectorPipeline)
		collectors.PUT("/pipelines/:id", h.UpdateCollectorPipeline)
		collectors.POST("/pipelines/:id/validate", h.ValidateCollectorPipeline)
		collectors.GET("/autoinstrumentation", h.ListAutoInstrumentation)
	}

	v1.GET("/metrics/catalog", h.MetricCatalog)
	v1.GET("/metrics/query", h.QueryMetric)
	v1.GET("/dashboards", h.ListDashboards)
	v1.POST("/dashboards", h.CreateDashboard)
	v1.GET("/dashboards/:id", h.GetDashboard)
	v1.PUT("/dashboards/:id", h.UpdateDashboard)
	v1.DELETE("/dashboards/:id", h.DeleteDashboard)

	v1.GET("/topology", h.Topology)
	v1.GET("/zones", h.ListZones)

	v1.GET("/logs/metric-rules", h.ListLogMetricRules)
	v1.POST("/logs/metric-rules", h.CreateLogMetricRule)
	v1.GET("/logs/parsing-rules", h.ListLogParsingRules)
	v1.POST("/logs/parsing-rules", h.CreateLogParsingRule)
	h.registerLogRoutes(v1)

	v1.GET("/slos", h.ListSLOs)
	v1.POST("/slos", h.CreateSLO)
	v1.GET("/anomalies/entities", h.ListAnomalies)

	infra := v1.Group("/infra")
	{
		infra.GET("/hosts", h.ListHosts)
		infra.GET("/k8s/clusters", h.ListK8sClusters)
		infra.GET("/k8s/pods", h.ListK8sPods)
		h.registerServerlessRoutes(infra)
	}

	v1.GET("/databases", h.ListDatabases)
	v1.GET("/databases/:id/statements", h.DBStatements)
	v1.GET("/middleware/kafka/lag", h.KafkaLag)

	v1.GET("/rum/sessions", h.ListRUMSessions)
	v1.GET("/rum/sessions/:sessionId/replay", h.GetSessionReplay)
	v1.POST("/rum/beacon", h.IngestRUMBeacon)
	v1.POST("/rum/replay", h.IngestRUMReplay)
	v1.GET("/synthetic/monitors", h.ListSyntheticMonitors)
	v1.POST("/synthetic/monitors", h.CreateSyntheticMonitor)
	v1.GET("/synthetic/monitors/:id/runs", h.SyntheticRuns)

	v1.GET("/metrics/promql", h.QueryPromQL)

	v1.GET("/workflows", h.ListWorkflows)
	v1.POST("/workflows", h.CreateWorkflow)
	v1.PUT("/workflows/:id", h.UpdateWorkflow)
	v1.DELETE("/workflows/:id", h.DeleteWorkflow)
	v1.GET("/notebooks", h.ListNotebooks)
	v1.POST("/notebooks", h.CreateNotebook)

	v1.GET("/security/vulnerabilities", h.ListVulnerabilities)
	v1.POST("/security/vulnerabilities/batch", h.BatchImportVulnerabilities)
	v1.GET("/security/attacks", h.ListAttacks)
	v1.GET("/security/attacks/:id", h.GetAttack)
	v1.GET("/security/findings", h.ListSecurityFindings)
	v1.POST("/security/sca/import", h.ImportSecurityFindings)
	v1.POST("/security/images/import", h.ImportSecurityFindings)
	v1.GET("/security/cspm/posture", h.GetSecurityPosture)
	v1.POST("/security/cspm/import", h.ImportCSPMFindings)
	v1.POST("/security/siem/exports", h.ExportSecurityToSIEM)
	v1.GET("/integrations", h.ListIntegrations)
	v1.POST("/integrations/:id/connect", h.ConnectIntegration)

	v1.GET("/marketplace", h.ListMarketplace)
	v1.POST("/marketplace/:key/install", h.InstallExtension)
	v1.POST("/marketplace/:key/uninstall", h.UninstallExtension)

	alertPolicies := v1.Group("/alerts/policies")
	{
		alertPolicies.GET("", h.ListAlertPolicies)
		alertPolicies.POST("", h.CreateAlertPolicy)
		alertPolicies.PUT("/:id", h.UpdateAlertPolicy)
		alertPolicies.DELETE("/:id", h.DeleteAlertPolicy)
	}
	suppressions := v1.Group("/alerts/suppressions")
	{
		suppressions.GET("", h.ListAlertSuppressions)
		suppressions.POST("", h.CreateAlertSuppression)
		suppressions.DELETE("/:id", h.DeleteAlertSuppression)
	}

	admin := v1.Group("/admin")
	{
		admin.GET("/users", h.ListAdminUsers)
		admin.POST("/users", h.CreateAdminUser)
		admin.PATCH("/users/:id", h.UpdateAdminUser)
		admin.GET("/api-keys", h.ListAPIKeys)
		admin.POST("/api-keys", h.CreateAPIKey)
		admin.GET("/audit", h.ListAudit)
		admin.GET("/usage", h.Usage)
		admin.GET("/regions", h.GetMultiRegionStatus)
	}
}

func (h *Handler) registerDomainRoutes(v1 *gin.RouterGroup) {
	h.registerNPMExtendedRoutes(v1)
	h.registerAPMSamplingRoutes(v1)
	h.registerAIRoutes(v1)
	h.registerCloudNetworkRoutes(v1)
	h.registerExperienceRoutes(v1)
	h.registerGovernanceRoutes(v1)
	h.registerNFRRoutes(v1)
	h.registerSRSRoutes(v1)
	h.registerUnifiedRoutes(v1)
	h.RegisterDepthRoutes(v1)
	h.RegisterExtendedRoutes(v1)
	h.registerExtensionRoutes(v1)
}
