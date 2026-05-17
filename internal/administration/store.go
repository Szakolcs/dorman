package administration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"dorm-man/internal/pagination"
	models "dorm-man/internal/models/administration"
	forummodels "dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	LoadPrincipal(userID uuid.UUID) (Principal, error)
	ListStaffUsers() ([]models.User, error)

	CreateTenant(tenant models.Tenant) (models.Tenant, error)
	UpdateTenantStatus(tenantID uuid.UUID, active bool) (models.Tenant, error)
	ListTenants(filter TenantListFilter) ([]models.Tenant, int64, error)
	GetTenant(tenantID uuid.UUID) (models.Tenant, error)

	ListRooms(filter RoomListFilter) ([]models.Room, int64, error)
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
	ListInventory(filter InventoryListFilter) ([]models.InventoryItem, int64, error)

	CreateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error)
	GetMaintenanceTicket(ticketID uuid.UUID) (models.MaintenanceTicket, error)
	UpdateMaintenanceTicket(ticket models.MaintenanceTicket) (models.MaintenanceTicket, error)
	CreateTicketStatusChange(change models.TicketStatusChange) error
	ListMaintenanceTickets(filter TicketListFilter) ([]models.MaintenanceTicket, int64, error)

	ListJobConflicts(assigneeID uuid.UUID, startsAt, endsAt time.Time) ([]models.OperationalJob, error)
	CreateOperationalJob(job models.OperationalJob) (models.OperationalJob, error)
	ListOperationalJobs(filter JobListFilter) ([]models.OperationalJob, int64, error)

	CreateForumPost(post forummodels.ForumPost) (forummodels.ForumPost, error)
	UpdateForumPost(post forummodels.ForumPost) (forummodels.ForumPost, error)
	GetForumPost(id uuid.UUID) (forummodels.ForumPost, error)
	ListForumPosts(filter PublicationListFilter) ([]forummodels.ForumPost, int64, error)

	CreateActivity(activity models.Activity) (models.Activity, error)
	UpdateActivity(activity models.Activity) (models.Activity, error)
	GetActivity(id uuid.UUID) (models.Activity, error)
	ListActivities(filter PublicationListFilter) ([]models.Activity, int64, error)

	CreateEvent(event models.Event) (models.Event, error)
	UpdateEvent(event models.Event) (models.Event, error)
	GetEvent(id uuid.UUID) (models.Event, error)
	ListEvents(filter PublicationListFilter) ([]models.Event, int64, error)

	CreateAudit(event models.AuditEvent) error
	ListAuditEvents(filter AuditListFilter) ([]models.AuditEvent, int64, error)
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

func (s *GormStore) ListStaffUsers() ([]models.User, error) {
	staffRoles := []models.RoleName{
		models.RoleAdministrator,
		models.RoleOfficeWorker,
		models.RoleDirector,
		models.RoleDoorman,
	}
	var users []models.User
	err := s.db.
		Joins("JOIN user_roles ON user_roles.user_id = users.id AND user_roles.revoked_at IS NULL").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name IN ?", staffRoles).
		Where("users.is_active = ?", true).
		Distinct("users.*").
		Order("users.name ASC").
		Find(&users).Error
	return users, err
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

func (s *GormStore) ListTenants(filter TenantListFilter) ([]models.Tenant, int64, error) {
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
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tenants []models.Tenant
	if err := q.Order("name ASC").Scopes(pagination.Scope(filter.Params)).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
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

func (s *GormStore) ListRooms(filter RoomListFilter) ([]models.Room, int64, error) {
	q := s.db.Model(&models.Room{})
	if strings.TrimSpace(filter.Search) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
		q = q.Where("LOWER(number) LIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rooms []models.Room
	if err := q.Preload("InventoryItems").Order("number ASC").Scopes(pagination.Scope(filter.Params)).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
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

func (s *GormStore) ListInventory(filter InventoryListFilter) ([]models.InventoryItem, int64, error) {
	q := s.db.Model(&models.InventoryItem{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.InventoryItem
	err := q.Preload("Room").Preload("Flat").Preload("Building").
		Order("name ASC").
		Scopes(pagination.Scope(filter.Params)).
		Find(&items).Error
	return items, total, err
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

func (s *GormStore) ListMaintenanceTickets(filter TicketListFilter) ([]models.MaintenanceTicket, int64, error) {
	q := s.db.Model(&models.MaintenanceTicket{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.ApprovalState == "pending" {
		q = q.Where("status = ?", models.MaintenanceStatusReported)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tickets []models.MaintenanceTicket
	err := q.Preload("CreatedByUser").Preload("AssigneeUser").
		Order("created_at DESC").
		Scopes(pagination.Scope(filter.Params)).
		Find(&tickets).Error
	return tickets, total, err
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

func (s *GormStore) ListOperationalJobs(filter JobListFilter) ([]models.OperationalJob, int64, error) {
	q := s.db.Model(&models.OperationalJob{})
	if filter.AssigneeUserID != nil {
		q = q.Where("assignee_user_id = ?", *filter.AssigneeUserID)
	}
	if filter.Date != nil {
		start := time.Date(filter.Date.Year(), filter.Date.Month(), filter.Date.Day(), 0, 0, 0, 0, time.UTC)
		end := start.Add(24 * time.Hour)
		q = q.Where("starts_at < ? AND ends_at > ?", end, start)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var jobs []models.OperationalJob
	err := q.Preload("AssigneeUser").
		Order("starts_at ASC").
		Scopes(pagination.Scope(filter.Params)).
		Find(&jobs).Error
	return jobs, total, err
}

func (s *GormStore) CreateForumPost(post forummodels.ForumPost) (forummodels.ForumPost, error) {
	if err := s.db.Create(&post).Error; err != nil {
		return forummodels.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) UpdateForumPost(post forummodels.ForumPost) (forummodels.ForumPost, error) {
	if err := s.db.Save(&post).Error; err != nil {
		return forummodels.ForumPost{}, err
	}
	return post, nil
}

func (s *GormStore) GetForumPost(id uuid.UUID) (forummodels.ForumPost, error) {
	var post forummodels.ForumPost
	err := s.db.First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return forummodels.ForumPost{}, ErrNotFound
	}
	return post, err
}

func (s *GormStore) ListForumPosts(filter PublicationListFilter) ([]forummodels.ForumPost, int64, error) {
	q := s.db.Model(&forummodels.ForumPost{}).Where("kind = ?", forummodels.ForumPostKindOfficialNews)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var posts []forummodels.ForumPost
	err := q.Preload("AuthorUser").
		Order("created_at DESC").
		Scopes(pagination.Scope(filter.Params)).
		Find(&posts).Error
	return posts, total, err
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

func (s *GormStore) ListActivities(filter PublicationListFilter) ([]models.Activity, int64, error) {
	q := s.db.Model(&models.Activity{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var activities []models.Activity
	err := q.Order("created_at DESC").Scopes(pagination.Scope(filter.Params)).Find(&activities).Error
	return activities, total, err
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

func (s *GormStore) ListEvents(filter PublicationListFilter) ([]models.Event, int64, error) {
	q := s.db.Model(&models.Event{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []models.Event
	err := q.Preload("Organizer").Order("starts_at ASC").Scopes(pagination.Scope(filter.Params)).Find(&events).Error
	return events, total, err
}

func (s *GormStore) ListAuditEvents(filter AuditListFilter) ([]models.AuditEvent, int64, error) {
	q := s.db.Model(&models.AuditEvent{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []models.AuditEvent
	err := q.Preload("ActorUser").
		Order("occurred_at DESC").
		Scopes(pagination.Scope(filter.Params)).
		Find(&events).Error
	return events, total, err
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
