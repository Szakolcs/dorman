package administration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	models "dorm-man/internal/models/administration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	LoadPrincipal(userID uuid.UUID) (Principal, error)

	CreateTenant(tenant models.Tenant) (models.Tenant, error)
	UpdateTenantStatus(tenantID uuid.UUID, active bool) (models.Tenant, error)
	ListTenants(filter TenantListFilter) ([]models.Tenant, error)
	GetTenant(tenantID uuid.UUID) (models.Tenant, error)

	ListRooms(filter RoomListFilter) ([]models.Room, error)
	GetRoom(roomID uuid.UUID) (models.Room, error)
	GetActiveAssignmentCount(roomID uuid.UUID) (int64, error)
	CloseActiveAssignmentByTenant(tx *gorm.DB, tenantID uuid.UUID, endedAt time.Time) error
	CreateAssignment(tx *gorm.DB, assignment models.RoomAssignment) (models.RoomAssignment, error)
	WithTx(fn func(tx *gorm.DB) error) error
	ListActiveAssignmentsByRoom(roomID uuid.UUID) ([]models.RoomAssignment, error)
	ListUnassignedActiveTenants() ([]models.Tenant, error)
	ListAllRoomsForPlanning() ([]models.Room, error)

	CreateInventoryItem(item models.InventoryItem) (models.InventoryItem, error)
	UpdateInventoryStatus(id uuid.UUID, status models.InventoryStatus, condition models.InventoryCondition, withdrawDate *time.Time) (models.InventoryItem, error)
	ListInventory() ([]models.InventoryItem, error)

	CreateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error)
	GetMaintenanceTicket(ticketID uuid.UUID) (models.MaintenanceTicket, error)
	UpdateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error)
	CreateTicketStatusChange(change models.TicketStatusChange) error
	ListMaintenanceTickets(filter TicketListFilter) ([]models.MaintenanceTicket, error)

	ListJobConflicts(assigneeID uuid.UUID, startsAt, endsAt time.Time) ([]models.OperationalJob, error)
	CreateOperationalJob(job models.OperationalJob) (models.OperationalJob, error)
	ListOperationalJobs(filter JobListFilter) ([]models.OperationalJob, error)

	CreateForumPost(post models.ForumPost) (models.ForumPost, error)
	UpdateForumPost(post models.ForumPost) (models.ForumPost, error)
	GetForumPost(id uuid.UUID) (models.ForumPost, error)
	ListForumPosts() ([]models.ForumPost, error)

	CreateActivity(activity models.Activity) (models.Activity, error)
	UpdateActivity(activity models.Activity) (models.Activity, error)
	GetActivity(id uuid.UUID) (models.Activity, error)
	ListActivities() ([]models.Activity, error)

	CreateEvent(event models.Event) (models.Event, error)
	UpdateEvent(event models.Event) (models.Event, error)
	GetEvent(id uuid.UUID) (models.Event, error)
	ListEvents() ([]models.Event, error)

	CreateAudit(event models.AuditEvent) error
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) LoadPrincipal(userID uuid.UUID) (Principal, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Principal{}, ErrUnauthorized
		}
		return Principal{}, err
	}

	var roleRows []models.UserRole
	if err := s.db.Preload("Role").Where("user_id = ? AND revoked_at IS NULL", userID).Find(&roleRows).Error; err != nil {
		return Principal{}, err
	}

	roles := make([]models.RoleName, 0, len(roleRows))
	for _, row := range roleRows {
		roles = append(roles, row.Role.Name)
	}

	return Principal{UserID: userID, Roles: roles}, nil
}

func (s *GormStore) CreateTenant(tenant models.Tenant) (models.Tenant, error) {
	if err := s.db.Create(&tenant).Error; err != nil {
		return models.Tenant{}, err
	}
	return tenant, nil
}

func (s *GormStore) UpdateTenantStatus(tenantID uuid.UUID, active bool) (models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Tenant{}, ErrNotFound
		}
		return models.Tenant{}, err
	}
	tenant.IsActive = active
	if err := s.db.Save(&tenant).Error; err != nil {
		return models.Tenant{}, err
	}
	return tenant, nil
}

func (s *GormStore) ListTenants(filter TenantListFilter) ([]models.Tenant, error) {
	var tenants []models.Tenant
	q := s.db.Model(&models.Tenant{})
	switch strings.ToLower(filter.Status) {
	case "active":
		q = q.Where("is_active = true")
	case "inactive":
		q = q.Where("is_active = false")
	}
	if strings.TrimSpace(filter.Search) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(student_code) LIKE ? OR LOWER(email) LIKE ?", like, like, like)
	}
	if err := q.Order("name ASC").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *GormStore) GetTenant(tenantID uuid.UUID) (models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.Preload("RoomAssignments", "ended_at IS NULL").
		Preload("RoomAssignments.Room").
		First(&tenant, "id = ?", tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (s *GormStore) ListRooms(filter RoomListFilter) ([]models.Room, error) {
	var rooms []models.Room
	q := s.db.Model(&models.Room{}).Preload("InventoryItems")
	if strings.TrimSpace(filter.Search) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
		q = q.Where("LOWER(number) LIKE ?", like)
	}
	if err := q.Order("number ASC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *GormStore) GetRoom(roomID uuid.UUID) (models.Room, error) {
	var room models.Room
	err := s.db.Preload("Assignments", "ended_at IS NULL").
		Preload("Assignments.Tenant").
		Preload("InventoryItems").
		First(&room, "id = ?", roomID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Room{}, ErrNotFound
	}
	return room, err
}

func (s *GormStore) GetActiveAssignmentCount(roomID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.RoomAssignment{}).
		Where("room_id = ? AND ended_at IS NULL", roomID).
		Count(&count).Error
	return count, err
}

func (s *GormStore) CloseActiveAssignmentByTenant(tx *gorm.DB, tenantID uuid.UUID, endedAt time.Time) error {
	return tx.Model(&models.RoomAssignment{}).
		Where("tenant_id = ? AND ended_at IS NULL", tenantID).
		Updates(map[string]any{"ended_at": endedAt}).Error
}

func (s *GormStore) CreateAssignment(tx *gorm.DB, assignment models.RoomAssignment) (models.RoomAssignment, error) {
	if err := tx.Create(&assignment).Error; err != nil {
		return models.RoomAssignment{}, err
	}
	return assignment, nil
}

func (s *GormStore) WithTx(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

func (s *GormStore) ListActiveAssignmentsByRoom(roomID uuid.UUID) ([]models.RoomAssignment, error) {
	var assignments []models.RoomAssignment
	err := s.db.Where("room_id = ? AND ended_at IS NULL", roomID).
		Preload("Tenant").
		Find(&assignments).Error
	return assignments, err
}

func (s *GormStore) ListUnassignedActiveTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	sub := s.db.Model(&models.RoomAssignment{}).
		Select("tenant_id").
		Where("ended_at IS NULL")
	err := s.db.Where("is_active = true AND id NOT IN (?)", sub).
		Order("name ASC").
		Find(&tenants).Error
	return tenants, err
}

func (s *GormStore) ListAllRoomsForPlanning() ([]models.Room, error) {
	var rooms []models.Room
	err := s.db.Order("number ASC").Find(&rooms).Error
	return rooms, err
}

func (s *GormStore) CreateInventoryItem(item models.InventoryItem) (models.InventoryItem, error) {
	if err := s.db.Create(&item).Error; err != nil {
		return models.InventoryItem{}, err
	}
	return item, nil
}

func (s *GormStore) UpdateInventoryStatus(id uuid.UUID, status models.InventoryStatus, condition models.InventoryCondition, withdrawDate *time.Time) (models.InventoryItem, error) {
	var item models.InventoryItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.InventoryItem{}, ErrNotFound
		}
		return models.InventoryItem{}, err
	}
	item.Status = status
	item.Condition = condition
	item.WithdrawDate = withdrawDate
	if err := s.db.Save(&item).Error; err != nil {
		return models.InventoryItem{}, err
	}
	return item, nil
}

func (s *GormStore) ListInventory() ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Preload("Room").Preload("Flat").Preload("Building").
		Order("name ASC").
		Find(&items).Error
	return items, err
}

func (s *GormStore) CreateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error) {
	if err := s.db.Create(&ticket).Error; err != nil {
		return models.MaintenanceTicket{}, err
	}
	return ticket, nil
}

func (s *GormStore) GetMaintenanceTicket(ticketID uuid.UUID) (models.MaintenanceTicket, error) {
	var ticket models.MaintenanceTicket
	err := s.db.Preload("CreatedByUser").
		Preload("AssigneeUser").
		Preload("StatusTransitions").
		First(&ticket, "id = ?", ticketID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.MaintenanceTicket{}, ErrNotFound
	}
	return ticket, err
}

func (s *GormStore) UpdateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error) {
	if err := s.db.Save(&ticket).Error; err != nil {
		return models.MaintenanceTicket{}, err
	}
	return ticket, nil
}

func (s *GormStore) CreateTicketStatusChange(change models.TicketStatusChange) error {
	return s.db.Create(&change).Error
}

func (s *GormStore) ListMaintenanceTickets(filter TicketListFilter) ([]models.MaintenanceTicket, error) {
	var tickets []models.MaintenanceTicket
	q := s.db.Model(&models.MaintenanceTicket{}).Preload("CreatedByUser").Preload("AssigneeUser")
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.ApprovalState == "pending" {
		q = q.Where("status = ?", models.MaintenanceStatusReported)
	}
	err := q.Order("created_at DESC").Find(&tickets).Error
	return tickets, err
}

func (s *GormStore) ListJobConflicts(assigneeID uuid.UUID, startsAt, endsAt time.Time) ([]models.OperationalJob, error) {
	var jobs []models.OperationalJob
	err := s.db.Where("assignee_user_id = ? AND starts_at < ? AND ends_at > ?", assigneeID, endsAt, startsAt).
		Find(&jobs).Error
	return jobs, err
}

func (s *GormStore) CreateOperationalJob(job models.OperationalJob) (models.OperationalJob, error) {
	if err := s.db.Create(&job).Error; err != nil {
		return models.OperationalJob{}, err
	}
	return job, nil
}

func (s *GormStore) ListOperationalJobs(filter JobListFilter) ([]models.OperationalJob, error) {
	var jobs []models.OperationalJob
	q := s.db.Model(&models.OperationalJob{}).Preload("AssigneeUser")
	if filter.AssigneeUserID != nil {
		q = q.Where("assignee_user_id = ?", *filter.AssigneeUserID)
	}
	if filter.Date != nil {
		start := time.Date(filter.Date.Year(), filter.Date.Month(), filter.Date.Day(), 0, 0, 0, 0, time.UTC)
		end := start.Add(24 * time.Hour)
		q = q.Where("starts_at < ? AND ends_at > ?", end, start)
	}
	err := q.Order("starts_at ASC").Find(&jobs).Error
	return jobs, err
}

func (s *GormStore) CreateForumPost(post models.ForumPost) (models.ForumPost, error) {
	if err := s.db.Create(&post).Error; err != nil {
		return models.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) UpdateForumPost(post models.ForumPost) (models.ForumPost, error) {
	if err := s.db.Save(&post).Error; err != nil {
		return models.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) GetForumPost(id uuid.UUID) (models.ForumPost, error) {
	var post models.ForumPost
	err := s.db.First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ForumPost{}, ErrNotFound
	}
	return post, err
}

func (s *GormStore) ListForumPosts() ([]models.ForumPost, error) {
	var posts []models.ForumPost
	err := s.db.Order("created_at DESC").Find(&posts).Error
	return posts, err
}

func (s *GormStore) CreateActivity(activity models.Activity) (models.Activity, error) {
	if err := s.db.Create(&activity).Error; err != nil {
		return models.Activity{}, err
	}
	return activity, nil
}

func (s *GormStore) UpdateActivity(activity models.Activity) (models.Activity, error) {
	if err := s.db.Save(&activity).Error; err != nil {
		return models.Activity{}, err
	}
	return activity, nil
}

func (s *GormStore) GetActivity(id uuid.UUID) (models.Activity, error) {
	var activity models.Activity
	err := s.db.First(&activity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Activity{}, ErrNotFound
	}
	return activity, err
}

func (s *GormStore) ListActivities() ([]models.Activity, error) {
	var activities []models.Activity
	err := s.db.Order("created_at DESC").Find(&activities).Error
	return activities, err
}

func (s *GormStore) CreateEvent(event models.Event) (models.Event, error) {
	if err := s.db.Create(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (s *GormStore) UpdateEvent(event models.Event) (models.Event, error) {
	if err := s.db.Save(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (s *GormStore) GetEvent(id uuid.UUID) (models.Event, error) {
	var event models.Event
	err := s.db.First(&event, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Event{}, ErrNotFound
	}
	return event, err
}

func (s *GormStore) ListEvents() ([]models.Event, error) {
	var events []models.Event
	err := s.db.Order("starts_at ASC").Find(&events).Error
	return events, err
}

func (s *GormStore) CreateAudit(event models.AuditEvent) error {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if err := s.db.Create(&event).Error; err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}
