package doorman

import (
	crosscutting "dorm-man/internal/models/cross-cutting"
	models "dorm-man/internal/models/doorman"
	doormanviews "dorm-man/web/templates/doorman/pages"
	"errors"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) dashboardPage(c echo.Context) error {
	entries, err := h.service.ListTenantAccess(TenantAccessFilter{})
	if err != nil {
		return h.writeError(c, err)
	}
	if len(entries) > 5 {
		entries = entries[:5]
	}
	tenants, err := h.service.GetTenants()
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, doormanviews.DashboardPage(entries, tenants))
}

func (h *Handler) listGuest(c echo.Context) error {
	filter := GuestFilter{}
	tenantID := c.QueryParam("tenant_id")
	if tenantID != "" {
		id, err := uuid.Parse(tenantID)
		if err == nil {
			filter.TenantID = id
		}
	}
	status := c.QueryParam("status")
	if status != "" {
		filter.Status = models.AccessStatus(status)
	}
	from := c.QueryParam("from")
	if from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err == nil {
			filter.From = t
		}
	}
	to := c.QueryParam("to")
	if to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err == nil {
			filter.To = t
		}
	}
	c.Response().Header().Set("HX-Redirect", "/doorman/guests")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) registerGuest(
	c echo.Context,
) error {
	hostTenantID, err := uuid.Parse(
		c.FormValue("host_tenant_id"),
	)
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	req := GuestRegisterRequest{
		HostTenant: &crosscutting.User{
			BaseModel: crosscutting.BaseModel{
				ID: hostTenantID,
			},
		},
		GuestName: c.FormValue("guest_name"),
		IDNotes:   c.FormValue("id_notes"),
	}
	_, err = h.service.RegisterGuest(req)
	if err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/doorman/guests")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) deleteGuest(c echo.Context) error {
	guestID, err := uuid.Parse(
		c.Param("guestID"),
	)
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	err = h.service.DeleteGuest(guestID)
	if err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/doorman/guests")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) listTenantAccess(c echo.Context) error {
	filter := TenantAccessFilter{}
	tenantID := c.QueryParam("tenant_id")
	if tenantID != "" {
		id, err := uuid.Parse(tenantID)
		if err == nil {
			filter.TenantID = id
		}
	}
	status := c.QueryParam("status")
	if status != "" {
		filter.Status = models.AccessStatus(status)
	}
	from := c.QueryParam("from")
	if from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err == nil {
			filter.From = t
		}
	}
	to := c.QueryParam("to")
	if to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err == nil {
			filter.To = t
		}
	}
	entries, err := h.service.ListTenantAccess(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, doormanviews.TenantAccessPage(entries))
}

func (h *Handler) createTenantAccess(c echo.Context) error {
	id, err := uuid.Parse(c.QueryParam("id"))
	direction := c.QueryParam("direction")
	req := TenantAccessRequest{
		TenantID:     id,
		AccessStatus: models.AccessStatus(direction),
	}
	_, err = h.service.RegisterTenantAccess(req)
	if err != nil {
		return h.writeError(c, err)
	}
	entries, err := h.service.ListTenantAccess(
		TenantAccessFilter{},
	)
	if err != nil {
		return h.writeError(c, err)
	}
	if len(entries) > 5 {
		entries = entries[:5]
	}
	c.Response().Header().Set("HX-Redirect", "/doorman")
	return c.NoContent(http.StatusOK)
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) writeError(c echo.Context, err error) error {
	category := "internal_error"
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrValidation):
		category = "validation_error"
		status = http.StatusBadRequest
	case errors.Is(err, ErrOutsideVisitWindow):
		category = "outside_visit_window"
		status = http.StatusForbidden
	case errors.Is(err, ErrPackageInvalidTransition):
		category = "package_invalid_transition"
		status = http.StatusConflict
	case errors.Is(err, ErrGuestVisitStateTransition):
		category = "guest_visit_state_invalid"
		status = http.StatusConflict
	case errors.Is(err, ErrLoanAlreadyReturned):
		category = "loan_already_returned"
		status = http.StatusConflict
	case errors.Is(err, ErrLoanItemUnavailable):
		category = "loan_item_unavailable"
		status = http.StatusConflict
	case errors.Is(err, ErrUnauthorized):
		category = "authorization_denied"
		status = http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		category = "not_found"
		status = http.StatusNotFound
	case errors.Is(err, ErrTenantNotFound):
		category = "tenant_not_found"
		status = http.StatusNotFound
	}
	return c.JSON(
		status,
		map[string]any{
			"error": map[string]string{
				"category": category,
				"message":  err.Error(),
			},
		},
	)
}
