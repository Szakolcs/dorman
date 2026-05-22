package administration

import (
	"dorm-man/internal/models"
	"errors"
	"time"

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

type Pagination struct {
	Page    int    `json:"page"     query:"page"`
	PerPage int    `json:"per_page" query:"per_page"`
	Sort    string `json:"sort"     query:"sort"`
	Order   string `json:"order"    query:"order"`
}

type TenantFilter struct {
	Search      string                  `json:"search"       query:"search"`
	Degree      *models.DegreeType      `json:"degree"       query:"degree"`
	Faculty     *models.FacultyType     `json:"faculty"      query:"faculty"`
	Sex         *models.SexType         `json:"sex"          query:"sex"`
	Nationality *models.NationalityType `json:"nationality"  query:"nationality"`
	IsActive    *bool                   `json:"is_active"    query:"is_active"`
	BuildingID  *uuid.UUID              `json:"building_id"  query:"building_id"`
	FlatID      *uuid.UUID              `json:"flat_id"      query:"flat_id"`
	RoomID      *uuid.UUID              `json:"room_id"      query:"room_id"`
	Pagination
}

type InventoryFilter struct {
	Search        string                     `json:"search"          query:"search"`
	Condition     *models.InventoryCondition `json:"condition"       query:"condition"`
	Status        *models.InventoryStatus    `json:"status"          query:"status"`
	BuildingID    *uuid.UUID                 `json:"building_id"     query:"building_id"`
	FlatID        *uuid.UUID                 `json:"flat_id"         query:"flat_id"`
	SharedAreaID  *uuid.UUID                 `json:"shared_area_id"  query:"shared_area_id"`
	RoomID        *uuid.UUID                 `json:"room_id"         query:"room_id"`
	PurchasedFrom *time.Time                 `json:"purchased_from"  query:"purchased_from"`
	PurchasedTo   *time.Time                 `json:"purchased_to"    query:"purchased_to"`
	Pagination
}

type JobsFilter struct {
	Search   string              `json:"search"   query:"search"`
	Priority *models.JobPriority `json:"priority" query:"priority"`
	Status   *models.JobStatus   `json:"status"   query:"status"`
	From     *time.Time          `json:"from"     query:"from"`
	To       *time.Time          `json:"to"       query:"to"`
	Pagination
}

type PublicationKind string

const (
	PublicationKindNews     PublicationKind = "news"
	PublicationKindActivity PublicationKind = "activity"
	PublicationKindEvent    PublicationKind = "event"
)

type PublicationFilter struct {
	Search string                   `json:"search" query:"search"`
	Kind   PublicationKind          `json:"kind"   query:"kind"`
	State  *models.PublicationState `json:"state"  query:"state"`
	From   *time.Time               `json:"from"   query:"from"`
	To     *time.Time               `json:"to"     query:"to"`
	Pagination
}

// ActivityStats holds booking counters for activity list rows.
type ActivityStats struct {
	Capacity    int `json:"capacity"`
	BookedCount int `json:"booked_count"`
}

// EventInterestStats holds attendance intent aggregates for event list rows.
type EventInterestStats struct {
	Interested    int `json:"interested"`
	NotInterested int `json:"not_interested"`
	Busy          int `json:"busy"`
}

// PublicationListItem is a publication row tagged with its forum kind for list UIs.
type PublicationListItem struct {
	models.Publication
	Kind     PublicationKind
	Activity *ActivityStats      `json:"activity,omitempty"`
	Event    *EventInterestStats `json:"event,omitempty"`
}

type BuildingFilter struct {
	Search string `json:"search" query:"search"`
	Pagination
}

type CreateBuildingRequest struct {
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}

type UpdateBuildingRequest struct {
	Name *string `json:"name,omitempty" form:"name"`
	Code *string `json:"code,omitempty" form:"code"`
}

type CreateFlatRequest struct {
	BuildingID uuid.UUID `json:"building_id" form:"building_id"`
	Name       string    `json:"name"         form:"name"`
	Floor      int       `json:"floor"        form:"floor"`
}

type CreateSharedAreaRequest struct {
	BuildingID uuid.UUID `json:"building_id" form:"building_id"`
	Name       string    `json:"name"         form:"name"`
	Code       string    `json:"code"         form:"code"`
}

type UpdateFlatRequest struct {
	Name  *string `json:"name,omitempty"  form:"name"`
	Floor *int    `json:"floor,omitempty" form:"floor"`
}

type UpdateSharedAreaRequest struct {
	Name *string `json:"name,omitempty" form:"name"`
	Code *string `json:"code,omitempty" form:"code"`
}

type CreateRoomRequest struct {
	FlatID   uuid.UUID `json:"flat_id"  form:"flat_id"`
	Number   string    `json:"number"   form:"number"`
	Capacity int       `json:"capacity" form:"capacity"`
}

type FlatFilter struct {
	Search     string     `json:"search"      query:"search"`
	BuildingID *uuid.UUID `json:"building_id" query:"building_id"`
	Floor      *int       `json:"floor"       query:"floor"`
	Pagination
}

type SharedAreaFilter struct {
	Search     string     `json:"search"      query:"search"`
	BuildingID *uuid.UUID `json:"building_id" query:"building_id"`
	Pagination
}

type RoomFilter struct {
	Search     string     `json:"search"        query:"search"`
	BuildingID *uuid.UUID `json:"building_id"   query:"building_id"`
	FlatID     *uuid.UUID `json:"flat_id"       query:"flat_id"`
	HasFreeBed *bool      `json:"has_free_bed"  query:"has_free_bed"`
	Pagination
}

type RoomAssignmentFilter struct {
	TenantID   *uuid.UUID `json:"tenant_id"   query:"tenant_id"`
	RoomID     *uuid.UUID `json:"room_id"     query:"room_id"`
	ActiveOnly *bool      `json:"active_only" query:"active_only"`
	From       *time.Time `json:"from"        query:"from"`
	To         *time.Time `json:"to"          query:"to"`
	Pagination
}

type AuditLogFilter struct {
	TableName string     `json:"table_name" query:"table_name"`
	Operation string     `json:"operation"  query:"operation"`
	ChangedBy *uuid.UUID `json:"changed_by" query:"changed_by"`
	From      *time.Time `json:"from"       query:"from"`
	To        *time.Time `json:"to"         query:"to"`
	Pagination
}

type CreateInventoryItemRequest struct {
	Name         string                    `json:"name"           form:"name"`
	Description  string                    `json:"description"    form:"description"`
	RoomID       *uuid.UUID                `json:"room_id"        form:"room_id"`
	FlatID       *uuid.UUID                `json:"flat_id"        form:"flat_id"`
	BuildingID   *uuid.UUID                `json:"building_id"    form:"building_id"`
	SharedAreaID *uuid.UUID                `json:"shared_area_id" form:"shared_area_id"`
	Condition    models.InventoryCondition `json:"condition"      form:"condition"`
	Status       models.InventoryStatus    `json:"status"         form:"status"`
	PurchaseDate time.Time                 `json:"purchase_date"  form:"purchase_date"`
}

type UpdateInventoryStatusRequest struct {
	ID        uuid.UUID                  `json:"id"                  param:"id"`
	Status    models.InventoryStatus     `json:"status"              form:"status"`
	Condition *models.InventoryCondition `json:"condition,omitempty" form:"condition"`
}

type CreateJobRequest struct {
	Title       string             `json:"title"       form:"title"`
	Description string             `json:"description" form:"description"`
	StartsAt    time.Time          `json:"starts_at"   form:"starts_at"`
	EndsAt      time.Time          `json:"ends_at"     form:"ends_at"`
	Priority    models.JobPriority `json:"priority"    form:"priority"`
	Status      models.JobStatus   `json:"status"      form:"status"`
}

type UpdateJobRequest struct {
	ID          uuid.UUID           `json:"id"                    form:"id"`
	Title       *string             `json:"title,omitempty"       form:"title"`
	Description *string             `json:"description,omitempty" form:"description"`
	StartsAt    *time.Time          `json:"starts_at,omitempty"   form:"starts_at"`
	EndsAt      *time.Time          `json:"ends_at,omitempty"     form:"ends_at"`
	Priority    *models.JobPriority `json:"priority,omitempty"    form:"priority"`
	Status      *models.JobStatus   `json:"status,omitempty"      form:"status"`
}

type DeleteJobRequest struct {
	ID uuid.UUID `json:"id" query:"id" form:"id"`
}

type CreateNewsRequest struct {
	Title       string                  `json:"title"       form:"title"`
	Description string                  `json:"description" form:"description"`
	State       models.PublicationState `json:"state"       form:"state"`
}

type CreateActivityRequest struct {
	Title        string                  `json:"title"          form:"title"`
	Description  string                  `json:"description"    form:"description"`
	State        models.PublicationState `json:"state"          form:"state"`
	SharedAreaID *uuid.UUID              `json:"shared_area_id" form:"shared_area_id"`
	Capacity     int                     `json:"capacity"       form:"capacity"`
}

type CreateEventRequest struct {
	Title        string                  `json:"title"          form:"title"`
	Description  string                  `json:"description"    form:"description"`
	State        models.PublicationState `json:"state"          form:"state"`
	SharedAreaID *uuid.UUID              `json:"shared_area_id" form:"shared_area_id"`
	StartsAt     time.Time               `json:"starts_at"      form:"starts_at"`
	EndsAt       time.Time               `json:"ends_at"        form:"ends_at"`
}

type UpdatePublicationStateRequest struct {
	State models.PublicationState `json:"state" form:"state"`
}

type AssignRoomRequest struct {
	TenantID    uuid.UUID  `json:"tenant_id"              form:"tenant_id"`
	RoomID      uuid.UUID  `json:"room_id"                form:"room_id"`
	EffectiveAt *time.Time `json:"effective_at,omitempty" form:"effective_at"`
}

type MassAssignRequest struct {
	Assignments []AssignRoomRequest `json:"assignments"`
}

type UpdateAssignmentRequest struct {
	ID          uuid.UUID  `json:"id"                     form:"id"`
	RoomID      *uuid.UUID `json:"room_id,omitempty"      form:"room_id"`
	EffectiveAt *time.Time `json:"effective_at,omitempty" form:"effective_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"     form:"ended_at"`
}

type DeleteAssignmentRequest struct {
	ID uuid.UUID `json:"id" query:"id" form:"id"`
}

type TenantMassAssignmentRequest struct {
	StrictGroups string `json:"strict_groups" form:"strict_groups"`
	Preferences  string `json:"preferences"  form:"preferences"`
}

type RegisterUserRequest struct {
	Name      string    `json:"name"                form:"name"`
	Email     string    `json:"email"               form:"email"`
	Nickname  string    `json:"nickname"            form:"nickname"`
	Password  string    `json:"password"            form:"password"`
	AvatarURL string    `json:"avatar_url,omitempty" form:"avatar_url"`
	PhotoURL  string    `json:"photo_url,omitempty"  form:"photo_url"`
	RoleID    uuid.UUID `json:"role_id"             form:"role_id"`

	StudentCode string                  `json:"student_code,omitempty" form:"student_code"`
	Degree      *models.DegreeType      `json:"degree,omitempty"       form:"degree"`
	Faculty     *models.FacultyType     `json:"faculty,omitempty"      form:"faculty"`
	Age         *int                    `json:"age,omitempty"          form:"age"`
	Sex         *models.SexType         `json:"sex,omitempty"          form:"sex"`
	Nationality *models.NationalityType `json:"nationality,omitempty"  form:"nationality"`
}

type UpdateUserRequest struct {
	ID        uuid.UUID  `json:"id"                   form:"id"`
	Name      *string    `json:"name,omitempty"       form:"name"`
	Email     *string    `json:"email,omitempty"      form:"email"`
	Nickname  *string    `json:"nickname,omitempty"   form:"nickname"`
	Password  *string    `json:"password,omitempty"   form:"password"`
	AvatarURL *string    `json:"avatar_url,omitempty" form:"avatar_url"`
	PhotoURL  *string    `json:"photo_url,omitempty"  form:"photo_url"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"    form:"role_id"`

	StudentCode *string                 `json:"student_code,omitempty" form:"student_code"`
	Degree      *models.DegreeType      `json:"degree,omitempty"       form:"degree"`
	Faculty     *models.FacultyType     `json:"faculty,omitempty"      form:"faculty"`
	Age         *int                    `json:"age,omitempty"          form:"age"`
	Sex         *models.SexType         `json:"sex,omitempty"          form:"sex"`
	Nationality *models.NationalityType `json:"nationality,omitempty"  form:"nationality"`
	IsActive    *bool                   `json:"is_active,omitempty"    form:"is_active"`
}

type SuccessResponse struct {
	Success bool      `json:"success"`
	ID      uuid.UUID `json:"id,omitempty"`
}

type ListResponse[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}
