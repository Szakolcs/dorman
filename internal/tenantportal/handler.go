package tenantportal

import (
	"errors"
	"net/http"
	"strings"

	"dorm-man/internal/administration"
	"dorm-man/internal/chat"
	"dorm-man/internal/maintenance"
	"dorm-man/internal/models"
	tenantviews "dorm-man/web/templates/tenant"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	admin       *administration.Service
	maintenance *maintenance.Service
	chat        *chat.Service
}

func NewHandler(admin *administration.Service, maint *maintenance.Service, chatSvc *chat.Service) *Handler {
	return &Handler{
		admin:       admin,
		maintenance: maint,
		chat:        chatSvc,
	}
}

func (h *Handler) publicationsPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	published := models.PublicationStatePublished
	filter := administration.PublicationFilter{
		Search: c.QueryParam("search"),
		Kind:   administration.PublicationKind(c.QueryParam("kind")),
		State:  &published,
	}
	pubs, err := h.admin.ListPublications(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	items, err := h.publicationCardItems(actorID, pubs)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, tenantviews.PublicationsPage(items, string(filter.Kind)))
}

func (h *Handler) publicationDetailPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	detail, err := h.admin.GetPublication(id)
	if err != nil {
		return h.writeError(c, err)
	}
	if !isPublished(detail) {
		return h.writeError(c, administration.ErrNotFound)
	}
	switch detail.Kind {
	case administration.PublicationKindNews:
		return renderComponent(c, tenantviews.NewsDetailPage(*detail.News))
	case administration.PublicationKindActivity:
		booked, err := h.admin.HasActivityBooking(actorID, id)
		if err != nil {
			return h.writeError(c, err)
		}
		return renderComponent(c, tenantviews.ActivityDetailPage(*detail.Activity, booked))
	case administration.PublicationKindEvent:
		intent, err := h.admin.GetEventAttendanceIntent(actorID, id)
		if err != nil {
			return h.writeError(c, err)
		}
		return renderComponent(c, tenantviews.EventDetailPage(*detail.Event, string(intent)))
	default:
		return h.writeError(c, administration.ErrNotFound)
	}
}

func (h *Handler) bookActivity(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.admin.BookActivity(actorID, id); err != nil {
		return h.writeError(c, err)
	}
	redirect := c.QueryParam("return")
	if redirect == "" || !strings.HasPrefix(redirect, "/tenant") {
		redirect = "/tenant/publications/" + id.String()
	}
	c.Response().Header().Set("HX-Redirect", redirect)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) cancelActivityBooking(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.admin.CancelActivityBooking(actorID, id); err != nil {
		return h.writeError(c, err)
	}
	redirect := c.QueryParam("return")
	if redirect == "" || !strings.HasPrefix(redirect, "/tenant") {
		redirect = "/tenant/publications/" + id.String()
	}
	c.Response().Header().Set("HX-Redirect", redirect)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) setEventIntent(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	intent := models.EventAttendanceIntent(c.FormValue("intent"))
	if err := h.admin.SetEventAttendanceIntent(actorID, id, intent); err != nil {
		return h.writeError(c, err)
	}
	redirect := c.QueryParam("return")
	if redirect == "" || !strings.HasPrefix(redirect, "/tenant") {
		redirect = "/tenant/publications/" + id.String()
	}
	c.Response().Header().Set("HX-Redirect", redirect)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) publicationCardItems(actorID uuid.UUID, pubs []administration.PublicationListItem) ([]tenantviews.PublicationCardItem, error) {
	items := make([]tenantviews.PublicationCardItem, len(pubs))
	for i, pub := range pubs {
		item := tenantviews.PublicationCardItem{PublicationListItem: pub}
		if pub.Kind == administration.PublicationKindActivity {
			booked, err := h.admin.HasActivityBooking(actorID, pub.ID)
			if err != nil {
				return nil, err
			}
			item.BookedByViewer = booked
		}
		if pub.Kind == administration.PublicationKindEvent {
			intent, err := h.admin.GetEventAttendanceIntent(actorID, pub.ID)
			if err != nil {
				return nil, err
			}
			item.ViewerEventIntent = string(intent)
		}
		items[i] = item
	}
	return items, nil
}

func isPublished(detail administration.PublicationDetail) bool {
	switch detail.Kind {
	case administration.PublicationKindNews:
		return detail.News != nil && detail.News.State == models.PublicationStatePublished
	case administration.PublicationKindActivity:
		return detail.Activity != nil && detail.Activity.State == models.PublicationStatePublished
	case administration.PublicationKindEvent:
		return detail.Event != nil && detail.Event.State == models.PublicationStatePublished
	default:
		return false
	}
}

func (h *Handler) ticketsPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	filter, err := maintenance.ParseTicketFilter(c)
	if err != nil {
		return h.writeError(c, err)
	}
	return h.renderTenantTickets(c, actorID, filter)
}

func (h *Handler) renderTenantTickets(c echo.Context, actorID uuid.UUID, filter maintenance.TicketFilter) error {
	tickets, err := h.maintenance.ListTicketsForTenantFlat(actorID, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	listFilter := tenantviews.TicketListFilterFromQuery(filter, c.QueryParam("from"), c.QueryParam("to"))
	if isHTMX(c) {
		return renderComponent(c, tenantviews.TicketsPanel(tickets, listFilter))
	}
	return renderComponent(c, tenantviews.TicketsPage(tickets, listFilter))
}

func (h *Handler) createTicket(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := maintenance.CreateTicketRequest{
		Category:    models.CategoryIncident,
		Severity:    models.Severity(c.FormValue("severity")),
		Impact:      models.Impact(c.FormValue("impact")),
		Description: c.FormValue("description"),
	}
	if _, err := h.maintenance.CreateTicket(actorID, req); err != nil {
		return h.writeError(c, err)
	}
	return h.renderTenantTickets(c, actorID, maintenance.TicketFilter{})
}

func (h *Handler) ticketDetailPage(c echo.Context) error {
	actorID, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	ticket, err := h.maintenance.GetTicketByID(id)
	if err != nil {
		return h.writeError(c, err)
	}
	ok, err := h.maintenance.TenantCanAccessTicket(actorID, ticket)
	if err != nil {
		return h.writeError(c, err)
	}
	if !ok {
		return h.writeError(c, maintenance.ErrUnauthorized)
	}
	return renderComponent(c, tenantviews.TicketDetailPage(ticket, actorID))
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
	if err := h.maintenance.DeleteOwnTicket(actorID, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/tenant/tickets")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) chatPage(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	result, err := h.chat.SearchChats(tenant.ID, "")
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := selectedRoomID(c, result.Rooms)
	if err != nil {
		return h.writeError(c, err)
	}
	messages, err := h.chatMessagesForRoom(tenant.ID, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, tenantviews.ChatPage(result, roomID, messages, tenant.ID))
}

func (h *Handler) chatSidebarFragment(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	result, err := h.chat.SearchChats(tenant.ID, c.QueryParam("q"))
	if err != nil {
		return h.writeError(c, err)
	}
	activeRoomID, _ := optionalRoomID(c.QueryParam("room"))
	return renderComponent(c, tenantviews.ChatSidebarBody(result, activeRoomID))
}

func (h *Handler) chatPanelFragment(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := paramUUID(c, "room")
	if err != nil {
		return h.writeError(c, err)
	}
	return h.renderChatPanel(c, tenant, roomID)
}

func (h *Handler) openDirectChat(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	otherID, err := uuid.Parse(c.FormValue("tenant_id"))
	if err != nil {
		return h.writeError(c, chat.ErrValidation)
	}
	room, err := h.chat.OpenDirectChat(tenant.ID, otherID)
	if err != nil {
		return h.writeError(c, err)
	}
	return h.renderChatOpened(c, tenant, room)
}

func (h *Handler) createGroupChat(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	room, err := h.chat.CreateGroup(tenant.ID, chat.CreateGroupRequest{
		Title: c.FormValue("title"),
		Topic: c.FormValue("topic"),
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return h.renderChatOpened(c, tenant, room)
}

func (h *Handler) renderChatOpened(c echo.Context, tenant models.Tenant, room models.ChatRoom) error {
	messages, err := h.chatMessagesForRoom(tenant.ID, room.ID)
	if err != nil {
		return h.writeError(c, err)
	}
	memberships, err := h.chat.ListRooms(tenant.ID)
	if err != nil {
		return h.writeError(c, err)
	}
	result := tenantviews.SidebarResultFromMemberships(memberships, "")
	c.Response().Header().Set("HX-Push-Url", tenantviews.ChatRoomURL(room.ID))
	return renderComponent(c, tenantviews.ChatOpenedUpdate(result, room, messages, tenant.ID))
}

func (h *Handler) renderChatPanel(c echo.Context, tenant models.Tenant, roomID uuid.UUID) error {
	messages, err := h.chatMessagesForRoom(tenant.ID, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	memberships, err := h.chat.ListRooms(tenant.ID)
	if err != nil {
		return h.writeError(c, err)
	}
	title := tenantviews.RoomTitle(memberships, roomID)
	return renderComponent(c, tenantviews.ChatPanel(roomID, messages, tenant.ID, title))
}

func (h *Handler) chatMessagesForRoom(tenantID, roomID uuid.UUID) ([]models.Message, error) {
	if roomID == uuid.Nil {
		return nil, nil
	}
	return h.chat.ListMessages(tenantID, roomID)
}

func (h *Handler) viewerTenant(c echo.Context) (models.Tenant, error) {
	actorID, err := actorID(c)
	if err != nil {
		return models.Tenant{}, err
	}
	return h.chat.GetTenantByUserID(actorID)
}

func (h *Handler) chatMessagesFragment(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := paramUUID(c, "room")
	if err != nil {
		return h.writeError(c, err)
	}
	messages, err := h.chat.ListMessages(tenant.ID, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, tenantviews.ChatMessages(messages, tenant.ID))
}

func (h *Handler) sendChatMessage(c echo.Context) error {
	tenant, err := h.viewerTenant(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.FormValue("room_id"))
	if err != nil {
		return h.writeError(c, chat.ErrValidation)
	}
	if _, err := h.chat.SendMessage(tenant.ID, roomID, c.FormValue("body")); err != nil {
		return h.writeError(c, err)
	}
	messages, err := h.chat.ListMessages(tenant.ID, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, tenantviews.ChatMessages(messages, tenant.ID))
}

func selectedRoomID(c echo.Context, memberships []models.Membership) (uuid.UUID, error) {
	if v := c.QueryParam("room"); v != "" {
		return uuid.Parse(v)
	}
	if len(memberships) == 0 {
		return uuid.Nil, nil
	}
	if memberships[0].ChatRoom != nil {
		return memberships[0].ChatRoom.ID, nil
	}
	return memberships[0].RoomID, nil
}

func optionalRoomID(raw string) (uuid.UUID, error) {
	if raw == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(raw)
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func isHTMX(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func actorID(c echo.Context) (uuid.UUID, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return uuid.Nil, chat.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, chat.ErrUnauthorized
	}
	raw, ok := claims["id"]
	if !ok {
		return uuid.Nil, chat.ErrUnauthorized
	}
	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, chat.ErrUnauthorized
	}
}

func paramUUID(c echo.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, chat.ErrValidation
	}
	return id, nil
}

func (h *Handler) writeError(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	category := "internal_error"

	switch {
	case errors.Is(err, chat.ErrValidation), errors.Is(err, maintenance.ErrValidation), errors.Is(err, administration.ErrValidation):
		status = http.StatusBadRequest
		category = "validation_error"
	case errors.Is(err, chat.ErrUnauthorized), errors.Is(err, maintenance.ErrUnauthorized), errors.Is(err, administration.ErrUnauthorized):
		status = http.StatusForbidden
		category = "authorization_denied"
	case errors.Is(err, chat.ErrNotAMember):
		status = http.StatusForbidden
		category = "not_a_member"
	case errors.Is(err, chat.ErrNotFound), errors.Is(err, maintenance.ErrNotFound), errors.Is(err, administration.ErrNotFound), errors.Is(err, chat.ErrTenantNotFound):
		status = http.StatusNotFound
		category = "not_found"
	case errors.Is(err, administration.ErrCapacityConflict), errors.Is(err, administration.ErrConcurrencyConflict):
		status = http.StatusConflict
		category = "conflict"
	}

	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}
