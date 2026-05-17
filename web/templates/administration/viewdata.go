package administration

import (
	adm "dorm-man/internal/models/administration"
	"dorm-man/internal/pagination"
	fm "dorm-man/internal/models/forum"
)

type DashboardStatsView struct {
	TenantsTotal      int64
	TenantsActive     int64
	TenantsInactive   int64
	TenantsUnassigned int64

	RoomsTotal             int64
	RoomsOccupied          int64
	RoomsAvailable         int64
	RoomsMaintenanceNeeded int64

	InventoryTotal     int64
	InventoryAttention int64

	MaintenanceTotal   int64
	MaintenancePending int64
	MaintenanceOpen    int64

	JobsTotal int64
	JobsToday int64

	NewsTotal       int64
	ActivitiesTotal int64
	EventsTotal     int64

	AuditTotal int64
}

type DashboardRoomMaintenanceRow struct {
	RoomID     string
	RoomNumber string
	TicketID   string
	Status     string
}

type DashboardOpenMaintenanceRow struct {
	TicketID   string
	Category   string
	Severity   string
	Status     string
	RoomNumber string
}

type DashboardJobTodayRow struct {
	JobID        string
	Title        string
	AssigneeName string
	StartsAt     string
	EndsAt       string
	Status       string
}

type DashboardPublicationRow struct {
	ID        string
	Kind      string
	Title     string
	State     string
	CreatedAt string
}

type DashboardEventRow struct {
	EventID   string
	Title     string
	StartsAt  string
	EndsAt    string
	State     string
}

type DashboardPageData struct {
	Stats                  DashboardStatsView
	JobsTodayDate          string
	RoomsOldestMaintenance []DashboardRoomMaintenanceRow
	OpenMaintenance        []DashboardOpenMaintenanceRow
	JobsToday              []DashboardJobTodayRow
	RecentPublications     []DashboardPublicationRow
	RecentEvents           []DashboardEventRow
}

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

type MaintenanceTicketDetailPageData struct {
	Ticket     adm.MaintenanceTicket
	StaffUsers []adm.User
}

type JobDetailPageData struct {
	Job adm.OperationalJob
}

type NewsDetailPageData struct {
	Post fm.ForumPost
}

type ActivityDetailPageData struct {
	Activity adm.Activity
}

type EventDetailPageData struct {
	Event adm.Event
}
