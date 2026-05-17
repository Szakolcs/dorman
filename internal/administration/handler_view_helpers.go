package administration

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func isHtmx(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func redirectView(c echo.Context, path string, err error) error {
	if err == nil {
		return c.Redirect(http.StatusSeeOther, path)
	}
	if isHtmx(c) {
		return c.String(http.StatusBadRequest, err.Error())
	}
	q := url.Values{}
	q.Set("error", err.Error())
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return c.Redirect(http.StatusSeeOther, path+sep+q.Encode())
}
