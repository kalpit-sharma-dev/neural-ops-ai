package observability

import (
	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/pkg/nexagent/ebpf/validate"
)

// GetCollectorEBPFMatrix returns host eBPF fleet compatibility (GAP-COLL-001).
func (h *Handler) GetCollectorEBPFMatrix(c *gin.Context) {
	_ = h
	writeSuccess(c, validate.ProbeHost())
}
