package chat

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	htmlchat "dorm-man/web/templates/chat"

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

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) tenantActor(c echo.Context) (TenantPrincipal, error) {
	for _, raw := range []string{
		c.FormValue("actor_tenant_id"),
		c.QueryParam("tenant_id"),
		c.Request().Header.Get("X-Actor-Tenant-ID"),
	} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		return TenantPrincipal{TenantID: id}, nil
	}

	value := c.Request().Header.Get("X-Actor-User-ID")
	if value == "" {
		return TenantPrincipal{}, ErrUnauthorized
	}
	userID, err := uuid.Parse(value)
	if err != nil {
		return TenantPrincipal{}, ErrUnauthorized
	}
	return h.service.ResolveTenantPrincipal(userID)
}

func (h *Handler) viewActor(c echo.Context) (TenantPrincipal, string, error) {
	p, err := h.tenantActor(c)
	if err != nil {
		return TenantPrincipal{}, "", err
	}
	return p, p.TenantID.String(), nil
}

func (h *Handler) writeError(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	category := "internal_error"

	switch {
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
		category = "validation_error"
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusForbidden
		category = "authorization_denied"
	case errors.Is(err, ErrNotAMember):
		status = http.StatusForbidden
		category = "not_a_member"
	case errors.Is(err, ErrNotGroupOwner), errors.Is(err, ErrFlatMembership), errors.Is(err, ErrRoomKindMismatch):
		status = http.StatusForbidden
		category = "authorization_denied"
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		category = "not_found"
	case errors.Is(err, ErrTenantNotFound):
		status = http.StatusNotFound
		category = "tenant_not_found"
	}

	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}

func (h *Handler) listConversations(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	list, err := h.service.ListConversations(p)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": list})
}

func (h *Handler) getRoom(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	detail, err := h.service.GetRoom(p, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": detail})
}

func (h *Handler) listMessages(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}

	var beforeID *uuid.UUID
	if raw := c.QueryParam("before_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		beforeID = &id
	}

	limit := queryPositiveInt(c, "limit", 50)
	msgs, err := h.service.ListMessages(p, roomID, MessageListFilter{
		BeforeMessageID: beforeID,
		Limit:           limit,
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": msgs})
}

func (h *Handler) sendMessage(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body SendMessageInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	msg, err := h.service.SendMessage(p, roomID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": msg})
}

func (h *Handler) markRead(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body MarkReadInput
	if err := c.Bind(&body); err != nil && c.Request().ContentLength > 0 {
		return h.writeError(c, ErrValidation)
	}
	member, err := h.service.MarkRoomRead(p, roomID, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": member})
}

func (h *Handler) openDirect(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body OpenDirectInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	room, err := h.service.OpenDirect(p, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": room})
}

func (h *Handler) createGroup(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body CreateGroupInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	room, err := h.service.CreateGroup(p, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": room})
}

func (h *Handler) addGroupMember(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body GroupMemberInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	member, err := h.service.AddGroupMember(p, roomID, body.TenantID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": member})
}

func (h *Handler) removeGroupMember(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	tenantID, err := uuid.Parse(c.Param("tenantId"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	member, err := h.service.RemoveGroupMember(p, roomID, tenantID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": member})
}

func (h *Handler) leaveGroup(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	member, err := h.service.LeaveGroup(p, roomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": member})
}

func (h *Handler) getProfile(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	view, err := h.service.GetProfile(p)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": view})
}

func (h *Handler) updateProfile(c echo.Context) error {
	p, err := h.tenantActor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body UpdateProfileInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	view, err := h.service.UpdateProfile(p, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": view})
}

func queryPositiveInt(c echo.Context, name string, def int) int {
	s := c.QueryParam(name)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (h *Handler) workspacePage(c echo.Context) error {
	_, actorID, err := h.viewActor(c)
	if err != nil {
		return renderComponent(c, htmlchat.ChatShell("Chat", "", htmlchat.ChatEmptyBody()))
	}
	return renderComponent(c, htmlchat.WorkspacePage(actorID))
}

func (h *Handler) conversationsPartial(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	list, err := h.service.ListConversations(p)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderComponent(c, htmlchat.ConversationListPartial(actorID, list))
}

func (h *Handler) roomPage(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return renderComponent(c, htmlchat.ChatShell("Chat", "", htmlchat.ChatEmptyBody()))
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}
	return h.renderRoom(c, p, actorID, roomID, true)
}

func (h *Handler) roomPartial(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}
	return h.renderRoom(c, p, actorID, roomID, false)
}

func (h *Handler) renderRoom(c echo.Context, p TenantPrincipal, actorID string, roomID uuid.UUID, fullPage bool) error {
	detail, err := h.service.GetRoom(p, roomID)
	if err != nil {
		if fullPage {
			return renderComponent(c, htmlchat.ChatShell("Chat room", actorID, htmlchat.RoomError(err.Error())))
		}
		return renderComponent(c, htmlchat.RoomError(err.Error()))
	}
	msgs, err := h.service.ListMessages(p, roomID, MessageListFilter{Limit: 50})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	clientID := uuid.NewString()
	panel := htmlchat.RoomPanelPartial(actorID, detail, msgs, clientID)
	if fullPage && c.Request().Header.Get("HX-Request") != "true" {
		return renderComponent(c, htmlchat.RoomPage(actorID, panel))
	}
	return renderComponent(c, panel)
}

func (h *Handler) messagesPartial(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}
	msgs, err := h.service.ListMessages(p, roomID, MessageListFilter{Limit: 50})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return renderComponent(c, htmlchat.MessageListPartial(actorID, roomID.String(), msgs))
}

func (h *Handler) sendMessageView(c echo.Context) error {
	p, _, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}

	clientRaw := strings.TrimSpace(c.FormValue("client_message_id"))
	clientID, err := uuid.Parse(clientRaw)
	if err != nil {
		return c.String(http.StatusBadRequest, "client_message_id required")
	}

	msg, err := h.service.SendMessage(p, roomID, SendMessageInput{
		Body:            c.FormValue("body"),
		ClientMessageID: &clientID,
	})
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if view, err := h.service.GetProfile(p); err == nil {
		msg.AuthorTenant = view.Tenant
	}
	return renderComponent(c, htmlchat.MessageItem(msg))
}

func (h *Handler) markReadView(c echo.Context) error {
	p, _, err := h.viewActor(c)
	if err != nil {
		return c.NoContent(http.StatusForbidden)
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	_, err = h.service.MarkRoomRead(p, roomID, MarkReadInput{})
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) openDirectView(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	otherRaw := strings.TrimSpace(c.FormValue("other_tenant_id"))
	otherID, err := uuid.Parse(otherRaw)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid other tenant id")
	}
	room, err := h.service.OpenDirect(p, OpenDirectInput{OtherTenantID: otherID})
	if err != nil {
		return renderComponent(c, htmlchat.RoomError(err.Error()))
	}
	return h.renderRoom(c, p, actorID, room.ID, false)
}

func (h *Handler) profilePage(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return renderComponent(c, htmlchat.ChatShell("Chat profile", "", htmlchat.ChatEmptyBody()))
	}
	view, err := h.service.GetProfile(p)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	bio := ""
	if view.Profile != nil {
		bio = view.Profile.Bio
	}
	return renderComponent(c, htmlchat.ProfilePage(actorID, view, bio))
}

func (h *Handler) profileSaveView(c echo.Context) error {
	p, _, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}

	var nickPtr *string
	if raw := strings.TrimSpace(c.FormValue("nickname")); raw != "" {
		nickPtr = &raw
	}
	bio := strings.TrimSpace(c.FormValue("bio"))
	avatar := strings.TrimSpace(c.FormValue("avatar_storage_key"))

	view, err := h.service.UpdateProfile(p, UpdateProfileInput{
		Nickname:         nickPtr,
		Bio:              &bio,
		AvatarStorageKey: &avatar,
	})
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderComponent(c, htmlchat.ProfileSaved(view))
	}
	return c.Redirect(http.StatusSeeOther, "/chat/view/profile?tenant_id="+p.TenantID.String())
}

func (h *Handler) groupNewPage(c echo.Context) error {
	_, actorID, err := h.viewActor(c)
	if err != nil {
		return renderComponent(c, htmlchat.ChatShell("New group", "", htmlchat.ChatEmptyBody()))
	}
	return renderComponent(c, htmlchat.GroupNewPage(actorID))
}

func (h *Handler) groupCreateView(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}

	memberIDs, err := parseTenantIDList(c.FormValue("member_tenant_ids"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	room, err := h.service.CreateGroup(p, CreateGroupInput{
		Title:            c.FormValue("title"),
		AvatarStorageKey: c.FormValue("avatar_storage_key"),
		MemberTenantIDs:  memberIDs,
	})
	if err != nil {
		return renderComponent(c, htmlchat.RoomError(err.Error()))
	}
	return h.renderRoom(c, p, actorID, room.ID, false)
}

func (h *Handler) roomMembersView(c echo.Context) error {
	p, actorID, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}

	action := strings.TrimSpace(c.FormValue("action"))
	memberRaw := strings.TrimSpace(c.FormValue("member_tenant_id"))
	memberID, err := uuid.Parse(memberRaw)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid member tenant id")
	}

	switch action {
	case "add":
		_, err = h.service.AddGroupMember(p, roomID, memberID)
	case "remove":
		_, err = h.service.RemoveGroupMember(p, roomID, memberID)
	default:
		return c.String(http.StatusBadRequest, "action must be add or remove")
	}
	if err != nil {
		return renderComponent(c, htmlchat.RoomError(err.Error()))
	}
	return h.renderRoom(c, p, actorID, roomID, false)
}

func (h *Handler) leaveGroupView(c echo.Context) error {
	p, _, err := h.viewActor(c)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid room id")
	}
	if _, err := h.service.LeaveGroup(p, roomID); err != nil {
		return renderComponent(c, htmlchat.RoomError(err.Error()))
	}
	return renderComponent(c, htmlchat.RoomError("You left the group."))
}

func parseTenantIDList(raw string) ([]uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := uuid.Parse(part)
		if err != nil {
			return nil, ErrValidation
		}
		out = append(out, id)
	}
	return out, nil
}
