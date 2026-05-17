package chat

import (
	"net/http"

	"dorm-man/internal/pagination"

	"github.com/labstack/echo/v4"
)

func pageParams(c echo.Context) pagination.Params {
	return pagination.Parse(c.QueryParam("page"), c.QueryParam("page_size"))
}

func writeListJSON(c echo.Context, data any, meta pagination.Meta) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data":       data,
		"pagination": meta,
	})
}
