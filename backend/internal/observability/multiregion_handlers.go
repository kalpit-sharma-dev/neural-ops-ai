package observability

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegionPeer describes a peer region in active-active topology.
type RegionPeer struct {
	Region    string `json:"region"`
	Gateway   string `json:"gatewayUrl"`
	Status    string `json:"status"`
	IsPrimary bool   `json:"isPrimary"`
}

// MultiRegionStatus is tenant multi-region deployment view.
type MultiRegionStatus struct {
	Enabled       bool         `json:"enabled"`
	LocalRegion   string       `json:"localRegion"`
	Peers         []RegionPeer `json:"peers"`
	CrossBorderOK bool         `json:"crossBorderOk"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

func (h *Handler) GetMultiRegionStatus(c *gin.Context) {
	local := os.Getenv("NEURALOPS_REGION")
	if local == "" {
		local = "us-east-1"
	}
	enabled := strings.EqualFold(os.Getenv("NEURALOPS_MULTI_REGION"), "true")
	peers := parseRegionPeers(os.Getenv("NEURALOPS_PEER_REGIONS"))
	out := make([]RegionPeer, 0, len(peers)+1)
	out = append(out, RegionPeer{Region: local, Gateway: "/api/v1", Status: "healthy", IsPrimary: true})
	for _, p := range peers {
		if !strings.EqualFold(p, local) {
			out = append(out, RegionPeer{
				Region: p, Gateway: "https://" + p + ".observe.neuralops.ai/api/v1", Status: "healthy",
			})
		}
	}
	writeSuccess(c, MultiRegionStatus{
		Enabled: enabled, LocalRegion: local, Peers: out,
		CrossBorderOK: !enabled || len(peers) > 0, UpdatedAt: time.Now().UTC(),
	})
}

func parseRegionPeers(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
