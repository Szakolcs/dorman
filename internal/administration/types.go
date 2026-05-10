package administration

import (
	"errors"
	"time"

	models "dorm-man/internal/models/administration"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized         = errors.New("authorization denied")
	ErrNotFound             = errors.New("not found")
	ErrValidation           = errors.New("validation error")
	ErrCapacityConflict     = errors.New("capacity conflict")
	ErrStateTransition      = errors.New("state transition invalid")
	ErrConcurrencyConflict  = errors.New("concurrency conflict")
	ErrStudentStatusInvalid = errors.New("student status invalid")
)

type Principal struct {
	UserID uuid.UUID
	Roles  []models.RoleName
}

type TenantListFilter struct {
	Status string
	Search string
}

type RoomListFilter struct {
	State  string
	Search string
}

type TicketListFilter struct {
	ApprovalState string
	Status        string
}

type JobListFilter struct {
	AssigneeUserID *uuid.UUID
	Date           *time.Time
}

type NewsUpsertInput struct {
	Title       string
	Body        string
	Tags        string
	PublishDate *time.Time
	State       models.PublicationState
}

type ActivityUpsertInput struct {
	Title       string
	Description string
	BuildingID  *uuid.UUID
	Location    string
	Capacity    int
	State       models.PublicationState
}

type EventUpsertInput struct {
	Title           string
	Description     string
	BuildingID      *uuid.UUID
	OrganizerUserID uuid.UUID
	Location        string
	StartsAt        time.Time
	EndsAt          time.Time
	Capacity        int
	State           models.PublicationState
}

type AssignmentPlanItem struct {
	TenantID uuid.UUID `json:"tenant_id"`
	RoomID   uuid.UUID `json:"room_id"`
}

type AssignmentPlan struct {
	Items []AssignmentPlanItem `json:"items"`
}

type JobConflict struct {
	JobID    uuid.UUID `json:"job_id"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

type CreateJobResult struct {
	Job       models.OperationalJob `json:"job"`
	Conflicts []JobConflict         `json:"conflicts"`
}
