package administration

import (
	models2 "dorm-man/internal/models"
	"errors"
	"strings"

	"dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditedTransaction runs fn inside a database transaction with the
// `app.current_app_user` PostgreSQL session variable set to the supplied
// userID, so that DB-level audit triggers can attribute the change.
func AuditedTransaction(db *gorm.DB, userID uuid.UUID, fn func(tx *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL app.current_app_user = ?", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}

type Store interface {
	getTenantsAll() ([]models2.Tenant, error)
	getTenants(filter TenantFilter) ([]models2.Tenant, error)
	getTenantByID(id uuid.UUID) (models2.Tenant, error)

	getInventory(filter InventoryFilter) ([]models2.InventoryItem, error)
	getInventoryByID(id uuid.UUID) (models2.InventoryItem, error)
	getInventoryByFlatID(flatID uuid.UUID) ([]models2.InventoryItem, error)
	getInventoryByBuildingID(buildingID uuid.UUID) ([]models2.InventoryItem, error)
	getInventoryBySharedAreaID(sharedAreaID uuid.UUID) ([]models2.InventoryItem, error)
	createInventoryItem(actorID uuid.UUID, item models2.InventoryItem) error
	updateInventoryItem(actorID uuid.UUID, item models2.InventoryItem) error
	deleteInventoryItem(actorID, id uuid.UUID) error

	getJobs(filter JobsFilter) ([]models2.OperationalJob, error)
	getJobByID(id uuid.UUID) (models2.OperationalJob, error)
	createJob(actorID uuid.UUID, job models2.OperationalJob) error
	updateJob(actorID uuid.UUID, job models2.OperationalJob) error
	deleteJob(actorID, id uuid.UUID) error

	getPublications(filter PublicationFilter) ([]models2.Publication, error)
	getNewsByID(id uuid.UUID) (models2.Publication, error)
	createPublication(actorID uuid.UUID, pub models2.Publication) error
	updatePublication(actorID uuid.UUID, pub models2.Publication) error
	deletePublication(actorID, id uuid.UUID) error
	archiveNews(actorID, id uuid.UUID) error

	getActivityByID(id uuid.UUID) (models2.Activity, error)
	createActivity(actorID uuid.UUID, a models2.Activity) error
	updateActivity(actorID uuid.UUID, a models2.Activity) error
	archiveActivity(actorID, id uuid.UUID) error

	getEventByID(id uuid.UUID) (models2.Event, error)
	createEvent(actorID uuid.UUID, e models2.Event) error
	updateEvent(actorID uuid.UUID, e models2.Event) error
	archiveEvent(actorID, id uuid.UUID) error

	getBuilding(filter BuildingFilter) ([]models2.Building, error)
	getBuildingByID(id uuid.UUID) (models2.Building, error)
	createBuilding(actorID uuid.UUID, building models2.Building) error
	updateBuilding(actorID uuid.UUID, building models2.Building) error
	deleteBuilding(actorID, id uuid.UUID) error

	getFlat(filter FlatFilter) ([]models2.Flat, error)
	getFlatByID(id uuid.UUID) (models2.Flat, error)
	createFlat(actorID uuid.UUID, flat models2.Flat) error
	updateFlat(actorID uuid.UUID, flat models2.Flat) error
	deleteFlat(actorID, id uuid.UUID) error

	getSharedArea(filter SharedAreaFilter) ([]models2.SharedArea, error)
	getSharedAreaByID(id uuid.UUID) (models2.SharedArea, error)
	createSharedArea(actorID uuid.UUID, sharedArea models2.SharedArea) error
	updateSharedArea(actorID uuid.UUID, sharedArea models2.SharedArea) error
	deleteSharedArea(actorID, id uuid.UUID) error

	getRoom(filter RoomFilter) ([]models2.Room, error)
	getRoomByID(id uuid.UUID) (models2.Room, error)
	createRoom(actorID uuid.UUID, room models2.Room) error
	updateRoom(actorID uuid.UUID, room models2.Room) error
	deleteRoom(actorID, id uuid.UUID) error

	getRoomAssigment(filter RoomAssignmentFilter) ([]models2.RoomAssignment, error)
	getRoomAssignmentByID(id uuid.UUID) (models2.RoomAssignment, error)
	createRoomAssignment(actorID uuid.UUID, assignment models2.RoomAssignment) error
	updateRoomAssignment(actorID uuid.UUID, assignment models2.RoomAssignment) error
	deleteRoomAssignment(actorID, id uuid.UUID) error

	registerUser(actorID uuid.UUID, user models2.User) error
	updateUser(actorID uuid.UUID, user models2.User) error

	getAuditLogs(filter AuditLogFilter) ([]models2.Audit, error)
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func likePattern(s string) string {
	return "%" + strings.ToLower(strings.TrimSpace(s)) + "%"
}

// applyPagination applies sane defaults (page 1, 25 per page, capped at 200)
// when the caller leaves the embedded Pagination zero-valued.
func applyPagination(q *gorm.DB, p Pagination) *gorm.DB {
	page := p.Page
	if page < 1 {
		page = 1
	}
	per := p.PerPage
	switch {
	case per <= 0:
		per = 25
	case per > 200:
		per = 200
	}
	return q.Limit(per).Offset((page - 1) * per)
}

func applyOrder(q *gorm.DB, p Pagination, defaultCol string) *gorm.DB {
	col := strings.TrimSpace(p.Sort)
	if col == "" {
		col = defaultCol
	}
	order := strings.ToLower(p.Order)
	if order != "asc" {
		order = "desc"
	}
	return q.Order(col + " " + order)
}

// ---------------------------------------------------------------------------
// tenants
// ---------------------------------------------------------------------------

func (s *GormStore) getTenantsAll() ([]models2.Tenant, error) {
	var tenants []models2.Tenant
	err := s.db.
		Preload("User").
		Preload("RoomAssignments").
		Order("registered_at DESC").
		Find(&tenants).Error
	return tenants, err
}

func (s *GormStore) getTenants(filter TenantFilter) ([]models2.Tenant, error) {
	q := s.db.Model(&models2.Tenant{}).
		Preload("User").
		Preload("RoomAssignments")

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Joins("LEFT JOIN users ON users.id = tenants.user_id").
			Where(
				"LOWER(tenants.student_code) LIKE ? OR LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?",
				like, like, like,
			)
	}
	if filter.Degree != nil {
		q = q.Where("tenants.degree = ?", *filter.Degree)
	}
	if filter.Faculty != nil {
		q = q.Where("tenants.faculty = ?", *filter.Faculty)
	}
	if filter.Sex != nil {
		q = q.Where("tenants.sex = ?", *filter.Sex)
	}
	if filter.Nationality != nil {
		q = q.Where("tenants.nationality = ?", *filter.Nationality)
	}
	if filter.IsActive != nil {
		q = q.Where("tenants.is_active = ?", *filter.IsActive)
	}

	switch {
	case filter.RoomID != nil:
		q = q.Joins("JOIN room_assignments ra ON ra.tenant_id = tenants.id AND ra.ended_at IS NULL").
			Where("ra.room_id = ?", *filter.RoomID)
	case filter.FlatID != nil:
		q = q.Joins("JOIN room_assignments ra ON ra.tenant_id = tenants.id AND ra.ended_at IS NULL").
			Joins("JOIN rooms r ON r.id = ra.room_id").
			Where("r.flat_id = ?", *filter.FlatID)
	case filter.BuildingID != nil:
		q = q.Joins("JOIN room_assignments ra ON ra.tenant_id = tenants.id AND ra.ended_at IS NULL").
			Joins("JOIN rooms r ON r.id = ra.room_id").
			Joins("JOIN flats f ON f.id = r.flat_id").
			Where("f.building_id = ?", *filter.BuildingID)
	}

	q = applyOrder(q, filter.Pagination, "tenants.registered_at")
	q = applyPagination(q, filter.Pagination)

	var tenants []models2.Tenant
	err := q.Find(&tenants).Error
	return tenants, err
}

// ---------------------------------------------------------------------------
// inventory
// ---------------------------------------------------------------------------

func (s *GormStore) getInventory(filter InventoryFilter) ([]models2.InventoryItem, error) {
	q := s.db.Model(&models2.InventoryItem{}).
		Preload("Room").
		Preload("Flat").
		Preload("Building").
		Preload("SharedArea")

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if filter.Condition != nil {
		q = q.Where("condition = ?", *filter.Condition)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.RoomID != nil {
		q = q.Where("room_id = ?", *filter.RoomID)
	}
	if filter.FlatID != nil {
		q = q.Where("flat_id = ?", *filter.FlatID)
	}
	if filter.BuildingID != nil {
		q = q.Where("building_id = ?", *filter.BuildingID)
	}
	if filter.SharedAreaID != nil {
		q = q.Where("shared_area_id = ?", *filter.SharedAreaID)
	}
	if filter.PurchasedFrom != nil {
		q = q.Where("purchase_date >= ?", *filter.PurchasedFrom)
	}
	if filter.PurchasedTo != nil {
		q = q.Where("purchase_date <= ?", *filter.PurchasedTo)
	}

	q = applyOrder(q, filter.Pagination, "purchase_date")
	q = applyPagination(q, filter.Pagination)

	var items []models2.InventoryItem
	err := q.Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryByID(id uuid.UUID) (models2.InventoryItem, error) {
	var item models2.InventoryItem
	err := s.db.
		Preload("Room").
		Preload("Flat").
		Preload("Building").
		Preload("SharedArea").
		First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.InventoryItem{}, ErrNotFound
	}
	return item, err
}

func (s *GormStore) getInventoryByFlatID(flatID uuid.UUID) ([]models2.InventoryItem, error) {
	var items []models2.InventoryItem
	err := s.db.Where("flat_id = ?", flatID).Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryByBuildingID(buildingID uuid.UUID) ([]models2.InventoryItem, error) {
	var items []models2.InventoryItem
	err := s.db.Where("building_id = ?", buildingID).Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryBySharedAreaID(sharedAreaID uuid.UUID) ([]models2.InventoryItem, error) {
	var items []models2.InventoryItem
	err := s.db.Where("shared_area_id = ?", sharedAreaID).Find(&items).Error
	return items, err
}

func (s *GormStore) createInventoryItem(actorID uuid.UUID, item models2.InventoryItem) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&item).Error
	})
}

func (s *GormStore) updateInventoryItem(actorID uuid.UUID, item models2.InventoryItem) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&item).Error
	})
}

func (s *GormStore) deleteInventoryItem(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.InventoryItem{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// operational jobs
// ---------------------------------------------------------------------------

func (s *GormStore) getJobs(filter JobsFilter) ([]models2.OperationalJob, error) {
	q := s.db.Model(&models2.OperationalJob{})

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if filter.Priority != nil {
		q = q.Where("priority = ?", *filter.Priority)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.From != nil {
		q = q.Where("ends_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("starts_at <= ?", *filter.To)
	}

	q = applyOrder(q, filter.Pagination, "starts_at")
	q = applyPagination(q, filter.Pagination)

	var jobs []models2.OperationalJob
	err := q.Find(&jobs).Error
	return jobs, err
}

func (s *GormStore) createJob(actorID uuid.UUID, job models2.OperationalJob) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&job).Error
	})
}

func (s *GormStore) updateJob(actorID uuid.UUID, job models2.OperationalJob) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&job).Error
	})
}

func (s *GormStore) deleteJob(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.OperationalJob{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// publications
//
// forum.Activity and forum.Event each embed forum.Publication as a base.
// There is no unified `publications` table in the current migration set, so
// getPublications queries the `activities` / `events` tables directly and
// projects only the shared base columns (id/title/state/timestamps) into
// forum.Publication.
//
// createPublication / updatePublication / deletePublication operate on the
// hypothetical `publications` table; if/when a stand-alone Publication model
// gets migrated they will work as-is.
// ---------------------------------------------------------------------------

func (s *GormStore) getPublications(filter PublicationFilter) ([]models2.Publication, error) {
	fetch := func(table string) ([]models2.Publication, error) {
		q := s.db.Table(table).
			Select("id, created_at, updated_at, deleted_at, title, description, state")

		if filter.Search != "" {
			q = q.Where("LOWER(title) LIKE ?", likePattern(filter.Search))
		}
		if filter.State != nil {
			q = q.Where("state = ?", *filter.State)
		}
		if filter.From != nil {
			q = q.Where("created_at >= ?", *filter.From)
		}
		if filter.To != nil {
			q = q.Where("created_at <= ?", *filter.To)
		}

		q = applyOrder(q, filter.Pagination, "created_at")

		var rows []models2.Publication
		err := q.Scan(&rows).Error
		return rows, err
	}

	var out []models2.Publication
	switch filter.Kind {
	case PublicationKindActivity:
		rows, err := fetch("activities")
		if err != nil {
			return nil, err
		}
		out = rows
	case PublicationKindEvent:
		rows, err := fetch("events")
		if err != nil {
			return nil, err
		}
		out = rows
	case PublicationKindNews:
		// no News model is migrated yet; return an empty slice so callers
		// can still distinguish "kind known, none found" from an error.
		return []models2.Publication{}, nil
	default:
		acts, err := fetch("activities")
		if err != nil {
			return nil, err
		}
		evts, err := fetch("events")
		if err != nil {
			return nil, err
		}
		out = append(acts, evts...)
	}

	// Apply pagination at the application level since the union spans tables.
	page := filter.Pagination.Page
	if page < 1 {
		page = 1
	}
	per := filter.Pagination.PerPage
	switch {
	case per <= 0:
		per = 25
	case per > 200:
		per = 200
	}
	start := (page - 1) * per
	if start >= len(out) {
		return []models2.Publication{}, nil
	}
	end := start + per
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], nil
}

func (s *GormStore) createPublication(actorID uuid.UUID, pub models2.Publication) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&pub).Error
	})
}

func (s *GormStore) updatePublication(actorID uuid.UUID, pub models2.Publication) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&pub).Error
	})
}

func (s *GormStore) deletePublication(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.Publication{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// buildings
// ---------------------------------------------------------------------------

func (s *GormStore) getBuilding(filter BuildingFilter) ([]models2.Building, error) {
	q := s.db.Model(&models2.Building{}).
		Preload("Flats").
		Preload("SharedAreas")

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}

	q = applyOrder(q, filter.Pagination, "name")
	q = applyPagination(q, filter.Pagination)

	var bs []models2.Building
	err := q.Find(&bs).Error
	return bs, err
}

func (s *GormStore) createBuilding(actorID uuid.UUID, b models2.Building) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&b).Error
	})
}

func (s *GormStore) updateBuilding(actorID uuid.UUID, b models2.Building) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&b).Error
	})
}

func (s *GormStore) deleteBuilding(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.Building{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// flats
// ---------------------------------------------------------------------------

func (s *GormStore) getFlat(filter FlatFilter) ([]models2.Flat, error) {
	q := s.db.Model(&models2.Flat{}).
		Preload("Building").
		Preload("Rooms").
		Preload("Inventory")

	if filter.Search != "" {
		q = q.Where("LOWER(name) LIKE ?", likePattern(filter.Search))
	}
	if filter.BuildingID != nil {
		q = q.Where("building_id = ?", *filter.BuildingID)
	}
	if filter.Floor != nil {
		q = q.Where("floor = ?", *filter.Floor)
	}

	q = applyOrder(q, filter.Pagination, "name")
	q = applyPagination(q, filter.Pagination)

	var flats []models2.Flat
	err := q.Find(&flats).Error
	return flats, err
}

func (s *GormStore) createFlat(actorID uuid.UUID, f models2.Flat) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&f).Error
	})
}

func (s *GormStore) updateFlat(actorID uuid.UUID, f models2.Flat) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&f).Error
	})
}

func (s *GormStore) deleteFlat(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.Flat{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// shared areas
// ---------------------------------------------------------------------------

func (s *GormStore) getSharedArea(filter SharedAreaFilter) ([]models2.SharedArea, error) {
	q := s.db.Model(&models2.SharedArea{}).
		Preload("Building").
		Preload("Inventory")

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}
	if filter.BuildingID != nil {
		q = q.Where("building_id = ?", *filter.BuildingID)
	}

	q = applyOrder(q, filter.Pagination, "name")
	q = applyPagination(q, filter.Pagination)

	var sas []models2.SharedArea
	err := q.Find(&sas).Error
	return sas, err
}

func (s *GormStore) createSharedArea(actorID uuid.UUID, sa models2.SharedArea) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&sa).Error
	})
}

func (s *GormStore) updateSharedArea(actorID uuid.UUID, sa models2.SharedArea) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&sa).Error
	})
}

func (s *GormStore) deleteSharedArea(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.SharedArea{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// rooms
// ---------------------------------------------------------------------------

func (s *GormStore) getRoom(filter RoomFilter) ([]models2.Room, error) {
	q := s.db.Model(&models2.Room{}).
		Preload("Flat").
		Preload("Flat.Building").
		Preload("Assignments", "ended_at IS NULL")

	if filter.Search != "" {
		q = q.Where("LOWER(rooms.number) LIKE ?", likePattern(filter.Search))
	}
	if filter.FlatID != nil {
		q = q.Where("rooms.flat_id = ?", *filter.FlatID)
	}
	if filter.BuildingID != nil {
		q = q.Joins("JOIN flats ON flats.id = rooms.flat_id").
			Where("flats.building_id = ?", *filter.BuildingID)
	}
	if filter.HasFreeBed != nil && *filter.HasFreeBed {
		q = q.Where(
			"rooms.capacity > (SELECT COUNT(*) FROM room_assignments ra " +
				"WHERE ra.room_id = rooms.id AND ra.ended_at IS NULL)",
		)
	}

	q = applyOrder(q, filter.Pagination, "rooms.number")
	q = applyPagination(q, filter.Pagination)

	var rooms []models2.Room
	err := q.Find(&rooms).Error
	return rooms, err
}

func (s *GormStore) createRoom(actorID uuid.UUID, r models2.Room) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&r).Error
	})
}

func (s *GormStore) updateRoom(actorID uuid.UUID, r models2.Room) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&r).Error
	})
}

func (s *GormStore) deleteRoom(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.Room{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// room assignments
// ---------------------------------------------------------------------------

func (s *GormStore) getRoomAssigment(filter RoomAssignmentFilter) ([]models2.RoomAssignment, error) {
	q := s.db.Model(&models2.RoomAssignment{}).
		Preload("Tenant").
		Preload("Tenant.User").
		Preload("Room").
		Preload("Room.Flat")

	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.RoomID != nil {
		q = q.Where("room_id = ?", *filter.RoomID)
	}
	if filter.ActiveOnly != nil && *filter.ActiveOnly {
		q = q.Where("ended_at IS NULL")
	}
	if filter.From != nil {
		q = q.Where("effective_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("effective_at <= ?", *filter.To)
	}

	q = applyOrder(q, filter.Pagination, "effective_at")
	q = applyPagination(q, filter.Pagination)

	var as []models2.RoomAssignment
	err := q.Find(&as).Error
	return as, err
}

func (s *GormStore) createRoomAssignment(actorID uuid.UUID, a models2.RoomAssignment) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&a).Error
	})
}

func (s *GormStore) updateRoomAssignment(actorID uuid.UUID, a models2.RoomAssignment) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&a).Error
	})
}

func (s *GormStore) deleteRoomAssignment(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models2.RoomAssignment{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// users
// ---------------------------------------------------------------------------

func (s *GormStore) registerUser(actorID uuid.UUID, user models2.User) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&user).Error
	})
}

func (s *GormStore) updateUser(actorID uuid.UUID, user models2.User) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&user).Error
	})
}

// ---------------------------------------------------------------------------
// by-id helpers
// ---------------------------------------------------------------------------

func (s *GormStore) getTenantByID(id uuid.UUID) (models2.Tenant, error) {
	var tenant models2.Tenant
	err := s.db.
		Preload("User").
		Preload("RoomAssignments").
		Preload("RoomAssignments.Room").
		Preload("RoomAssignments.Room.Flat").
		First(&tenant, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (s *GormStore) getJobByID(id uuid.UUID) (models2.OperationalJob, error) {
	var job models2.OperationalJob
	err := s.db.First(&job, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.OperationalJob{}, ErrNotFound
	}
	return job, err
}

func (s *GormStore) getBuildingByID(id uuid.UUID) (models2.Building, error) {
	var b models2.Building
	err := s.db.
		Preload("Flats").
		Preload("SharedAreas").
		First(&b, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Building{}, ErrNotFound
	}
	return b, err
}

func (s *GormStore) getFlatByID(id uuid.UUID) (models2.Flat, error) {
	var f models2.Flat
	err := s.db.
		Preload("Building").
		Preload("Rooms").
		Preload("Rooms.Assignments", "ended_at IS NULL").
		Preload("Inventory").
		First(&f, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Flat{}, ErrNotFound
	}
	return f, err
}

func (s *GormStore) getSharedAreaByID(id uuid.UUID) (models2.SharedArea, error) {
	var sa models2.SharedArea
	err := s.db.
		Preload("Building").
		Preload("Inventory").
		First(&sa, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.SharedArea{}, ErrNotFound
	}
	return sa, err
}

func (s *GormStore) getRoomByID(id uuid.UUID) (models2.Room, error) {
	var r models2.Room
	err := s.db.
		Preload("Flat").
		Preload("Flat.Building").
		Preload("Assignments", "ended_at IS NULL").
		Preload("Assignments.Tenant").
		Preload("Assignments.Tenant.User").
		First(&r, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Room{}, ErrNotFound
	}
	return r, err
}

func (s *GormStore) getRoomAssignmentByID(id uuid.UUID) (models2.RoomAssignment, error) {
	var a models2.RoomAssignment
	err := s.db.
		Preload("Tenant").
		Preload("Tenant.User").
		Preload("Room").
		Preload("Room.Flat").
		First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.RoomAssignment{}, ErrNotFound
	}
	return a, err
}

// ---------------------------------------------------------------------------
// news (forum.Publication) detail + archive
// ---------------------------------------------------------------------------

func (s *GormStore) getNewsByID(id uuid.UUID) (models2.Publication, error) {
	var p models2.Publication
	err := s.db.Preload("Author").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Publication{}, ErrNotFound
	}
	return p, err
}

func (s *GormStore) archiveNews(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models2.Publication{}).
			Where("id = ?", id).
			Update("state", forum.PublicationStateArchived)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// activities
// ---------------------------------------------------------------------------

func (s *GormStore) getActivityByID(id uuid.UUID) (models2.Activity, error) {
	var a models2.Activity
	err := s.db.
		Preload("Author").
		Preload("Bookings").
		First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Activity{}, ErrNotFound
	}
	return a, err
}

func (s *GormStore) createActivity(actorID uuid.UUID, a models2.Activity) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&a).Error
	})
}

func (s *GormStore) updateActivity(actorID uuid.UUID, a models2.Activity) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&a).Error
	})
}

func (s *GormStore) archiveActivity(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models2.Activity{}).
			Where("id = ?", id).
			Update("state", forum.PublicationStateArchived)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// events
// ---------------------------------------------------------------------------

func (s *GormStore) getEventByID(id uuid.UUID) (models2.Event, error) {
	var e models2.Event
	err := s.db.
		Preload("Author").
		Preload("Attendances").
		First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models2.Event{}, ErrNotFound
	}
	return e, err
}

func (s *GormStore) createEvent(actorID uuid.UUID, e models2.Event) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&e).Error
	})
}

func (s *GormStore) updateEvent(actorID uuid.UUID, e models2.Event) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&e).Error
	})
}

func (s *GormStore) archiveEvent(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models2.Event{}).
			Where("id = ?", id).
			Update("state", forum.PublicationStateArchived)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// audit
// ---------------------------------------------------------------------------

func (s *GormStore) getAuditLogs(filter AuditLogFilter) ([]models2.Audit, error) {
	q := s.db.Model(&models2.Audit{}).Preload("User")

	if filter.TableName != "" {
		q = q.Where("table_name = ?", filter.TableName)
	}
	if filter.Operation != "" {
		q = q.Where("operation = ?", filter.Operation)
	}
	if filter.ChangedBy != nil {
		q = q.Where("changed_by = ?", *filter.ChangedBy)
	}
	if filter.From != nil {
		q = q.Where("changed_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("changed_at <= ?", *filter.To)
	}

	q = applyOrder(q, filter.Pagination, "changed_at")
	q = applyPagination(q, filter.Pagination)

	var as []models2.Audit
	err := q.Find(&as).Error
	return as, err
}
