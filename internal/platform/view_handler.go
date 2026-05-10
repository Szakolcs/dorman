package platform

import (
	platformviews "dorm-man/web/templates/platform"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) loginPage(c echo.Context) error {
	return renderComponent(c, platformviews.LoginPage())
}

func (h *Handler) aboutPage(c echo.Context) error {
	return renderComponent(c, platformviews.AboutPage())
}
