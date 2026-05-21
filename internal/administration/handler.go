package administration

import (
	adminviews "dorm-man/web/templates/administration"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func (h *Handler) writeError(c echo.Context, err error) error {
	category := classifyError(err)
	status := http.StatusInternalServerError
	switch category {
	case "validation_error", "state_transition_invalid":
		status = http.StatusBadRequest
	case "capacity_conflict", "concurrency_conflict":
		status = http.StatusConflict
	case "authorization_denied":
		status = http.StatusForbidden
	case "not_found":
		status = http.StatusNotFound
	}
	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) dashboardPage(c echo.Context) error {
	return renderComponent(c, adminviews.DashboardPage())
}

func (h *Handler) tenantsPage(c echo.Context) error {

}

func (h *Handler) tenantDetailPage(c echo.Context) error {

}

func (h *Handler) inventoryPage(c echo.Context) error {

}

func (h *Handler) inventoryDetailPage(c echo.Context) error {

}

func (h *Handler) createInventoryItem(c echo.Context) error {

}

func (h *Handler) updateInventoryItemStatus(c echo.Context) error {

}

func (h *Handler) deleteInventoryItem(c echo.Context) error {

}

func (h *Handler) jobsPage(c echo.Context) error {

}

func (h *Handler) jobDetailPage(c echo.Context) error {

}

func (h *Handler) createJob(c echo.Context) error {

}

func (h *Handler) updateJob(c echo.Context) error {

}

func (h *Handler) DeleteJob(c echo.Context) error {

}

func (h *Handler) publicationsPage(c echo.Context) error {

}

func (h *Handler) publicationDetailPage(c echo.Context) error {

}

func (h *Handler) createNews(c echo.Context) error {

}

func (h *Handler) archiveNews(c echo.Context) error {

}

func (h *Handler) createActivity(c echo.Context) error {

}

func (h *Handler) archiveActivity(c echo.Context) error {

}

func (h *Handler) createEvent(c echo.Context) error {

}

func (h *Handler) archiveEvent(c echo.Context) error {

}

func (h *Handler) housingPage(c echo.Context) error {

}

func (h *Handler) buildingDetail(c echo.Context) error {

}

func (h *Handler) flatDetail(c echo.Context) error {

}

func (h *Handler) sharedAreaDetail(c echo.Context) error {

}

func (h *Handler) roomDetail(c echo.Context) error {

}

func (h *Handler) assign(c echo.Context) error {

}

func (h *Handler) massAssignment(c echo.Context) error {

}

func (h *Handler) updateAssignment(c echo.Context) error {

}

func (h *Handler) deleteAssignment(c echo.Context) error {

}

func (h *Handler) auditPage(c echo.Context) error {

}
