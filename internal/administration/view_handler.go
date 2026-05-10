package administration

import (
	adminviews "dorm-man/web/templates/administration"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) dashboardPage(c echo.Context) error {
	return renderComponent(c, adminviews.DashboardPage())
}

func (h *Handler) tenantsPage(c echo.Context) error {
	return renderComponent(c, adminviews.TenantsPage())
}

func (h *Handler) roomsPage(c echo.Context) error {
	return renderComponent(c, adminviews.RoomsPage())
}

func (h *Handler) inventoryPage(c echo.Context) error {
	return renderComponent(c, adminviews.InventoryPage())
}

func (h *Handler) maintenancePage(c echo.Context) error {
	return renderComponent(c, adminviews.MaintenancePage())
}

func (h *Handler) jobsPage(c echo.Context) error {
	return renderComponent(c, adminviews.JobsPage())
}

func (h *Handler) publicationsPage(c echo.Context) error {
	return renderComponent(c, adminviews.PublicationsPage())
}

func (h *Handler) auditPage(c echo.Context) error {
	return renderComponent(c, adminviews.AuditPage())
}
