package pagination

import (
	"math"
	"net/url"
	"strconv"
)

const (
	DefaultPageSize = 25
	MaxPageSize     = 100
)

// Params holds 1-based page index and page size.
type Params struct {
	Page     int
	PageSize int
}

// Parse builds params from query string values.
// Unpaged returns params that disable SQL LIMIT (for dropdowns and full scans).
func Unpaged() Params {
	return Params{Page: 1, PageSize: 0}
}

func Parse(pageStr, pageSizeStr string) Params {
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Params{Page: page, PageSize: pageSize}
}

func (p Params) Limit() int {
	return p.PageSize
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Meta describes a paginated result set for views and JSON APIs.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasPrev    bool  `json:"has_prev"`
	HasNext    bool  `json:"has_next"`
}

func NewMeta(p Params, total int64) Meta {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(p.PageSize)))
	}
	if totalPages == 0 {
		totalPages = 1
	}
	return Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasPrev:    p.Page > 1,
		HasNext:    int64(p.Page) < int64(totalPages) && total > int64(p.Offset()+p.PageSize),
	}
}

// Slice paginates an in-memory slice (used when filtering happens outside SQL).
func Slice[T any](items []T, p Params) ([]T, Meta) {
	total := int64(len(items))
	meta := NewMeta(p, total)
	if total == 0 {
		return []T{}, meta
	}
	start := p.Offset()
	if start >= len(items) {
		return []T{}, meta
	}
	end := start + p.Limit()
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], meta
}

// AppendPageQuery adds page and page_size to an existing query string.
func AppendPageQuery(basePath string, p Params, extra url.Values) string {
	return AppendPageQueryNamed(basePath, p, "page", "page_size", extra)
}

// AppendPageQueryNamed adds named page query keys.
func AppendPageQueryNamed(basePath string, p Params, pageKey, pageSizeKey string, extra url.Values) string {
	v := url.Values{}
	for k, vals := range extra {
		for _, val := range vals {
			v.Add(k, val)
		}
	}
	if pageKey == "" {
		pageKey = "page"
	}
	if pageSizeKey == "" {
		pageSizeKey = "page_size"
	}
	v.Set(pageKey, strconv.Itoa(p.Page))
	v.Set(pageSizeKey, strconv.Itoa(p.PageSize))
	if len(v) == 0 {
		return basePath
	}
	return basePath + "?" + v.Encode()
}

// PageQuery returns query string for a specific page preserving other params.
func PageQuery(basePath string, page int, pageSize int, extra url.Values) string {
	return PageQueryNamed(basePath, page, pageSize, "page", "page_size", extra)
}

// PageQueryNamed is like PageQuery but uses custom query parameter names.
func PageQueryNamed(basePath string, page int, pageSize int, pageKey, pageSizeKey string, extra url.Values) string {
	return AppendPageQueryNamed(basePath, Params{Page: page, PageSize: pageSize}, pageKey, pageSizeKey, extra)
}
