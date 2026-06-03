package observability

import (
	"time"

	"github.com/gin-gonic/gin"
)

// NPMWirelessLink is a wireless access point link metric.
type NPMWirelessLink struct {
	ID         string    `json:"id"`
	APName     string    `json:"apName"`
	ClientMAC  string    `json:"clientMac"`
	SSID       string    `json:"ssid"`
	RSSI       float64   `json:"rssiDbm"`
	Throughput float64   `json:"throughputMbps"`
	PacketLoss float64   `json:"packetLossPct"`
	Timestamp  time.Time `json:"timestamp"`
}

// NPMSDWANTunnel is SD-WAN path health.
type NPMSDWANTunnel struct {
	ID        string    `json:"id"`
	Site      string    `json:"site"`
	Provider  string    `json:"provider"`
	LatencyMs float64   `json:"latencyMs"`
	JitterMs  float64   `json:"jitterMs"`
	LossPct   float64   `json:"lossPct"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// NetFlowRecord is an extended flow export record.
type NetFlowRecord struct {
	ID          string    `json:"id"`
	Exporter    string    `json:"exporter"`
	SrcIP       string    `json:"srcIp"`
	DstIP       string    `json:"dstIp"`
	Protocol    string    `json:"protocol"`
	Bytes       int64     `json:"bytes"`
	Packets     int64     `json:"packets"`
	Application string    `json:"application"`
	Timestamp   time.Time `json:"timestamp"`
}

func (h *Handler) registerNPMExtendedRoutes(v1 *gin.RouterGroup) {
	network := v1.Group("/network")
	{
		network.GET("/flows/netflow", h.ListNetFlowRecords)
		network.POST("/flows/netflow/ingest", h.IngestNetFlowRecords)
		network.GET("/sdwan/tunnels", h.ListSDWANTunnels)
		network.GET("/wireless/links", h.ListWirelessLinks)
	}
}

func (h *Handler) IngestNetFlowRecords(c *gin.Context) {
	var body struct {
		Records []NetFlowRecord `json:"records"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Records) == 0 {
		writeError(c, 400, "records array required")
		return
	}
	writeSuccess(c, h.deps.Mem.IngestNetFlowRecords(body.Records))
}

func (h *Handler) ListNetFlowRecords(c *gin.Context) {
	if stored := h.deps.Mem.ListNetFlowRecordsStored(50); len(stored) > 0 {
		writeSuccess(c, stored)
		return
	}
	now := time.Now().UTC()
	writeSuccess(c, []NetFlowRecord{
		{ID: "nf-1", Exporter: "edge-1", SrcIP: "10.0.1.5", DstIP: "10.0.2.8", Protocol: "TCP", Bytes: 4_500_000, Packets: 3200, Application: "https", Timestamp: now},
		{ID: "nf-2", Exporter: "edge-2", SrcIP: "10.0.3.2", DstIP: "10.0.4.1", Protocol: "UDP", Bytes: 890_000, Packets: 1200, Application: "dns", Timestamp: now.Add(-2 * time.Minute)},
	})
}

func (h *Handler) ListSDWANTunnels(c *gin.Context) {
	now := time.Now().UTC()
	writeSuccess(c, []NPMSDWANTunnel{
		{ID: "sd-1", Site: "nyc-hq", Provider: "MPLS", LatencyMs: 12, JitterMs: 1.2, LossPct: 0.01, Status: "up", Timestamp: now},
		{ID: "sd-2", Site: "lon-dc", Provider: "internet-vpn", LatencyMs: 48, JitterMs: 4.5, LossPct: 0.08, Status: "degraded", Timestamp: now},
	})
}

func (h *Handler) ListWirelessLinks(c *gin.Context) {
	now := time.Now().UTC()
	writeSuccess(c, []NPMWirelessLink{
		{ID: "wl-1", APName: "ap-floor-3", ClientMAC: "aa:bb:cc:dd:ee:01", SSID: "corp-wifi", RSSI: -58, Throughput: 120, PacketLoss: 0.2, Timestamp: now},
	})
}
