package administration

import (
	"errors"
	"net/http"
	"strings"
	"time"

	models "dorm-man/internal/models/administration"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func (h *Handler) actor(c echo.Context) (Principal, error) {
	value := c.Request().Header.Get("X-Actor-User-ID")
	if value == "" {
		return Principal{}, ErrUnauthorized
	}
	actorID, err := uuid.Parse(value)
	if err != nil {
		return Principal{}, ErrUnauthorized
	}
	return h.service.ResolvePrincipal(actorID)
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

func (h *Handler) listTenants(c echo.Context) error {
	tenants, err := h.service.ListTenants(TenantListFilter{
		Status: c.QueryParam("status"),
		Search: c.QueryParam("search"),
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": tenants})
}

func (h *Handler) getTenant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": tenant})
}

func (h *Handler) createTenant(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body models.Tenant
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	created, err := h.service.RegisterTenant(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": created})
}

func (h *Handler) activateTenant(c echo.Context) error {
	return h.updateTenantStatus(c, true)
}

func (h *Handler) deactivateTenant(c echo.Context) error {
	return h.updateTenantStatus(c, false)
}

func (h *Handler) updateTenantStatus(c echo.Context, active bool) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	tenant, err := h.service.SetTenantActive(principal, id, active)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": tenant})
}

func (h *Handler) listRooms(c echo.Context) error {
	rooms, err := h.service.ListRooms(RoomListFilter{
		State:  strings.ToLower(c.QueryParam("state")),
		Search: c.QueryParam("search"),
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": rooms})
}

func (h *Handler) getRoom(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	room, err := h.service.GetRoom(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": room})
}

func (h *Handler) assignTenant(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body struct {
		TenantID uuid.UUID `json:"tenant_id"`
		RoomID   uuid.UUID `json:"room_id"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	assignment, err := h.service.AssignTenant(principal, body.TenantID, body.RoomID)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": assignment})
}

func (h *Handler) generatePlan(c echo.Context) error {
	plan, err := h.service.GenerateAllocationPlan()
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": plan})
}

func (h *Handler) approvePlan(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body AssignmentPlan
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	err = h.service.ApproveAllocationPlan(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) listInventory(c echo.Context) error {
	items, err := h.service.ListInventory()
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createInventory(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body models.InventoryItem
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	item, err := h.service.CreateInventoryItem(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) updateInventoryStatus(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body struct {
		Status       models.InventoryStatus    `json:"status"`
		Condition    models.InventoryCondition `json:"condition"`
		WithdrawDate *time.Time                `json:"withdraw_date"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	item, err := h.service.UpdateInventoryStatus(principal, id, body.Status, body.Condition, body.WithdrawDate)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listMaintenanceTickets(c echo.Context) error {
	tickets, err := h.service.ListMaintenanceTickets(TicketListFilter{
		ApprovalState: c.QueryParam("approval_state"),
		Status:        c.QueryParam("status"),
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": tickets})
}

func (h *Handler) createMaintenanceTicket(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body models.MaintenanceTicket
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	ticket, err := h.service.CreateMaintenanceTicket(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": ticket})
}

func (h *Handler) approveMaintenanceTicket(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body struct {
		AssigneeUserID *uuid.UUID `json:"assignee_user_id"`
		DueAt          *time.Time `json:"due_at"`
	}
	if err := c.Bind(&body); err != nil && !errors.Is(err, echo.ErrUnsupportedMediaType) {
		return h.writeError(c, ErrValidation)
	}
	ticket, err := h.service.ApproveMaintenanceTicket(principal, id, body.AssigneeUserID, body.DueAt)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ticket})
}

func (h *Handler) transitionMaintenanceTicket(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body struct {
		ToStatus models.MaintenanceStatus `json:"to_status"`
		Note     string                   `json:"note"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	ticket, err := h.service.TransitionMaintenanceTicket(principal, id, body.ToStatus, body.Note)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": ticket})
}

func (h *Handler) listJobs(c echo.Context) error {
	var assigneeID *uuid.UUID
	if q := c.QueryParam("assignee_user_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		assigneeID = &id
	}
	var date *time.Time
	if q := c.QueryParam("date"); q != "" {
		parsed, err := time.Parse("2006-01-02", q)
		if err != nil {
			return h.writeError(c, ErrValidation)
		}
		date = &parsed
	}
	jobs, err := h.service.ListOperationalJobs(JobListFilter{
		AssigneeUserID: assigneeID,
		Date:           date,
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": jobs})
}

func (h *Handler) createJob(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body struct {
		models.OperationalJob
		AllowConflict bool `json:"allow_conflict"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	result, err := h.service.CreateOperationalJob(principal, body.OperationalJob, body.AllowConflict)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": result})
}

func (h *Handler) listNews(c echo.Context) error {
	posts, err := h.service.ListNews()
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": posts})
}

func (h *Handler) createNews(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body NewsUpsertInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.CreateNews(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": post})
}

func (h *Handler) publishNews(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	post, err := h.service.PublishNews(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": post})
}

func (h *Handler) listActivities(c echo.Context) error {
	activities, err := h.service.ListActivities()
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": activities})
}

func (h *Handler) createActivity(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body ActivityUpsertInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	activity, err := h.service.CreateActivity(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": activity})
}

func (h *Handler) publishActivity(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	activity, err := h.service.PublishActivity(principal, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": activity})
}

func (h *Handler) listEvents(c echo.Context) error {
	events, err := h.service.ListEvents()
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": events})
}

func (h *Handler) createEvent(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var body EventUpsertInput
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	if body.OrganizerUserID == uuid.Nil {
		body.OrganizerUserID = principal.UserID
	}
	event, err := h.service.CreateEvent(principal, body)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": event})
}

func (h *Handler) updateEventState(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	var body struct {
		State models.PublicationState `json:"state"`
	}
	if err := c.Bind(&body); err != nil {
		return h.writeError(c, ErrValidation)
	}
	event, err := h.service.UpdateEventState(principal, id, body.State)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": event})
}
