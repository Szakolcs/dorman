package administration

import (
	"net/http"
	"strings"
	"time"

	"dorm-man/internal/pagination"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func pageParams(c echo.Context) pagination.Params {
	return pagination.Parse(c.QueryParam("page"), c.QueryParam("page_size"))
}

func pageParamsNamed(c echo.Context, pageKey, pageSizeKey string) pagination.Params {
	pageKey = defaultKey(pageKey, "page")
	pageSizeKey = defaultKey(pageSizeKey, "page_size")
	return pagination.Parse(c.QueryParam(pageKey), c.QueryParam(pageSizeKey))
}

func defaultKey(k, fallback string) string {
	if k == "" {
		return fallback
	}
	return k
}

func writeListJSON(c echo.Context, data any, meta pagination.Meta) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data":       data,
		"pagination": meta,
	})
}

func parseJobsFilter(c echo.Context) (assigneeStr, dateStr string, assignee *uuid.UUID, date *time.Time, err error) {
	assigneeStr = strings.TrimSpace(c.QueryParam("assignee_user_id"))
	dateStr = strings.TrimSpace(c.QueryParam("date"))
	if assigneeStr != "" {
		id, parseErr := uuid.Parse(assigneeStr)
		if parseErr != nil {
			return "", "", nil, nil, redirectView(c, "/administration/jobs", ErrValidation)
		}
		assignee = &id
	}
	if dateStr != "" {
		t, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			return "", "", nil, nil, redirectView(c, "/administration/jobs", ErrValidation)
		}
		date = &t
	}
	return assigneeStr, dateStr, assignee, date, nil
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
