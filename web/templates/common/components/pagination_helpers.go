package components

import (
	"fmt"
	"net/url"
	"strconv"

	"dorm-man/internal/pagination"
)

func fmtPage(n int) string {
	return strconv.Itoa(n)
}

func fmtTotal(n int64) string {
	return strconv.FormatInt(n, 10)
}

func pageLink(basePath string, page, pageSize int, preserve map[string]string) string {
	return pageLinkNamed(basePath, page, pageSize, "page", "page_size", preserve)
}

func pageLinkNamed(basePath string, page, pageSize int, pageKey, pageSizeKey string, preserve map[string]string) string {
	v := url.Values{}
	for k, val := range preserve {
		if val != "" {
			v.Set(k, val)
		}
	}
	return pagination.PageQueryNamed(basePath, page, pageSize, pageKey, pageSizeKey, v)
}

// PreserveFromMap builds url.Values from a string map for pagination links.
func PreserveFromMap(m map[string]string) map[string]string {
	return m
}

// FormatPageLabel is used by templates via fmtPage.
var _ = fmt.Sprintf
