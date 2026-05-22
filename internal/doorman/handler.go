package doorman

import (
	"dorm-man/internal/models"
	doormanviews "dorm-man/web/templates/doorman/pages"
	dv "dorm-man/web/templates/doorman/partials"
	"errors"
	"fmt"
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
	form := accessFilterFormValues(c)
	filter := GuestFilter{}
	if c.Request().Header.Get("HX-Request") == "true" {
		var err error
		filter, err = guestFilterFromForm(c)
		if err != nil {
			return h.writeError(c, err)
		}
	}
	guests, err := h.service.ListGuests(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	tenants, err := h.service.GetTenants()
	if err != nil {
		return h.writeError(c, err)
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderComponent(c, dv.GuestTable(guests, tenants))
	}
	return renderComponent(c, doormanviews.GuestsPage(
		guests,
		tenants,
		form.FromDate,
		form.ToDate,
		form.FromHour,
		form.ToHour,
		form.Status,
	))
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
	form := accessFilterFormValues(c)
	filter := TenantAccessFilter{}
	if c.Request().Header.Get("HX-Request") == "true" {
		var err error
		filter, err = tenantAccessFilterFromForm(c)
		if err != nil {
			return h.writeError(c, err)
		}
	}
	entries, err := h.service.ListTenantAccess(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderComponent(c, dv.TenantAccessTable(entries))
	}
	return renderComponent(c, doormanviews.TenantAccessPage(
		entries,
		form.FromDate,
		form.ToDate,
		form.FromHour,
		form.ToHour,
		form.Status,
	))
}

func (h *Handler) createTenantAccess(c echo.Context) error {
	req := TenantAccessRequest{
		TenantID:     uuid.MustParse(c.QueryParam("id")),
		AccessStatus: models.AccessStatus(c.QueryParam("direction")),
	}
	fmt.Println("req", req)
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

func accessFilterFormValues(c echo.Context) AccessFilterForm {
	now := time.Now()
	form := AccessFilterForm{
		FromDate: now.Format(time.DateOnly),
		FromHour: now.Format("15:04"),
		Status:   "checked_in",
	}

	if c.Request().Header.Get("HX-Request") != "true" {
		return form
	}

	if from := c.QueryParam("from"); from != "" {
		form.FromDate = from
	}
	if to := c.QueryParam("to"); to != "" {
		form.ToDate = to
	}
	if fromHour := c.QueryParam("from_hour"); fromHour != "" {
		form.FromHour = fromHour
	}
	if toHour := c.QueryParam("to_hour"); toHour != "" {
		form.ToHour = toHour
	}
	if status := c.QueryParam("status"); status != "" {
		form.Status = status
	}

	return form
}

func guestFilterFromForm(c echo.Context) (GuestFilter, error) {
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
	if err := applyAccessTimeFilter(c, &filter.From, &filter.To); err != nil {
		return GuestFilter{}, err
	}
	return filter, nil
}

func tenantAccessFilterFromForm(c echo.Context) (TenantAccessFilter, error) {
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
	if err := applyAccessTimeFilter(c, &filter.From, &filter.To); err != nil {
		return TenantAccessFilter{}, err
	}
	return filter, nil
}

func applyAccessTimeFilter(c echo.Context, from, to *time.Time) error {
	if fromDate := c.QueryParam("from"); fromDate != "" {
		fromHour := c.QueryParam("from_hour")
		if fromHour == "" {
			fromHour = "00:00"
		}
		t, err := combineFilterDateTime(fromDate, fromHour, false)
		if err != nil {
			return ErrValidation
		}
		*from = t
	}
	if toDate := c.QueryParam("to"); toDate != "" {
		toHour := c.QueryParam("to_hour")
		t, err := combineFilterDateTime(toDate, toHour, toHour == "")
		if err != nil {
			return ErrValidation
		}
		*to = t
	}
	return nil
}

func combineFilterDateTime(dateStr, timeStr string, endOfDay bool) (time.Time, error) {
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	if timeStr == "" {
		if endOfDay {
			return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.Local), nil
		}
		return date, nil
	}
	hour, minute, err := parseFilterClock(timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		hour,
		minute,
		0,
		0,
		time.Local,
	), nil
}

func parseFilterClock(timeStr string) (int, int, error) {
	for _, layout := range []string{"15:04", "15:04:05"} {
		clock, err := time.Parse(layout, timeStr)
		if err == nil {
			return clock.Hour(), clock.Minute(), nil
		}
	}
	return 0, 0, fmt.Errorf("invalid time %q", timeStr)
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
