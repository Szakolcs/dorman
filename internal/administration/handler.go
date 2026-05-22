package administration

import (
	"dorm-man/internal/models"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dorm-man/internal/models/forum"
	adminviews "dorm-man/web/templates/administration"

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

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// actorID pulls the current user's UUID from the JWT claim. The JWT is
// expected to have been parsed by upstream middleware and stored on the echo
// context under the key "user" (the convention used by echo-jwt).
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

func queryUUID(c echo.Context, key string) (uuid.UUID, error) {
	raw := c.QueryParam(key)
	if raw == "" {
		raw = c.FormValue(key)
	}
	if raw == "" {
		return uuid.Nil, ErrValidation
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, ErrValidation
	}
	return id, nil
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, time.DateTime, time.DateOnly} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func parsePagination(c echo.Context) Pagination {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	per, _ := strconv.Atoi(c.QueryParam("per_page"))
	return Pagination{
		Page:    page,
		PerPage: per,
		Sort:    c.QueryParam("sort"),
		Order:   c.QueryParam("order"),
	}
}

func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
	case errors.Is(err, ErrCapacityConflict):
		category = "capacity_conflict"
		status = http.StatusConflict
	case errors.Is(err, ErrConcurrencyConflict):
		category = "concurrency_conflict"
		status = http.StatusConflict
	case errors.Is(err, ErrNotFound):
		category = "not_found"
		status = http.StatusNotFound
	case errors.Is(err, ErrStateTransition):
		category = "state_transition_error"
		status = http.StatusConflict
	case errors.Is(err, ErrStudentStatusInvalid):
		category = "student_status_invalid"
		status = http.StatusConflict
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

// ---------------------------------------------------------------------------
// dashboard
// ---------------------------------------------------------------------------

func (h *Handler) dashboardPage(c echo.Context) error {
	summary, err := h.service.GetDashboardSummary()
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(
		c,
		adminviews.DashboardPage(
			summary.ActiveTenants,
			summary.OpenJobs,
			summary.Buildings,
			summary.RecentAudits,
		),
	)
}

// ---------------------------------------------------------------------------
// tenants
// ---------------------------------------------------------------------------

func (h *Handler) tenantsPage(c echo.Context) error {
	filter := TenantFilter{
		Search:     c.QueryParam("search"),
		Pagination: parsePagination(c),
	}
	if v := c.QueryParam("degree"); v != "" {
		d := models.DegreeType(v)
		filter.Degree = &d
	}
	if v := c.QueryParam("faculty"); v != "" {
		f := models.FacultyType(v)
		filter.Faculty = &f
	}
	if v := c.QueryParam("sex"); v != "" {
		x := models.SexType(v)
		filter.Sex = &x
	}
	if v := c.QueryParam("nationality"); v != "" {
		n := models.NationalityType(v)
		filter.Nationality = &n
	}
	if v := c.QueryParam("is_active"); v != "" {
		b := strings.EqualFold(v, "true")
		filter.IsActive = &b
	}

	tenants, err := h.service.ListTenants(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.TenantsPage(tenants))
}

func (h *Handler) tenantDetailPage(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.TenantDetailPage(tenant))
}

// ---------------------------------------------------------------------------
// inventory
// ---------------------------------------------------------------------------

func (h *Handler) inventoryPage(c echo.Context) error {
	filter := InventoryFilter{
		Search:        c.QueryParam("search"),
		PurchasedFrom: parseDate(c.QueryParam("purchased_from")),
		PurchasedTo:   parseDate(c.QueryParam("purchased_to")),
		Pagination:    parsePagination(c),
	}
	if v := c.QueryParam("condition"); v != "" {
		cond := models.InventoryCondition(v)
		filter.Condition = &cond
	}
	if v := c.QueryParam("status"); v != "" {
		st := models.InventoryStatus(v)
		filter.Status = &st
	}
	items, err := h.service.ListInventory(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.InventoryPage(items))
}

func (h *Handler) inventoryDetailPage(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	item, err := h.service.GetInventoryItem(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.InventoryDetailPage(item))
}

func (h *Handler) createInventoryItem(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := CreateInventoryItemRequest{
		Name:         c.FormValue("name"),
		Description:  c.FormValue("description"),
		Condition:    models.InventoryCondition(c.FormValue("condition")),
		Status:       models.InventoryStatus(c.FormValue("status")),
		PurchaseDate: derefTime(parseDate(c.FormValue("purchase_date"))),
	}
	if v := c.FormValue("room_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.RoomID = &id
		}
	}
	if v := c.FormValue("flat_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.FlatID = &id
		}
	}
	if v := c.FormValue("building_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.BuildingID = &id
		}
	}
	if v := c.FormValue("shared_area_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.SharedAreaID = &id
		}
	}
	if _, err := h.service.CreateInventoryItem(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/inventory")
	return c.NoContent(http.StatusOK)
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (h *Handler) updateInventoryItemStatus(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	req := UpdateInventoryStatusRequest{
		ID:     id,
		Status: models.InventoryStatus(c.FormValue("status")),
	}
	if v := c.FormValue("condition"); v != "" {
		cond := models.InventoryCondition(v)
		req.Condition = &cond
	}
	if _, err := h.service.UpdateInventoryItemStatus(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/inventory/"+id.String())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) deleteInventoryItem(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.DeleteInventoryItem(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/inventory")
	return c.NoContent(http.StatusOK)
}

// ---------------------------------------------------------------------------
// jobs
// ---------------------------------------------------------------------------

func (h *Handler) jobsPage(c echo.Context) error {
	filter := JobsFilter{
		Search:     c.QueryParam("search"),
		From:       parseDate(c.QueryParam("from")),
		To:         parseDate(c.QueryParam("to")),
		Pagination: parsePagination(c),
	}
	if v := c.QueryParam("priority"); v != "" {
		p := models.JobPriority(v)
		filter.Priority = &p
	}
	if v := c.QueryParam("status"); v != "" {
		s := models.JobStatus(v)
		filter.Status = &s
	}
	jobs, err := h.service.ListJobs(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.JobsPage(jobs))
}

func (h *Handler) jobDetailPage(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	job, err := h.service.GetJob(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.JobDetailPage(job))
}

func (h *Handler) createJob(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := CreateJobRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		StartsAt:    derefTime(parseDate(c.FormValue("starts_at"))),
		EndsAt:      derefTime(parseDate(c.FormValue("ends_at"))),
		Priority:    models.JobPriority(c.FormValue("priority")),
		Status:      models.JobStatus(c.FormValue("status")),
	}
	if _, err := h.service.CreateJob(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/jobs")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) updateJob(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := queryUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	req := UpdateJobRequest{
		ID:          id,
		Title:       optString(c.FormValue("title")),
		Description: optString(c.FormValue("description")),
		StartsAt:    parseDate(c.FormValue("starts_at")),
		EndsAt:      parseDate(c.FormValue("ends_at")),
	}
	if v := c.FormValue("priority"); v != "" {
		p := models.JobPriority(v)
		req.Priority = &p
	}
	if v := c.FormValue("status"); v != "" {
		st := models.JobStatus(v)
		req.Status = &st
	}
	if _, err := h.service.UpdateJob(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/jobs/"+id.String())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteJob(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := queryUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.DeleteJob(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/jobs")
	return c.NoContent(http.StatusOK)
}

// ---------------------------------------------------------------------------
// publications (news / activities / events)
// ---------------------------------------------------------------------------

func (h *Handler) publicationsPage(c echo.Context) error {
	filter := PublicationFilter{
		Search:     c.QueryParam("search"),
		Kind:       PublicationKind(c.QueryParam("kind")),
		From:       parseDate(c.QueryParam("from")),
		To:         parseDate(c.QueryParam("to")),
		Pagination: parsePagination(c),
	}
	if v := c.QueryParam("state"); v != "" {
		st := forum.PublicationState(v)
		filter.State = &st
	}
	pubs, err := h.service.ListPublications(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.PublicationsPage(pubs))
}

func (h *Handler) publicationDetailPage(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	detail, err := h.service.GetPublication(id)
	if err != nil {
		return h.writeError(c, err)
	}
	switch detail.Kind {
	case PublicationKindNews:
		return renderComponent(c, adminviews.NewsDetailPage(*detail.News))
	case PublicationKindActivity:
		return renderComponent(c, adminviews.ActivityDetailPage(*detail.Activity))
	case PublicationKindEvent:
		return renderComponent(c, adminviews.EventDetailPage(*detail.Event))
	default:
		return h.writeError(c, ErrNotFound)
	}
}

func (h *Handler) createNews(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := CreateNewsRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		State:       forum.PublicationState(c.FormValue("state")),
	}
	if _, err := h.service.CreateNews(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) archiveNews(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.ArchiveNews(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) createActivity(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	capacity, _ := strconv.Atoi(c.FormValue("capacity"))
	req := CreateActivityRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		State:       forum.PublicationState(c.FormValue("state")),
		Capacity:    capacity,
	}
	if v := c.FormValue("shared_area_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.SharedAreaID = &id
		}
	}
	if _, err := h.service.CreateActivity(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) archiveActivity(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.ArchiveActivity(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) createEvent(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := CreateEventRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		State:       forum.PublicationState(c.FormValue("state")),
		StartsAt:    derefTime(parseDate(c.FormValue("starts_at"))),
		EndsAt:      derefTime(parseDate(c.FormValue("ends_at"))),
	}
	if v := c.FormValue("shared_area_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.SharedAreaID = &id
		}
	}
	if _, err := h.service.CreateEvent(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) archiveEvent(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.ArchiveEvent(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/publications")
	return c.NoContent(http.StatusOK)
}

// ---------------------------------------------------------------------------
// housing
// ---------------------------------------------------------------------------

func (h *Handler) housingPage(c echo.Context) error {
	overview, err := h.service.GetHousingOverview()
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(
		c,
		adminviews.HousingPage(
			overview.Buildings,
			overview.Flats,
			overview.SharedAreas,
			overview.Rooms,
		),
	)
}

func (h *Handler) buildingDetail(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	b, err := h.service.GetBuilding(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.BuildingDetailPage(b))
}

func (h *Handler) flatDetail(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	f, err := h.service.GetFlat(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.FlatDetailPage(f))
}

func (h *Handler) sharedAreaDetail(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	sa, err := h.service.GetSharedArea(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.SharedAreaDetailPage(sa))
}

func (h *Handler) roomDetail(c echo.Context) error {
	id, err := paramUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	r, err := h.service.GetRoom(id)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.RoomDetailPage(r))
}

// ---------------------------------------------------------------------------
// room assignments
// ---------------------------------------------------------------------------

func (h *Handler) assign(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	tenantID, err := uuid.Parse(c.FormValue("tenant_id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	roomID, err := uuid.Parse(c.FormValue("room_id"))
	if err != nil {
		return h.writeError(c, ErrValidation)
	}
	req := AssignRoomRequest{
		TenantID:    tenantID,
		RoomID:      roomID,
		EffectiveAt: parseDate(c.FormValue("effective_at")),
	}
	if _, err := h.service.AssignRoom(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/housing/room/"+roomID.String())
	return c.NoContent(http.StatusOK)
}

func (h *Handler) massAssignment(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	var req MassAssignRequest
	if err := c.Bind(&req); err != nil {
		return h.writeError(c, ErrValidation)
	}
	if _, err := h.service.MassAssignRooms(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/housing")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) updateAssignment(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := queryUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	req := UpdateAssignmentRequest{
		ID:          id,
		EffectiveAt: parseDate(c.FormValue("effective_at")),
		EndedAt:     parseDate(c.FormValue("ended_at")),
	}
	if v := c.FormValue("room_id"); v != "" {
		if rid, err := uuid.Parse(v); err == nil {
			req.RoomID = &rid
		}
	}
	if _, err := h.service.UpdateAssignment(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/housing")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) deleteAssignment(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := queryUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	if err := h.service.DeleteAssignment(actor, id); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/housing")
	return c.NoContent(http.StatusOK)
}

// ---------------------------------------------------------------------------
// users (registration page + create/update endpoints)
// ---------------------------------------------------------------------------

func (h *Handler) registerUserPage(c echo.Context) error {
	return renderComponent(c, adminviews.RegisterUserPage())
}

func (h *Handler) registerUser(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	req := RegisterUserRequest{
		Name:        c.FormValue("name"),
		Email:       c.FormValue("email"),
		Nickname:    c.FormValue("nickname"),
		Password:    c.FormValue("password"),
		AvatarURL:   c.FormValue("avatar_url"),
		PhotoURL:    c.FormValue("photo_url"),
		StudentCode: c.FormValue("student_code"),
	}
	if v := c.FormValue("role_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.RoleID = id
		}
	}
	if v := c.FormValue("degree"); v != "" {
		d := models.DegreeType(v)
		req.Degree = &d
	}
	if v := c.FormValue("faculty"); v != "" {
		f := models.FacultyType(v)
		req.Faculty = &f
	}
	if v := c.FormValue("age"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.Age = &n
		}
	}
	if v := c.FormValue("sex"); v != "" {
		x := models.SexType(v)
		req.Sex = &x
	}
	if v := c.FormValue("nationality"); v != "" {
		n := models.NationalityType(v)
		req.Nationality = &n
	}
	if _, err := h.service.RegisterUser(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/tenants")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) updateUser(c echo.Context) error {
	actor, err := actorID(c)
	if err != nil {
		return h.writeError(c, err)
	}
	id, err := queryUUID(c, "id")
	if err != nil {
		return h.writeError(c, err)
	}
	req := UpdateUserRequest{
		ID:        id,
		Name:      optString(c.FormValue("name")),
		Email:     optString(c.FormValue("email")),
		Nickname:  optString(c.FormValue("nickname")),
		Password:  optString(c.FormValue("password")),
		AvatarURL: optString(c.FormValue("avatar_url")),
		PhotoURL:  optString(c.FormValue("photo_url")),
	}
	if v := c.FormValue("role_id"); v != "" {
		if rid, err := uuid.Parse(v); err == nil {
			req.RoleID = &rid
		}
	}
	if v := c.FormValue("is_active"); v != "" {
		b := strings.EqualFold(v, "true")
		req.IsActive = &b
	}
	if _, err := h.service.UpdateUser(actor, req); err != nil {
		return h.writeError(c, err)
	}
	c.Response().Header().Set("HX-Redirect", "/administration/tenants/"+id.String())
	return c.NoContent(http.StatusOK)
}

// ---------------------------------------------------------------------------
// audit log
// ---------------------------------------------------------------------------

func (h *Handler) auditPage(c echo.Context) error {
	filter := AuditLogFilter{
		TableName:  c.QueryParam("table_name"),
		Operation:  c.QueryParam("operation"),
		From:       parseDate(c.QueryParam("from")),
		To:         parseDate(c.QueryParam("to")),
		Pagination: parsePagination(c),
	}
	if v := c.QueryParam("changed_by"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ChangedBy = &id
		}
	}
	logs, err := h.service.ListAuditLogs(filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return renderComponent(c, adminviews.AuditPage(logs))
}
