package chat

import (
	"net/http"
	"strings"

	htmlchat "dorm-man/web/templates/chat"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) viewActor(c echo.Context) (TenantPrincipal, string, error) {
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
		return TenantPrincipal{TenantID: id}, id.String(), nil
	}
	p, err := h.tenantActor(c)
	if err != nil {
		return TenantPrincipal{}, "", err
	}
	return p, p.TenantID.String(), nil
}

func (h *Handler) workspacePage(c echo.Context) error {
	_, actorID, err := h.viewActor(c)
	if err != nil {
		return renderComponent(c, htmlchat.ChatShell("Chat", "", htmlchat.ActorStubForm()))
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
		return renderComponent(c, htmlchat.ChatShell("Chat", "", htmlchat.ActorStubForm()))
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
		return renderComponent(c, htmlchat.ChatShell("Chat profile", "", htmlchat.ActorStubForm()))
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
		return renderComponent(c, htmlchat.ChatShell("New group", "", htmlchat.ActorStubForm()))
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
