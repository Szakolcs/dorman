package doorman

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"dorm-man/internal/administration"
	dm "dorm-man/internal/models/doorman"
	doormanviews "dorm-man/web/templates/doorman"

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

func (h *Handler) actor(c echo.Context) (administration.Principal, error) {
	value := c.Request().Header.Get("X-Actor-User-ID")
	if value == "" {
		return administration.Principal{}, ErrUnauthorized
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return administration.Principal{}, ErrUnauthorized
	}
	return h.service.ResolvePrincipal(id)
}

func (h *Handler) staffActor(c echo.Context) (administration.Principal, error) {
	p, err := h.actor(c)
	if err != nil {
		return p, err
	}
	return p, h.service.RequireStaffDoormanPrincipal(p)
}

func (h *Handler) writeError(c echo.Context, err error) error {
	category := classifyDoormanError(err)
	status := http.StatusInternalServerError
	switch category {
	case "validation_error":
		status = http.StatusBadRequest
	case "outside_visit_window":
		status = http.StatusForbidden
	case "package_invalid_transition", "guest_visit_state_invalid", "loan_already_returned":
		status = http.StatusConflict
	case "loan_item_unavailable":
		status = http.StatusConflict
	case "authorization_denied":
		status = http.StatusForbidden
	case "not_found", "tenant_not_found":
		status = http.StatusNotFound
	}
	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}

func classifyDoormanError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrValidation):
		return "validation_error"
	case errors.Is(err, ErrOutsideVisitWindow):
		return "outside_visit_window"
	case errors.Is(err, ErrPackageInvalidTransition):
		return "package_invalid_transition"
	case errors.Is(err, ErrLoanItemUnavailable):
		return "loan_item_unavailable"
	case errors.Is(err, ErrUnauthorized):
		return "authorization_denied"
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case errors.Is(err, ErrTenantNotFound):
		return "tenant_not_found"
	case errors.Is(err, ErrGuestVisitStateTransition):
		return "guest_visit_state_invalid"
	case errors.Is(err, ErrLoanAlreadyReturned):
		return "loan_already_returned"
	default:
		return "internal_error"
	}
}

func queryPositiveInt(c echo.Context, name string, def int) int {
	s := c.QueryParam(name)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func (h *Handler) registerPackage(c echo.Context) error {
	p, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body PackageRegisterInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	pkg, err := h.service.RegisterPackage(p, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": pkg})
}

func (h *Handler) listPackages(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	filter := PackageListFilter{
		Status: c.QueryParam("status"),
		Limit:  queryPositiveInt(c, "limit", 100),
		Offset: queryPositiveInt(c, "offset", 0),
	}
	if tid := c.QueryParam("tenant_id"); tid != "" {
		id, err := uuid.Parse(tid)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.TenantID = &id
	}
	list, err := h.service.ListPackages(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
}

func (h *Handler) getPackage(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	pkg, err := h.service.GetPackage(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": pkg})
}

func (h *Handler) transitionPackage(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body struct {
		To dm.PackageStatus `json:"to"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	pkg, err := h.service.TransitionPackage(principal, id, body.To)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": pkg})
}

func (h *Handler) pickupPackage(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	pkg, err := h.service.ConfirmPackagePickup(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": pkg})
}

func (h *Handler) notifyPackage(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body PackageNotifyInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	n, err := h.service.NotifyPackageTenant(principal, id, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": n})
}

func (h *Handler) createGuestVisit(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body GuestVisitCreateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	v, err := h.service.RegisterGuestVisit(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": v})
}

func (h *Handler) listGuestVisits(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	filter := GuestVisitListFilter{
		Limit:  queryPositiveInt(c, "limit", 100),
		Offset: queryPositiveInt(c, "offset", 0),
	}
	if on := c.QueryParam("on"); on != "" {
		t, err := time.Parse("2006-01-02", on)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.OnDate = &t
	}
	list, err := h.service.ListGuestVisits(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
}

func (h *Handler) getGuestVisit(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	v, err := h.service.GetGuestVisit(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": v})
}

func (h *Handler) guestCheckIn(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	v, err := h.service.GuestCheckIn(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": v})
}

func (h *Handler) guestCheckOut(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	v, err := h.service.GuestCheckOut(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": v})
}

func (h *Handler) issueTenantEntryToken(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body TenantEntryTokenCreateInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	t, err := h.service.IssueTenantEntryToken(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": t})
}

func (h *Handler) validateEntryQR(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body QRValidateBody
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	out, err := h.service.ValidateEntryByQR(principal, body.PublicRef)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": out})
}

func (h *Handler) validateEntryManual(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body ManualAccessBody
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	out, err := h.service.ValidateEntryManual(principal, body.TenantID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": out})
}

func (h *Handler) listAccessEvents(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	filter := AccessEventListFilter{
		Limit:  queryPositiveInt(c, "limit", 100),
		Offset: queryPositiveInt(c, "offset", 0),
	}
	if tid := c.QueryParam("tenant_id"); tid != "" {
		id, err := uuid.Parse(tid)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.TenantID = &id
	}
	if from := c.QueryParam("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.From = &t
	}
	if to := c.QueryParam("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.To = &t
	}
	if oc := c.QueryParam("outcome"); oc != "" {
		filter.Outcome = dm.AccessEventOutcome(oc)
	}
	list, err := h.service.ListAccessEvents(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
}

func (h *Handler) checkoutLoan(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body CheckoutLoanBody
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	loan, err := h.service.CheckoutItem(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": loan})
}

func (h *Handler) returnLoan(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	loan, err := h.service.ReturnItem(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": loan})
}

func (h *Handler) listLoans(c echo.Context) error {
	if _, err := h.staffActor(c); err != nil {
		return h.writeError(c, err)
	}
	filter := ItemLoanListFilter{
		OpenOnly:    c.QueryParam("open_only") == "true",
		OverdueOnly: c.QueryParam("overdue_only") == "true",
		Limit:       queryPositiveInt(c, "limit", 100),
		Offset:      queryPositiveInt(c, "offset", 0),
	}
	if tid := c.QueryParam("tenant_id"); tid != "" {
		id, err := uuid.Parse(tid)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		filter.TenantID = &id
	}
	list, err := h.service.ListItemLoans(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
}

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
