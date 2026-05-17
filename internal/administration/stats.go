package administration

import (
	"time"

	"github.com/google/uuid"

	models "dorm-man/internal/models/administration"
)

// DashboardStats holds aggregate counts for the administration dashboard.
type DashboardStats struct {
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

// RoomOldestAssignedTicket is a room paired with its oldest assigned maintenance ticket.
type RoomOldestAssignedTicket struct {
	RoomID     uuid.UUID
	RoomNumber string
	TicketID   uuid.UUID
	Status     models.MaintenanceStatus
}

// DashboardOpenMaintenanceTicket is an open maintenance ticket for dashboard display.
type DashboardOpenMaintenanceTicket struct {
	ID         uuid.UUID
	Category   models.MaintenanceCategory
	Severity   models.MaintenanceSeverity
	Status     models.MaintenanceStatus
	RoomNumber string
}

// DashboardJobToday is an operational job scheduled for the current UTC day.
type DashboardJobToday struct {
	ID           uuid.UUID
	Title        string
	AssigneeName string
	StartsAt     time.Time
	EndsAt       time.Time
	Status       models.JobStatus
}

// DashboardPublication is a recent news post or activity.
type DashboardPublication struct {
	ID        uuid.UUID
	Kind      string
	Title     string
	State     models.PublicationState
	CreatedAt time.Time
}

// DashboardEventRow is a recent dorm event.
type DashboardEventRow struct {
	ID       uuid.UUID
	Title    string
	StartsAt time.Time
	EndsAt   time.Time
	State    models.PublicationState
}
