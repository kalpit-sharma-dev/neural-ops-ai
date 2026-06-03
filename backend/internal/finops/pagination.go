package finops

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// PageMeta describes list pagination (REQ §5).
type PageMeta struct {
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Total int    `json:"total"`
	Sort  string `json:"sort,omitempty"`
}

// PagedResult wraps a paginated list response.
type PagedResult[T any] struct {
	Items []T    `json:"items"`
	Page  PageMeta `json:"page"`
}

// ParsePage reads page/size/sort query params with sane defaults.
func ParsePage(c *gin.Context) (page, size int, sortKey string, paged bool) {
	page = 1
	size = 50
	if c.Query("page") != "" || c.Query("size") != "" {
		paged = true
	}
	if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.DefaultQuery("size", "50")); err == nil && v > 0 {
		size = v
	}
	if size > 200 {
		size = 200
	}
	sortKey = strings.TrimSpace(c.Query("sort"))
	return page, size, sortKey, paged
}

// Paginate slices a list after optional sort by a string key function.
func Paginate[T any](items []T, page, size int, sortKey string, keyFn func(T) string) ([]T, PageMeta) {
	if sortKey != "" && keyFn != nil {
		desc := strings.HasPrefix(sortKey, "-")
		field := strings.TrimPrefix(sortKey, "-")
		if field != "" {
			sort.SliceStable(items, func(i, j int) bool {
				a, b := keyFn(items[i]), keyFn(items[j])
				if desc {
					return a > b
				}
				return a < b
			})
		}
	}
	total := len(items)
	start := (page - 1) * size
	if start >= total {
		return []T{}, PageMeta{Page: page, Size: size, Total: total, Sort: sortKey}
	}
	end := start + size
	if end > total {
		end = total
	}
	out := make([]T, end-start)
	copy(out, items[start:end])
	return out, PageMeta{Page: page, Size: size, Total: total, Sort: sortKey}
}
