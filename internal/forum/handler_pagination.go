package forum

import (
	"net/http"

	"dorm-man/internal/pagination"

	"github.com/labstack/echo/v4"
)

func pageParams(c echo.Context) pagination.Params {
	return pagination.Parse(c.QueryParam("page"), c.QueryParam("page_size"))
}

func pageParamsNamed(c echo.Context, pageKey, pageSizeKey string) pagination.Params {
	if pageKey == "" {
		pageKey = "page"
	}
	if pageSizeKey == "" {
		pageSizeKey = "page_size"
	}
	return pagination.Parse(c.QueryParam(pageKey), c.QueryParam(pageSizeKey))
}

func writeListJSON(c echo.Context, data any, meta pagination.Meta) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data":       data,
		"pagination": meta,
	})
}

func queryPreserve(c echo.Context, omit ...string) map[string]string {
	skip := map[string]bool{"page": true, "page_size": true}
	for _, k := range omit {
		skip[k] = true
	}
	out := make(map[string]string)
	for k, vals := range c.QueryParams() {
		if skip[k] || len(vals) == 0 {
			continue
		}
		out[k] = vals[0]
	}
	return out
}
