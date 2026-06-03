package finops

import (
	"strings"

	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/gateway/auth"
)

// MaskAccountID redacts cloud account identifiers for non-privileged roles (REQ §8).
func MaskAccountID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if len(id) <= 4 {
		return "****"
	}
	return id[:2] + strings.Repeat("*", len(id)-4) + id[len(id)-2:]
}

// ShouldMaskAccounts returns true when the caller lacks FinOps admin visibility.
func ShouldMaskAccounts(role auth.Role) bool {
	switch role {
	case auth.RoleAdmin, auth.RoleSRE:
		return false
	default:
		return true
	}
}

// MaskLineItems applies account masking to cost line items when required.
func MaskLineItems(items []domain.CostLineItem, mask bool) []domain.CostLineItem {
	if !mask {
		return items
	}
	out := make([]domain.CostLineItem, len(items))
	copy(out, items)
	for i := range out {
		out[i].AccountID = MaskAccountID(out[i].AccountID)
	}
	return out
}
