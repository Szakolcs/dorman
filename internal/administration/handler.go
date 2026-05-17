package administration

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"dorm-man/internal/pagination"
	"dorm-man/internal/platform"
	models "dorm-man/internal/models/administration"
	adminviews "dorm-man/web/templates/administration"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func (h *Handler) actor(c echo.Context) (Principal, error) {
	actorID, ok := platform.ActorUserID(c)
	if !ok {
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
	params := pageParams(c)
	filter := TenantListFilter{
		Status: c.QueryParam("status"),
		Search: c.QueryParam("search"),
		Params: params,
	}
	tenants, total, err := h.service.ListTenants(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, tenants, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	filter := RoomListFilter{
		State:  strings.ToLower(c.QueryParam("state")),
		Search: c.QueryParam("search"),
		Params: params,
	}
	rooms, total, err := h.service.ListRooms(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, rooms, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	items, total, err := h.service.ListInventory(InventoryListFilter{Params: params})
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, items, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	filter := TicketListFilter{
		ApprovalState: c.QueryParam("approval_state"),
		Status:        c.QueryParam("status"),
		Params:        params,
	}
	tickets, total, err := h.service.ListMaintenanceTickets(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, tickets, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	jobs, total, err := h.service.ListOperationalJobs(JobListFilter{
		AssigneeUserID: assigneeID,
		Date:           date,
		Params:         params,
	})
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, jobs, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	posts, total, err := h.service.ListNews(PublicationListFilter{Params: params})
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, posts, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	activities, total, err := h.service.ListActivities(PublicationListFilter{Params: params})
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, activities, pagination.NewMeta(params, total))
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
	params := pageParams(c)
	events, total, err := h.service.ListEvents(PublicationListFilter{Params: params})
	if err != nil {
		return h.writeError(c, err)
	}
	return writeListJSON(c, events, pagination.NewMeta(params, total))
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

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) dashboardPage(c echo.Context) error {
	stats, err := h.service.DashboardStats()
	if err != nil {
		return err
	}
	roomRows, err := h.service.RoomsWithOldestAssignedTickets(5)
	if err != nil {
		return err
	}
	openTickets, err := h.service.OpenMaintenanceTickets(5)
	if err != nil {
		return err
	}
	jobsToday, err := h.service.JobsScheduledToday(5)
	if err != nil {
		return err
	}
	publications, err := h.service.RecentPublications(5)
	if err != nil {
		return err
	}
	events, err := h.service.RecentEvents(5)
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.DashboardPage(dashboardPageData(
		stats, roomRows, openTickets, jobsToday, publications, events,
	)))
}

