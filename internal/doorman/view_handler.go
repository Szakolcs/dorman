package doorman

import (
	doormanviews "dorm-man/web/templates/doorman"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) dashboardPage(c echo.Context) error {
	return renderComponent(c, doormanviews.DashboardPage())
}

func (h *Handler) packagesPage(c echo.Context) error {
	return renderComponent(c, doormanviews.PackagesPage())
}

func (h *Handler) guestsPage(c echo.Context) error {
	return renderComponent(c, doormanviews.GuestsPage())
}

func (h *Handler) accessPage(c echo.Context) error {
	return renderComponent(c, doormanviews.AccessPage())
}

func (h *Handler) lendingPage(c echo.Context) error {
	return renderComponent(c, doormanviews.LendingPage())
}
