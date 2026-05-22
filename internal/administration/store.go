package administration

import (
	"dorm-man/internal/models"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditedTransaction runs fn inside a database transaction with the
// `app.current_app_user` PostgreSQL session variable set to the supplied
// userID, so that DB-level audit triggers can attribute the change.
func AuditedTransaction(db *gorm.DB, userID uuid.UUID, fn func(tx *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT set_config('app.current_app_user', ?, true)", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}

type Store interface {
	getTenantsAll() ([]models.Tenant, error)
	getTenants(filter TenantFilter) ([]models.Tenant, error)
	getTenantByID(id uuid.UUID) (models.Tenant, error)

	getInventory(filter InventoryFilter) ([]models.InventoryItem, error)
	getInventoryByID(id uuid.UUID) (models.InventoryItem, error)
	getInventoryByFlatID(flatID uuid.UUID) ([]models.InventoryItem, error)
	getInventoryByBuildingID(buildingID uuid.UUID) ([]models.InventoryItem, error)
	getInventoryBySharedAreaID(sharedAreaID uuid.UUID) ([]models.InventoryItem, error)
	createInventoryItem(actorID uuid.UUID, item models.InventoryItem) error
	updateInventoryItem(actorID uuid.UUID, item models.InventoryItem) error
	deleteInventoryItem(actorID, id uuid.UUID) error

	getJobs(filter JobsFilter) ([]models.OperationalJob, error)
	getJobByID(id uuid.UUID) (models.OperationalJob, error)
	createJob(actorID uuid.UUID, job models.OperationalJob) error
	updateJob(actorID uuid.UUID, job models.OperationalJob) error
	deleteJob(actorID, id uuid.UUID) error

	getPublications(filter PublicationFilter) ([]PublicationListItem, error)
	getNewsByID(id uuid.UUID) (models.Publication, error)
	createPublication(actorID uuid.UUID, pub models.Publication) error
	updatePublication(actorID uuid.UUID, pub models.Publication) error
	deletePublication(actorID, id uuid.UUID) error
	archiveNews(actorID, id uuid.UUID) error
	updateNewsState(actorID, id uuid.UUID, state models.PublicationState) error

	getActivityByID(id uuid.UUID) (models.Activity, error)
	countActivities() (int, error)
	hasActivityBooking(activityID, userID uuid.UUID) (bool, error)
	bookActivity(userID, activityID uuid.UUID) error
	cancelActivityBooking(userID, activityID uuid.UUID) error
	getEventAttendanceIntent(eventID, userID uuid.UUID) (models.EventAttendanceIntent, error)
	setEventAttendanceIntent(userID, eventID uuid.UUID, intent models.EventAttendanceIntent) error
	createActivity(actorID uuid.UUID, a models.Activity) error
	updateActivity(actorID uuid.UUID, a models.Activity) error
	archiveActivity(actorID, id uuid.UUID) error
	updateActivityState(actorID, id uuid.UUID, state models.PublicationState) error

	getEventByID(id uuid.UUID) (models.Event, error)
	createEvent(actorID uuid.UUID, e models.Event) error
	updateEvent(actorID uuid.UUID, e models.Event) error
	archiveEvent(actorID, id uuid.UUID) error
	updateEventState(actorID, id uuid.UUID, state models.PublicationState) error

	getBuilding(filter BuildingFilter) ([]models.Building, error)
	getBuildingByID(id uuid.UUID) (models.Building, error)
	createBuilding(actorID uuid.UUID, building models.Building) error
	updateBuilding(actorID uuid.UUID, building models.Building) error
	deleteBuilding(actorID, id uuid.UUID) error

	getFlat(filter FlatFilter) ([]models.Flat, error)
	getFlatByID(id uuid.UUID) (models.Flat, error)
	createFlat(actorID uuid.UUID, flat models.Flat) error
	updateFlat(actorID uuid.UUID, flat models.Flat) error
	deleteFlat(actorID, id uuid.UUID) error

	getSharedArea(filter SharedAreaFilter) ([]models.SharedArea, error)
	getSharedAreaByID(id uuid.UUID) (models.SharedArea, error)
	createSharedArea(actorID uuid.UUID, sharedArea models.SharedArea) error
	updateSharedArea(actorID uuid.UUID, sharedArea models.SharedArea) error
	deleteSharedArea(actorID, id uuid.UUID) error

	getRoom(filter RoomFilter) ([]models.Room, error)
	getRoomByID(id uuid.UUID) (models.Room, error)
	createRoom(actorID uuid.UUID, room models.Room) error
	updateRoom(actorID uuid.UUID, room models.Room) error
	deleteRoom(actorID, id uuid.UUID) error

	getRoomAssigment(filter RoomAssignmentFilter) ([]models.RoomAssignment, error)
	getRoomAssignmentByID(id uuid.UUID) (models.RoomAssignment, error)
	createRoomAssignment(actorID uuid.UUID, assignment models.RoomAssignment) error
	updateRoomAssignment(actorID uuid.UUID, assignment models.RoomAssignment) error
	deleteRoomAssignment(actorID, id uuid.UUID) error

	registerUser(actorID uuid.UUID, user models.User) error
	updateUser(actorID uuid.UUID, user models.User) error
	createTenant(actorID uuid.UUID, user models.User, tenant models.Tenant) error
	getAuditLogs(filter AuditLogFilter) ([]models.Audit, error)
	getRoles() ([]models.Role, error)
	getRoleByID(id uuid.UUID) (models.Role, error)
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

func (s *GormStore) getTenantsAll() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.db.
		Preload("User").
		Preload("RoomAssignments").
		Order("registered_at DESC").
		Find(&tenants).Error
	return tenants, err
}

func (s *GormStore) getTenants(filter TenantFilter) ([]models.Tenant, error) {
	q := s.db.Model(&models.Tenant{}).
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

	var tenants []models.Tenant
	err := q.Find(&tenants).Error
	return tenants, err
}

// ---------------------------------------------------------------------------
// inventory
// ---------------------------------------------------------------------------

func (s *GormStore) getInventory(filter InventoryFilter) ([]models.InventoryItem, error) {
	q := s.db.Model(&models.InventoryItem{}).
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

	var items []models.InventoryItem
	err := q.Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryByID(id uuid.UUID) (models.InventoryItem, error) {
	var item models.InventoryItem
	err := s.db.
		Preload("Room").
		Preload("Flat").
		Preload("Building").
		Preload("SharedArea").
		First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.InventoryItem{}, ErrNotFound
	}
	return item, err
}

func (s *GormStore) getInventoryByFlatID(flatID uuid.UUID) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Where("flat_id = ?", flatID).Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryByBuildingID(buildingID uuid.UUID) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Where("building_id = ?", buildingID).Find(&items).Error
	return items, err
}

func (s *GormStore) getInventoryBySharedAreaID(sharedAreaID uuid.UUID) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Where("shared_area_id = ?", sharedAreaID).Find(&items).Error
	return items, err
}

func (s *GormStore) createInventoryItem(actorID uuid.UUID, item models.InventoryItem) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&item).Error
	})
}

func (s *GormStore) updateInventoryItem(actorID uuid.UUID, item models.InventoryItem) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&item).Error
	})
}

func (s *GormStore) deleteInventoryItem(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.InventoryItem{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// operational jobs
// ---------------------------------------------------------------------------

func (s *GormStore) getJobs(filter JobsFilter) ([]models.OperationalJob, error) {
	q := s.db.Model(&models.OperationalJob{})

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

	q = applyOrder(q, filter.Pagination, "ends_at")
	q = applyPagination(q, filter.Pagination)

	var jobs []models.OperationalJob
	err := q.Find(&jobs).Error
	return jobs, err
}

func (s *GormStore) createJob(actorID uuid.UUID, job models.OperationalJob) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&job).Error
	})
}

func (s *GormStore) updateJob(actorID uuid.UUID, job models.OperationalJob) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&job).Error
	})
}

func (s *GormStore) deleteJob(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.OperationalJob{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// publications
//
// models.Activity and models.Event each embed models.Publication as a base.
// There is no unified `publications` table in the current migration set, so
// getPublications queries the `activities` / `events` tables directly and
// projects only the shared base columns (id/title/state/timestamps) into
// models.Publication.
//
// createPublication / updatePublication / deletePublication operate on the
// hypothetical `publications` table; if/when a stand-alone Publication model
// gets migrated they will work as-is.
// ---------------------------------------------------------------------------

func (s *GormStore) getPublications(filter PublicationFilter) ([]PublicationListItem, error) {
	var out []PublicationListItem
	var err error

	switch filter.Kind {
	case PublicationKindActivity:
		out, err = s.fetchActivityPublications(filter)
	case PublicationKindEvent:
		out, err = s.fetchEventPublications(filter)
	case PublicationKindNews:
		out, err = s.fetchNewsPublications(filter)
	default:
		news, err := s.fetchNewsPublications(filter)
		if err != nil {
			return nil, err
		}
		acts, err := s.fetchActivityPublications(filter)
		if err != nil {
			return nil, err
		}
		evts, err := s.fetchEventPublications(filter)
		if err != nil {
			return nil, err
		}
		out = append(news, acts...)
		out = append(out, evts...)
		sortPublicationListItems(out)
	}
	if err != nil {
		return nil, err
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
		return []PublicationListItem{}, nil
	}
	end := start + per
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], nil
}

func applyPublicationFilters(q *gorm.DB, filter PublicationFilter) *gorm.DB {
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
	return q
}

func sortPublicationListItems(out []PublicationListItem) {
	sort.Slice(out, func(i, j int) bool {
		left := publicationListSortKey(out[i])
		right := publicationListSortKey(out[j])
		if left.Equal(right) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return left.After(right)
	})
}

func publicationListSortKey(item PublicationListItem) time.Time {
	if !item.UpdatedAt.IsZero() {
		return item.UpdatedAt
	}
	return item.CreatedAt
}

func (s *GormStore) fetchNewsPublications(filter PublicationFilter) ([]PublicationListItem, error) {
	q := applyPublicationFilters(s.db.Model(&models.Publication{}), filter)
	q = applyOrder(q, filter.Pagination, "updated_at")
	var rows []models.Publication
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]PublicationListItem, len(rows))
	for i, row := range rows {
		out[i] = PublicationListItem{
			Publication: row,
			Kind:        PublicationKindNews,
		}
	}
	return out, nil
}

func (s *GormStore) fetchActivityPublications(filter PublicationFilter) ([]PublicationListItem, error) {
	q := applyPublicationFilters(s.db.Model(&models.Activity{}), filter)
	q = applyOrder(q, filter.Pagination, "updated_at")
	var rows []models.Activity
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]PublicationListItem, len(rows))
	for i, row := range rows {
		out[i] = PublicationListItem{
			Publication: publicationFromPost(row.Post),
			Kind:        PublicationKindActivity,
			Activity: &ActivityStats{
				Capacity:    row.Capacity,
				BookedCount: row.BookedCount,
			},
		}
	}
	return out, nil
}

func (s *GormStore) fetchEventPublications(filter PublicationFilter) ([]PublicationListItem, error) {
	q := applyPublicationFilters(s.db.Model(&models.Event{}), filter)
	q = applyOrder(q, filter.Pagination, "starts_at")
	var rows []models.Event
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	statsByEvent, err := s.eventInterestStatsFor(rows)
	if err != nil {
		return nil, err
	}
	out := make([]PublicationListItem, len(rows))
	for i, row := range rows {
		st := statsByEvent[row.ID]
		out[i] = PublicationListItem{
			Publication: publicationFromPost(row.Post),
			Kind:        PublicationKindEvent,
			Event:       &st,
		}
	}
	return out, nil
}

func publicationFromPost(post models.Post) models.Publication {
	return models.Publication{Post: post}
}

type eventIntentCount struct {
	EventID uuid.UUID `gorm:"column:event_id"`
	Intent  string    `gorm:"column:intent"`
	Count   int       `gorm:"column:count"`
}

func (s *GormStore) eventInterestStatsFor(events []models.Event) (map[uuid.UUID]EventInterestStats, error) {
	out := make(map[uuid.UUID]EventInterestStats, len(events))
	for _, evt := range events {
		out[evt.ID] = EventInterestStats{}
	}
	if len(events) == 0 {
		return out, nil
	}
	ids := make([]uuid.UUID, len(events))
	for i, evt := range events {
		ids[i] = evt.ID
	}
	var counts []eventIntentCount
	err := s.db.Model(&models.EventAttendance{}).
		Select("event_id, intent, COUNT(*) AS count").
		Where("event_id IN ?", ids).
		Group("event_id, intent").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	for _, row := range counts {
		st := out[row.EventID]
		switch models.EventAttendanceIntent(row.Intent) {
		case models.EventAttendanceInterested:
			st.Interested = row.Count
		case models.EventAttendanceNotInterested:
			st.NotInterested = row.Count
		case models.EventAttendanceBusy:
			st.Busy = row.Count
		}
		out[row.EventID] = st
	}
	return out, nil
}

func (s *GormStore) createPublication(actorID uuid.UUID, pub models.Publication) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&pub).Error
	})
}

func (s *GormStore) updatePublication(actorID uuid.UUID, pub models.Publication) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&pub).Error
	})
}

func (s *GormStore) deletePublication(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.Publication{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// buildings
// ---------------------------------------------------------------------------

func (s *GormStore) getBuilding(filter BuildingFilter) ([]models.Building, error) {
	q := s.db.Model(&models.Building{}).
		Preload("Flats").
		Preload("SharedAreas")

	if filter.Search != "" {
		like := likePattern(filter.Search)
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}

	q = applyOrder(q, filter.Pagination, "name")
	q = applyPagination(q, filter.Pagination)

	var bs []models.Building
	err := q.Find(&bs).Error
	return bs, err
}

func (s *GormStore) createBuilding(actorID uuid.UUID, b models.Building) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&b).Error
	})
}

func (s *GormStore) updateBuilding(actorID uuid.UUID, b models.Building) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&b).Error
	})
}

func (s *GormStore) deleteBuilding(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.Building{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// flats
// ---------------------------------------------------------------------------

func (s *GormStore) getFlat(filter FlatFilter) ([]models.Flat, error) {
	q := s.db.Model(&models.Flat{}).
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

	var flats []models.Flat
	err := q.Find(&flats).Error
	return flats, err
}

func (s *GormStore) createFlat(actorID uuid.UUID, f models.Flat) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&f).Error
	})
}

func (s *GormStore) updateFlat(actorID uuid.UUID, f models.Flat) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&f).Error
	})
}

func (s *GormStore) deleteFlat(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.Flat{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// shared areas
// ---------------------------------------------------------------------------

func (s *GormStore) getSharedArea(filter SharedAreaFilter) ([]models.SharedArea, error) {
	q := s.db.Model(&models.SharedArea{}).
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

	var sas []models.SharedArea
	err := q.Find(&sas).Error
	return sas, err
}

func (s *GormStore) createSharedArea(actorID uuid.UUID, sa models.SharedArea) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&sa).Error
	})
}

func (s *GormStore) updateSharedArea(actorID uuid.UUID, sa models.SharedArea) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&sa).Error
	})
}

func (s *GormStore) deleteSharedArea(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.SharedArea{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// rooms
// ---------------------------------------------------------------------------

func (s *GormStore) getRoom(filter RoomFilter) ([]models.Room, error) {
	q := s.db.Model(&models.Room{}).
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

	var rooms []models.Room
	err := q.Find(&rooms).Error
	return rooms, err
}

func (s *GormStore) createRoom(actorID uuid.UUID, r models.Room) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&r).Error
	})
}

func (s *GormStore) updateRoom(actorID uuid.UUID, r models.Room) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&r).Error
	})
}

func (s *GormStore) deleteRoom(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.Room{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// room assignments
// ---------------------------------------------------------------------------

func (s *GormStore) getRoomAssigment(filter RoomAssignmentFilter) ([]models.RoomAssignment, error) {
	q := s.db.Model(&models.RoomAssignment{}).
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

	var as []models.RoomAssignment
	err := q.Find(&as).Error
	return as, err
}

func (s *GormStore) createRoomAssignment(actorID uuid.UUID, a models.RoomAssignment) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&a).Error
	})
}

func (s *GormStore) updateRoomAssignment(actorID uuid.UUID, a models.RoomAssignment) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&a).Error
	})
}

func (s *GormStore) deleteRoomAssignment(actorID, id uuid.UUID) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Delete(&models.RoomAssignment{}, "id = ?", id).Error
	})
}

// ---------------------------------------------------------------------------
// users
// ---------------------------------------------------------------------------

func (s *GormStore) registerUser(actorID uuid.UUID, user models.User) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&user).Error
	})
}
func (s *GormStore) createTenant(actorID uuid.UUID, user models.User, tenant models.Tenant) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		tenant.UserID = &user.ID

		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		return nil
	})
}
func (s *GormStore) updateUser(actorID uuid.UUID, user models.User) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&user).Error
	})
}

// ---------------------------------------------------------------------------
// by-id helpers
// ---------------------------------------------------------------------------

func (s *GormStore) getTenantByID(id uuid.UUID) (models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.
		Preload("User").
		Preload("RoomAssignments").
		Preload("RoomAssignments.Room").
		Preload("RoomAssignments.Room.Flat").
		First(&tenant, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (s *GormStore) getJobByID(id uuid.UUID) (models.OperationalJob, error) {
	var job models.OperationalJob
	err := s.db.First(&job, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OperationalJob{}, ErrNotFound
	}
	return job, err
}

func (s *GormStore) getBuildingByID(id uuid.UUID) (models.Building, error) {
	var b models.Building
	err := s.db.
		Preload("Flats").
		Preload("SharedAreas").
		First(&b, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Building{}, ErrNotFound
	}
	return b, err
}

func (s *GormStore) getFlatByID(id uuid.UUID) (models.Flat, error) {
	var f models.Flat
	err := s.db.
		Preload("Building").
		Preload("Rooms").
		Preload("Rooms.Assignments", "ended_at IS NULL").
		Preload("Inventory").
		First(&f, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Flat{}, ErrNotFound
	}
	return f, err
}

func (s *GormStore) getSharedAreaByID(id uuid.UUID) (models.SharedArea, error) {
	var sa models.SharedArea
	err := s.db.
		Preload("Building").
		Preload("Inventory").
		First(&sa, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.SharedArea{}, ErrNotFound
	}
	return sa, err
}

func (s *GormStore) getRoomByID(id uuid.UUID) (models.Room, error) {
	var r models.Room
	err := s.db.
		Preload("Flat").
		Preload("Flat.Building").
		Preload("Assignments", "ended_at IS NULL").
		Preload("Assignments.Tenant").
		Preload("Assignments.Tenant.User").
		First(&r, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Room{}, ErrNotFound
	}
	return r, err
}

func (s *GormStore) getRoomAssignmentByID(id uuid.UUID) (models.RoomAssignment, error) {
	var a models.RoomAssignment
	err := s.db.
		Preload("Tenant").
		Preload("Tenant.User").
		Preload("Room").
		Preload("Room.Flat").
		First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.RoomAssignment{}, ErrNotFound
	}
	return a, err
}

// ---------------------------------------------------------------------------
// news (models.Publication) detail + archive
// ---------------------------------------------------------------------------

func (s *GormStore) getNewsByID(id uuid.UUID) (models.Publication, error) {
	var p models.Publication
	err := s.db.Preload("Author").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Publication{}, ErrNotFound
	}
	return p, err
}

func (s *GormStore) archiveNews(actorID, id uuid.UUID) error {
	return s.updateNewsState(actorID, id, models.PublicationStateArchived)
}

func (s *GormStore) updateNewsState(actorID, id uuid.UUID, state models.PublicationState) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models.Publication{}).
			Where("id = ?", id).
			Update("state", state)
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

func (s *GormStore) getActivityByID(id uuid.UUID) (models.Activity, error) {
	var a models.Activity
	err := s.db.
		Preload("Author").
		Preload("Bookings").
		First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Activity{}, ErrNotFound
	}
	return a, err
}

func (s *GormStore) countActivities() (int, error) {
	var n int64
	err := s.db.Model(&models.Activity{}).
		Where("state != ?", models.PublicationStateArchived).
		Count(&n).Error
	return int(n), err
}

func (s *GormStore) hasActivityBooking(activityID, userID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&models.ActivityBooking{}).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		Count(&count).Error
	return count > 0, err
}

func (s *GormStore) bookActivity(userID, activityID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var activity models.Activity
		if err := tx.First(&activity, "id = ?", activityID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if activity.State != models.PublicationStatePublished {
			return ErrValidation
		}
		if activity.BookedCount >= activity.Capacity {
			return ErrCapacityConflict
		}
		var existing int64
		if err := tx.Model(&models.ActivityBooking{}).
			Where("activity_id = ? AND user_id = ?", activityID, userID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrConcurrencyConflict
		}
		booking := models.ActivityBooking{
			ActivityID: activityID,
			UserID:     userID,
		}
		if err := tx.Create(&booking).Error; err != nil {
			return err
		}
		res := tx.Model(&models.Activity{}).
			Where("id = ? AND booked_count < capacity", activityID).
			UpdateColumn("booked_count", gorm.Expr("booked_count + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrCapacityConflict
		}
		return nil
	})
}

func (s *GormStore) cancelActivityBooking(userID, activityID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var booking models.ActivityBooking
		err := tx.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&booking).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&booking).Error; err != nil {
			return err
		}
		res := tx.Model(&models.Activity{}).
			Where("id = ? AND booked_count > 0", activityID).
			UpdateColumn("booked_count", gorm.Expr("booked_count - 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *GormStore) getEventAttendanceIntent(eventID, userID uuid.UUID) (models.EventAttendanceIntent, error) {
	var attendance models.EventAttendance
	err := s.db.
		Where("event_id = ? AND user_id = ?", eventID, userID).
		First(&attendance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return attendance.Intent, nil
}

func (s *GormStore) setEventAttendanceIntent(userID, eventID uuid.UUID, intent models.EventAttendanceIntent) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.First(&event, "id = ?", eventID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if event.State != models.PublicationStatePublished {
			return ErrValidation
		}

		var attendance models.EventAttendance
		err := tx.Where("event_id = ? AND user_id = ?", eventID, userID).First(&attendance).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&models.EventAttendance{
				EventID: eventID,
				UserID:  userID,
				Intent:  intent,
			}).Error
		}
		if err != nil {
			return err
		}
		attendance.Intent = intent
		return tx.Save(&attendance).Error
	})
}

func (s *GormStore) createActivity(actorID uuid.UUID, a models.Activity) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&a).Error
	})
}

func (s *GormStore) updateActivity(actorID uuid.UUID, a models.Activity) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&a).Error
	})
}

func (s *GormStore) archiveActivity(actorID, id uuid.UUID) error {
	return s.updateActivityState(actorID, id, models.PublicationStateArchived)
}

func (s *GormStore) updateActivityState(actorID, id uuid.UUID, state models.PublicationState) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models.Activity{}).
			Where("id = ?", id).
			Update("state", state)
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

func (s *GormStore) getEventByID(id uuid.UUID) (models.Event, error) {
	var e models.Event
	err := s.db.
		Preload("Author").
		Preload("Attendances").
		First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Event{}, ErrNotFound
	}
	return e, err
}

func (s *GormStore) createEvent(actorID uuid.UUID, e models.Event) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Create(&e).Error
	})
}

func (s *GormStore) updateEvent(actorID uuid.UUID, e models.Event) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		return tx.Save(&e).Error
	})
}

func (s *GormStore) archiveEvent(actorID, id uuid.UUID) error {
	return s.updateEventState(actorID, id, models.PublicationStateArchived)
}

func (s *GormStore) updateEventState(actorID, id uuid.UUID, state models.PublicationState) error {
	return AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		res := tx.Model(&models.Event{}).
			Where("id = ?", id).
			Update("state", state)
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

func (s *GormStore) getAuditLogs(filter AuditLogFilter) ([]models.Audit, error) {
	q := s.db.Model(&models.Audit{}).Preload("User")

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

	var as []models.Audit
	err := q.Find(&as).Error
	return as, err
}

func (s *GormStore) getRoles() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Order("name ASC").Find(&roles).Error
	return roles, err
}

func (s *GormStore) getRoleByID(id uuid.UUID) (models.Role, error) {
	var role models.Role
	err := s.db.First(&role, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Role{}, ErrNotFound
	}
	return role, err
}
