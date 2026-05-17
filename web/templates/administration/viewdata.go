package administration

import (
	adm "dorm-man/internal/models/administration"
	"dorm-man/internal/pagination"
	fm "dorm-man/internal/models/forum"
)

type TenantsPageData struct {
	Tenants    []adm.Tenant
	Search     string
	Status     string
	Pagination pagination.Meta
}

type TenantDetailPageData struct {
	Tenant adm.Tenant
}

type RoomOverviewRow struct {
	Room            adm.Room
	Occupancy       int64
	InventoryStatus string
}

type RoomsPageData struct {
	Rooms      []RoomOverviewRow
	State      string
	Search     string
	Tenants    []adm.Tenant
	Unassigned []adm.Tenant
	PlanJSON   string
	Pagination pagination.Meta
}

type RoomDetailPageData struct {
	Room      adm.Room
	Occupancy int64
}

type InventoryPageData struct {
	Items      []adm.InventoryItem
	Rooms      []adm.Room
	Pagination pagination.Meta
}

type MaintenancePageData struct {
	Tickets       []adm.MaintenanceTicket
	ApprovalState string
	Status        string
	StaffUsers    []adm.User
	Pagination    pagination.Meta
}

type JobsPageData struct {
	Jobs           []adm.OperationalJob
	StaffUsers     []adm.User
	AssigneeUserID string
	Date           string
	Pagination     pagination.Meta
}

type PublicationsPageData struct {
	News                   []fm.ForumPost
	NewsPagination         pagination.Meta
	NewsPreserve           map[string]string
	Activities             []adm.Activity
	ActivitiesPagination   pagination.Meta
	ActivitiesPreserve     map[string]string
	Events                 []adm.Event
	EventsPagination       pagination.Meta
	EventsPreserve         map[string]string
}

type AuditPageData struct {
	Events     []adm.AuditEvent
	Pagination pagination.Meta
}
