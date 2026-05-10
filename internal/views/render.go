package views

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Render(c echo.Context, component templ.Component) error {
	var buffer bytes.Buffer

	if err := component.Render(c.Request().Context(), &buffer); err != nil {
		return err
	}

	return c.HTMLBlob(200, buffer.Bytes())
}
