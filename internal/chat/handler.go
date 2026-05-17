package chat

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) tenantActor(c echo.Context) (TenantPrincipal, error) {
	if value := c.Request().Header.Get("X-Actor-Tenant-ID"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return TenantPrincipal{}, ErrUnauthorized
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
