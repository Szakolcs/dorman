package administration

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	models "dorm-man/internal/models/administration"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) createTenantView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/tenants", err)
	}
	n := models.NationalityHungarian
	if c.FormValue("nationality") == "international" {
		n = models.NationalityInternational
	}
	nationality := n
	tenant := models.Tenant{
		Name:        c.FormValue("name"),
		StudentCode: c.FormValue("student_code"),
		Email:       c.FormValue("email"),
		Nationality: &nationality,
		IsActive:    true,
	}
	_, err = h.service.RegisterTenant(principal, tenant)
	return redirectView(c, "/administration/tenants", err)
}

func (h *Handler) deactivateTenantView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/tenants", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectView(c, "/administration/tenants", ErrValidation)
	}
	_, err = h.service.SetTenantActive(principal, id, false)
	return redirectView(c, "/administration/tenants", err)
}

func (h *Handler) assignTenantView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/rooms", err)
	}
	tenantID, err := uuid.Parse(c.FormValue("tenant_id"))
	if err != nil {
		return redirectView(c, "/administration/rooms", ErrValidation)
	}
	roomID, err := uuid.Parse(c.FormValue("room_id"))
	if err != nil {
		return redirectView(c, "/administration/rooms", ErrValidation)
	}
	_, err = h.service.AssignTenant(principal, tenantID, roomID)
	return redirectView(c, "/administration/rooms", err)
}

func (h *Handler) generatePlanView(c echo.Context) error {
	plan, err := h.service.GenerateAllocationPlan()
	if err != nil {
		return redirectView(c, "/administration/rooms", err)
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return redirectView(c, "/administration/rooms", err)
	}
	return h.roomsPageWithPlan(c, string(raw))
}

func (h *Handler) approvePlanView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/rooms", err)
	}
	var plan AssignmentPlan
	if err := json.Unmarshal([]byte(c.FormValue("plan_json")), &plan); err != nil {
		return redirectView(c, "/administration/rooms", ErrValidation)
	}
	err = h.service.ApproveAllocationPlan(principal, plan)
	return redirectView(c, "/administration/rooms", err)
}

func (h *Handler) createInventoryView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/inventory", err)
	}
	item := models.InventoryItem{
		Name:         c.FormValue("name"),
		LocationType: models.InventoryLocationType(c.FormValue("location_type")),
		Condition:    models.InventoryCondition(c.FormValue("condition")),
		Status:       models.InventoryStatusInStock,
	}
	if rid := strings.TrimSpace(c.FormValue("room_id")); rid != "" {
		id, parseErr := uuid.Parse(rid)
		if parseErr != nil {
			return redirectView(c, "/administration/inventory", ErrValidation)
		}
		item.RoomID = &id
	}
	_, err = h.service.CreateInventoryItem(principal, item)
	return redirectView(c, "/administration/inventory", err)
}

func (h *Handler) updateInventoryStatusView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/inventory", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectView(c, "/administration/inventory", ErrValidation)
	}
	status := models.InventoryStatus(c.FormValue("status"))
	condition := models.InventoryCondition(c.FormValue("condition"))
	_, err = h.service.UpdateInventoryStatus(principal, id, status, condition, nil)
	return redirectView(c, "/administration/inventory", err)
}

func (h *Handler) createMaintenanceTicketView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/maintenance", err)
	}
	ticket := models.MaintenanceTicket{
		Description:     c.FormValue("description"),
		Category:        models.MaintenanceCategory(c.FormValue("category")),
		Severity:        models.MaintenanceSeverity(c.FormValue("severity")),
		Impact:          models.MaintenanceImpact(c.FormValue("impact")),
		CreatedByUserID: principal.UserID,
	}
	_, err = h.service.CreateMaintenanceTicket(principal, ticket)
	return redirectView(c, "/administration/maintenance", err)
}

func (h *Handler) approveMaintenanceTicketView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/maintenance", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectView(c, "/administration/maintenance", ErrValidation)
	}
	var assigneeID *uuid.UUID
	if v := strings.TrimSpace(c.FormValue("assignee_user_id")); v != "" {
		parsed, parseErr := uuid.Parse(v)
		if parseErr != nil {
			return redirectView(c, "/administration/maintenance", ErrValidation)
		}
		assigneeID = &parsed
	}
	var dueAt *time.Time
	if v := strings.TrimSpace(c.FormValue("due_at")); v != "" {
		t, parseErr := time.Parse("2006-01-02T15:04", v)
		if parseErr != nil {
			return redirectView(c, "/administration/maintenance", ErrValidation)
		}
		dueAt = &t
	}
	_, err = h.service.ApproveMaintenanceTicket(principal, id, assigneeID, dueAt)
	return redirectView(c, "/administration/maintenance", err)
}

func (h *Handler) transitionMaintenanceTicketView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/maintenance", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectView(c, "/administration/maintenance", ErrValidation)
	}
	toStatus := models.MaintenanceStatus(c.FormValue("to_status"))
	_, err = h.service.TransitionMaintenanceTicket(principal, id, toStatus, c.FormValue("note"))
	return redirectView(c, "/administration/maintenance", err)
}

func (h *Handler) createJobView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/jobs", err)
	}
	assigneeID, err := uuid.Parse(c.FormValue("assignee_user_id"))
	if err != nil {
		return redirectView(c, "/administration/jobs", ErrValidation)
	}
	startsAt, err := time.Parse("2006-01-02T15:04", c.FormValue("starts_at"))
	if err != nil {
		return redirectView(c, "/administration/jobs", ErrValidation)
	}
	endsAt, err := time.Parse("2006-01-02T15:04", c.FormValue("ends_at"))
	if err != nil {
		return redirectView(c, "/administration/jobs", ErrValidation)
	}
	job := models.OperationalJob{
		Title:          c.FormValue("title"),
		Description:    c.FormValue("description"),
		AssigneeUserID: assigneeID,
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		Priority:       models.JobPriority(c.FormValue("priority")),
	}
	allowConflict := c.FormValue("allow_conflict") == "true"
	_, err = h.service.CreateOperationalJob(principal, job, allowConflict)
	return redirectView(c, "/administration/jobs", err)
}

func (h *Handler) createNewsView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/publications", err)
	}
	input := NewsUpsertInput{
		Title: c.FormValue("title"),
		Body:  c.FormValue("body"),
		Tags:  c.FormValue("tags"),
		State: models.PublicationState(c.FormValue("state")),
	}
	_, err = h.service.CreateNews(principal, input)
	return redirectView(c, "/administration/publications", err)
}

func (h *Handler) createActivityView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/publications", err)
	}
	capacity, _ := strconv.Atoi(c.FormValue("capacity"))
	input := ActivityUpsertInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Location:    c.FormValue("location"),
		Capacity:    capacity,
		State:       models.PublicationState(c.FormValue("state")),
	}
	_, err = h.service.CreateActivity(principal, input)
	return redirectView(c, "/administration/publications", err)
}

func (h *Handler) createEventView(c echo.Context) error {
	principal, err := h.actor(c)
	if err != nil {
		return redirectView(c, "/administration/publications", err)
	}
	startsAt, err := time.Parse("2006-01-02T15:04", c.FormValue("starts_at"))
	if err != nil {
		return redirectView(c, "/administration/publications", ErrValidation)
	}
	endsAt, err := time.Parse("2006-01-02T15:04", c.FormValue("ends_at"))
	if err != nil {
		return redirectView(c, "/administration/publications", ErrValidation)
	}
	capacity, _ := strconv.Atoi(c.FormValue("capacity"))
	input := EventUpsertInput{
		Title:           c.FormValue("title"),
		Description:     c.FormValue("description"),
		Location:        c.FormValue("location"),
		OrganizerUserID: principal.UserID,
		StartsAt:        startsAt,
		EndsAt:          endsAt,
		Capacity:        capacity,
		State:           models.PublicationState(c.FormValue("state")),
	}
	_, err = h.service.CreateEvent(principal, input)
	return redirectView(c, "/administration/publications", err)
}
