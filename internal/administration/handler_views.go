package administration

import (
	"net/http"

	"dorm-man/internal/pagination"
	models "dorm-man/internal/models/administration"
	adminviews "dorm-man/web/templates/administration"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) tenantsPage(c echo.Context) error {
	params := pageParams(c)
	filter := TenantListFilter{
		Search: c.QueryParam("search"),
		Status: c.QueryParam("status"),
		Params: params,
	}
	tenants, total, err := h.service.ListTenants(filter)
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.TenantsPage(adminviews.TenantsPageData{
		Tenants:    tenants,
		Search:     filter.Search,
		Status:     filter.Status,
		Pagination: pagination.NewMeta(params, total),
	}))
}

func (h *Handler) tenantDetailPage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant id")
	}
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return renderComponent(c, adminviews.TenantDetailPage(adminviews.TenantDetailPageData{Tenant: tenant}))
}

func (h *Handler) roomsPage(c echo.Context) error {
	return h.roomsPageWithPlan(c, c.QueryParam("plan_json"))
}

func (h *Handler) roomsPageWithPlan(c echo.Context, planJSON string) error {
	params := pageParams(c)
	state := c.QueryParam("state")
	filter := RoomListFilter{
		State:  state,
		Search: c.QueryParam("search"),
		Params: params,
	}
	rooms, total, err := h.service.ListRooms(filter)
	if err != nil {
		return err
	}
	rows := make([]adminviews.RoomOverviewRow, 0, len(rooms))
	for _, room := range rooms {
		occ, err := h.service.RoomOccupancy(room.ID)
		if err != nil {
			return err
		}
		flag := "ok"
		for _, item := range room.InventoryItems {
			if item.Condition == models.InventoryConditionDamaged || item.Condition == models.InventoryConditionBroken {
				flag = "attention"
				break
			}
		}
		rows = append(rows, adminviews.RoomOverviewRow{
			Room:            room,
			Occupancy:       occ,
			InventoryStatus: flag,
		})
	}
	var pageMeta pagination.Meta
	if state != "" {
		rows, pageMeta = pagination.Slice(rows, params)
	} else {
		pageMeta = pagination.NewMeta(params, total)
	}
	tenants, _, err := h.service.ListTenants(TenantListFilter{Status: "active", Params: pagination.Unpaged()})
	if err != nil {
		return err
	}
	unassigned, err := h.service.ListUnassignedActiveTenants()
	if err != nil {
		return err
	}
	if planJSON == "" {
		planJSON = c.QueryParam("plan_json")
	}
	return renderComponent(c, adminviews.RoomsPage(adminviews.RoomsPageData{
		Rooms:      rows,
		State:      filter.State,
		Search:     filter.Search,
		Tenants:    tenants,
		Unassigned: unassigned,
		PlanJSON:   planJSON,
		Pagination: pageMeta,
	}))
}

func (h *Handler) roomDetailPage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}
	room, err := h.service.GetRoom(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	occ, err := h.service.RoomOccupancy(id)
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.RoomDetailPage(adminviews.RoomDetailPageData{
		Room:      room,
		Occupancy: occ,
	}))
}

func (h *Handler) inventoryPage(c echo.Context) error {
	params := pageParams(c)
	items, total, err := h.service.ListInventory(InventoryListFilter{Params: params})
	if err != nil {
		return err
	}
	rooms, _, err := h.service.ListRooms(RoomListFilter{Params: pagination.Unpaged()})
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.InventoryPage(adminviews.InventoryPageData{
		Items:      items,
		Rooms:      rooms,
		Pagination: pagination.NewMeta(params, total),
	}))
}

func (h *Handler) maintenancePage(c echo.Context) error {
	params := pageParams(c)
	filter := TicketListFilter{
		ApprovalState: c.QueryParam("approval_state"),
		Status:        c.QueryParam("status"),
		Params:        params,
	}
	tickets, total, err := h.service.ListMaintenanceTickets(filter)
	if err != nil {
		return err
	}
	staff, err := h.service.ListStaffUsers()
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.MaintenancePage(adminviews.MaintenancePageData{
		Tickets:       tickets,
		ApprovalState: filter.ApprovalState,
		Status:        filter.Status,
		StaffUsers:    staff,
		Pagination:    pagination.NewMeta(params, total),
	}))
}

func (h *Handler) jobsPage(c echo.Context) error {
	assigneeStr, dateStr, assignee, date, err := parseJobsFilter(c)
	if err != nil {
		return err
	}
	params := pageParams(c)
	jobs, total, err := h.service.ListOperationalJobs(JobListFilter{
		AssigneeUserID: assignee,
		Date:           date,
		Params:         params,
	})
	if err != nil {
		return err
	}
	staff, err := h.service.ListStaffUsers()
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.JobsPage(adminviews.JobsPageData{
		Jobs:           jobs,
		StaffUsers:     staff,
		AssigneeUserID: assigneeStr,
		Date:           dateStr,
		Pagination:     pagination.NewMeta(params, total),
	}))
}

func (h *Handler) publicationsPage(c echo.Context) error {
	newsParams := pageParamsNamed(c, "news_page", "news_page_size")
	actParams := pageParamsNamed(c, "activities_page", "activities_page_size")
	eventParams := pageParamsNamed(c, "events_page", "events_page_size")

	news, newsTotal, err := h.service.ListNews(PublicationListFilter{Params: newsParams})
	if err != nil {
		return err
	}
	activities, actTotal, err := h.service.ListActivities(PublicationListFilter{Params: actParams})
	if err != nil {
		return err
	}
	events, eventTotal, err := h.service.ListEvents(PublicationListFilter{Params: eventParams})
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.PublicationsPage(adminviews.PublicationsPageData{
		News:                 news,
		NewsPagination:       pagination.NewMeta(newsParams, newsTotal),
		NewsPreserve:         queryPreserve(c, "news_page", "news_page_size"),
		Activities:           activities,
		ActivitiesPagination: pagination.NewMeta(actParams, actTotal),
		ActivitiesPreserve:   queryPreserve(c, "activities_page", "activities_page_size"),
		Events:               events,
		EventsPagination:     pagination.NewMeta(eventParams, eventTotal),
		EventsPreserve:       queryPreserve(c, "events_page", "events_page_size"),
	}))
}

func (h *Handler) auditPage(c echo.Context) error {
	params := pageParams(c)
	events, total, err := h.service.ListAuditEvents(AuditListFilter{Params: params})
	if err != nil {
		return err
	}
	return renderComponent(c, adminviews.AuditPage(adminviews.AuditPageData{
		Events:     events,
		Pagination: pagination.NewMeta(params, total),
	}))
}
