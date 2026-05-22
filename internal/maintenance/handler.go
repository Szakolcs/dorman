package maintenance

import (
	"errors"
	"net/http"
	"time"

	"dorm-man/internal/models"
	maintenanceviews "dorm-man/web/templates/maintenance"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
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
	tickets, err := h.service.ListTickets(TicketFilter{})
	if err != nil {
		return h.writeError(c, err)
	}
	if len(tickets) > 10 {
		tickets = tickets[:10]
	}
	return renderComponent(c, maintenanceviews.DashboardPage(tickets))
}

func (h *Handler) ticketsPage(c echo.Context) error {
	filter, err := h.ticketFilter(c)
	if err != nil {
		return h.writeError(c, err)
	}
	tickets, err := h.service.ListTickets(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, maintenanceviews.TicketsPage(tickets))
}

func (h *Handler) ticketDetailPage(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	ticket, err := h.service.GetTicketByID(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, maintenanceviews.TicketDetailPage(ticket))
}

func (h *Handler) updateTicket(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}

	req := UpdateTicketRequest{
		TicketID:    id,
		Status:      models.Status(c.FormValue("status")),
		Category:    models.Category(c.FormValue("category")),
		Severity:    models.Severity(c.FormValue("severity")),
		Impact:      models.Impact(c.FormValue("impact")),
		Description: c.FormValue("description"),
		Note:        c.FormValue("note"),
	}
	if _, err := h.service.UpdateTicket(actorID, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/maintenance/tickets/"+id.String())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) deleteTicket(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.DeleteTicket(actorID, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/maintenance/tickets")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) tenantDashboardPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	tickets, err := h.service.ListTicketsByUser(actorID, TicketFilter{})
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, maintenanceviews.TenantDashboardPage(tickets))
}

func (h *Handler) tenantTicketsPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	filter, err := h.ticketFilter(c)
	if err != nil {
		return h.writeError(c, err)
	}
	tickets, err := h.service.ListTicketsByUser(actorID, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, maintenanceviews.TenantTicketsPage(tickets))
}

func (h *Handler) createTicket(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}

	req := CreateTicketRequest{
		Category:    models.Category(c.FormValue("category")),
		Severity:    models.Severity(c.FormValue("severity")),
		Impact:      models.Impact(c.FormValue("impact")),
		Description: c.FormValue("description"),
	}

	ticket, err := h.service.CreateTicket(actorID, req)
	if err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/tenant/tickets/"+ticket.ID.String())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) tenantTicketDetailPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	ticket, err := h.service.GetTicketByID(id)
	if err != nil {
		return h.writeError(c, err)
	}
	if ticket.CreatedByUserID != actorID {
		return h.writeError(c, ErrUnauthorized)
	}
	return renderComponent(c, maintenanceviews.TenantTicketDetailPage(ticket))
}

func (h *Handler) tenantDeleteTicket(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.DeleteOwnTicket(actorID, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/tenant")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) ticketFilter(c echo.Context) (TicketFilter, error) {
	filter := TicketFilter{}
	if status := c.QueryParam("status"); status != "" {
		filter.Status = models.Status(status)
	}
	if category := c.QueryParam("category"); category != "" {
		filter.Category = models.Category(category)
	}
	if severity := c.QueryParam("severity"); severity != "" {
		filter.Severity = models.Severity(severity)
	}
	if from := c.QueryParam("from"); from != "" {
		t, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.From = t
	}
	if to := c.QueryParam("to"); to != "" {
		t, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return TicketFilter{}, ErrValidation
		}
		filter.To = t
	}
	return filter, nil
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func actorID(c echo.Context) (uuid.UUID, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return uuid.Nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}
	raw, ok := claims["id"]
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}
	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, ErrUnauthorized
	}
}

func paramUUID(c echo.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, ErrValidation
	}
	return id, nil
}

func (h *Handler) writeError(c echo.Context, err error) error {
	category := "internal_error"
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrValidation):
		category = "validation_error"
		status = http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		category = "authorization_denied"
		status = http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		category = "not_found"
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
