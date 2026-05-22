package doorman

import (
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
	if _, err := guestFilter(c); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/doorman/guests")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) registerGuest(c echo.Context) error {
	var req GuestRegisterRequest
	if err := c.Bind(&req); err != nil {
		return h.writeError(c, ErrValidation)
	}
	if _, err := h.service.RegisterGuest(req); err != nil {
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
	filter, err := tenantAccessFilter(c)
	if err != nil {
		return h.writeError(c, err)
	}
	entries, err := h.service.ListTenantAccess(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, doormanviews.TenantAccessPage(entries))
}

func (h *Handler) createTenantAccess(c echo.Context) error {
	var req TenantAccessRequest
	if err := c.Bind(&req); err != nil {
		return h.writeError(c, ErrValidation)
	}
	_, err := h.service.RegisterTenantAccess(req)
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

func guestFilter(c echo.Context) (GuestFilter, error) {
	filter := GuestFilter{}
	if tenantID := c.QueryParam("tenant_id"); tenantID != "" {
		id, err := uuid.Parse(tenantID)
		if err != nil {
			return GuestFilter{}, ErrValidation
		}
		filter.TenantID = id
	}
	if status := c.QueryParam("status"); status != "" {
		filter.Status = models.AccessStatus(status)
	}
	if from := c.QueryParam("from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return GuestFilter{}, ErrValidation
		}
		filter.From = t
	}
	if to := c.QueryParam("to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return GuestFilter{}, ErrValidation
		}
		filter.To = t
	}
	return filter, nil
}

func tenantAccessFilter(c echo.Context) (TenantAccessFilter, error) {
	filter := TenantAccessFilter{}
	if tenantID := c.QueryParam("tenant_id"); tenantID != "" {
		id, err := uuid.Parse(tenantID)
		if err != nil {
			return TenantAccessFilter{}, ErrValidation
		}
		filter.TenantID = id
	}
	if status := c.QueryParam("status"); status != "" {
		filter.Status = models.AccessStatus(status)
	}
	if from := c.QueryParam("from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return TenantAccessFilter{}, ErrValidation
		}
		filter.From = t
	}
	if to := c.QueryParam("to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return TenantAccessFilter{}, ErrValidation
		}
		filter.To = t
	}
	return filter, nil
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
