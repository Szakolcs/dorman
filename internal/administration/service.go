package administration

import (
	"dorm-man/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// ---------------------------------------------------------------------------
// dashboard
// ---------------------------------------------------------------------------

// DashboardSummary aggregates the headline numbers shown on the admin home
// page so the handler can render them with a single call.
type DashboardSummary struct {
	ActiveTenants int
	OpenJobs      int
	Buildings     int
	RecentAudits  []models.Audit
}

func (s *Service) GetDashboardSummary() (DashboardSummary, error) {
	tenants, err := s.store.getTenantsAll()
	if err != nil {
		return DashboardSummary{}, err
	}
	activeTenants := 0
	for _, t := range tenants {
		if t.IsActive {
			activeTenants++
		}
	}

	plannedFilter := JobsFilter{}
	planned := models.JobStatusPlanned
	plannedFilter.Status = &planned
	jobs, err := s.store.getJobs(plannedFilter)
	if err != nil {
		return DashboardSummary{}, err
	}

	buildings, err := s.store.getBuilding(BuildingFilter{})
	if err != nil {
		return DashboardSummary{}, err
	}

	audits, err := s.store.getAuditLogs(AuditLogFilter{
		Pagination: Pagination{Page: 1, PerPage: 5},
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		ActiveTenants: activeTenants,
		OpenJobs:      len(jobs),
		Buildings:     len(buildings),
		RecentAudits:  audits,
	}, nil
}

// ---------------------------------------------------------------------------
// tenants
// ---------------------------------------------------------------------------

func (s *Service) ListTenants(filter TenantFilter) ([]models.Tenant, error) {
	return s.store.getTenants(filter)
}

func (s *Service) GetTenant(id uuid.UUID) (models.Tenant, error) {
	if id == uuid.Nil {
		return models.Tenant{}, ErrValidation
	}
	return s.store.getTenantByID(id)
}

// ---------------------------------------------------------------------------
// inventory
// ---------------------------------------------------------------------------

func (s *Service) ListInventory(filter InventoryFilter) ([]models.InventoryItem, error) {
	return s.store.getInventory(filter)
}

func (s *Service) GetInventoryItem(id uuid.UUID) (models.InventoryItem, error) {
	if id == uuid.Nil {
		return models.InventoryItem{}, ErrValidation
	}
	return s.store.getInventoryByID(id)
}

func (s *Service) CreateInventoryItem(actorID uuid.UUID, req CreateInventoryItemRequest) (models.InventoryItem, error) {
	if actorID == uuid.Nil {
		return models.InventoryItem{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Name) == "" {
		return models.InventoryItem{}, ErrValidation
	}
	if req.PurchaseDate.IsZero() {
		req.PurchaseDate = time.Now()
	}
	if req.Condition == "" {
		req.Condition = models.InventoryConditionGood
	}
	if req.Status == "" {
		req.Status = models.InventoryStatusInUse
	}
	item := models.InventoryItem{
		Name:         req.Name,
		Description:  req.Description,
		RoomID:       req.RoomID,
		FlatID:       req.FlatID,
		BuildingID:   req.BuildingID,
		SharedAreaID: req.SharedAreaID,
		Condition:    req.Condition,
		Status:       req.Status,
		PurchaseDate: req.PurchaseDate,
	}
	if err := s.store.createInventoryItem(actorID, item); err != nil {
		return models.InventoryItem{}, err
	}
	return item, nil
}

func (s *Service) UpdateInventoryItemStatus(actorID uuid.UUID, req UpdateInventoryStatusRequest) (models.InventoryItem, error) {
	if actorID == uuid.Nil {
		return models.InventoryItem{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models.InventoryItem{}, ErrValidation
	}
	item, err := s.store.getInventoryByID(req.ID)
	if err != nil {
		return models.InventoryItem{}, err
	}
	if !isInventoryStatusTransitionValid(item.Status, req.Status) {
		return models.InventoryItem{}, ErrStateTransition
	}
	now := time.Now()
	item.Status = req.Status
	if req.Condition != nil {
		item.Condition = *req.Condition
	}
	switch req.Status {
	case models.InventoryStatusInUse:
		if item.InUseDate == nil {
			item.InUseDate = &now
		}
	case models.InventoryStatusWithdrawn, models.InventoryStatusDestroyed:
		item.WithdrawDate = &now
	}
	if err := s.store.updateInventoryItem(actorID, item); err != nil {
		return models.InventoryItem{}, err
	}
	return item, nil
}

// isInventoryStatusTransitionValid encodes the lifecycle: in_stock -> in_use,
// in_stock|in_use -> withdrawn, anything -> destroyed. Withdrawn/destroyed are
// terminal.
func isInventoryStatusTransitionValid(from, to models.InventoryStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case models.InventoryStatusInStock:
		return to == models.InventoryStatusInUse ||
			to == models.InventoryStatusWithdrawn ||
			to == models.InventoryStatusDestroyed
	case models.InventoryStatusInUse:
		return to == models.InventoryStatusInStock ||
			to == models.InventoryStatusWithdrawn ||
			to == models.InventoryStatusDestroyed
	default:
		return false
	}
}

func (s *Service) DeleteInventoryItem(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.deleteInventoryItem(actorID, id)
}

// ---------------------------------------------------------------------------
// operational jobs
// ---------------------------------------------------------------------------

func (s *Service) ListJobs(filter JobsFilter) ([]models.OperationalJob, error) {
	return s.store.getJobs(filter)
}

func (s *Service) GetJob(id uuid.UUID) (models.OperationalJob, error) {
	if id == uuid.Nil {
		return models.OperationalJob{}, ErrValidation
	}
	return s.store.getJobByID(id)
}

func (s *Service) CreateJob(actorID uuid.UUID, req CreateJobRequest) (models.OperationalJob, error) {
	if actorID == uuid.Nil {
		return models.OperationalJob{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models.OperationalJob{}, ErrValidation
	}
	if !req.StartsAt.IsZero() && !req.EndsAt.IsZero() && req.EndsAt.Before(req.StartsAt) {
		return models.OperationalJob{}, ErrValidation
	}
	if req.Priority == "" {
		req.Priority = models.JobPriorityMedium
	}
	if req.Status == "" {
		req.Status = models.JobStatusPlanned
	}
	job := models.OperationalJob{
		Title:       req.Title,
		Description: req.Description,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Priority:    req.Priority,
		Status:      req.Status,
	}
	if err := s.store.createJob(actorID, job); err != nil {
		return models.OperationalJob{}, err
	}
	return job, nil
}

func (s *Service) UpdateJob(actorID uuid.UUID, req UpdateJobRequest) (models.OperationalJob, error) {
	if actorID == uuid.Nil {
		return models.OperationalJob{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models.OperationalJob{}, ErrValidation
	}
	job, err := s.store.getJobByID(req.ID)
	if err != nil {
		return models.OperationalJob{}, err
	}
	if req.Title != nil {
		job.Title = *req.Title
	}
	if req.Description != nil {
		job.Description = *req.Description
	}
	if req.StartsAt != nil {
		job.StartsAt = *req.StartsAt
	}
	if req.EndsAt != nil {
		job.EndsAt = *req.EndsAt
	}
	if req.Priority != nil {
		job.Priority = *req.Priority
	}
	if req.Status != nil {
		if !isJobStatusTransitionValid(job.Status, *req.Status) {
			return models.OperationalJob{}, ErrStateTransition
		}
		job.Status = *req.Status
	}
	if !job.StartsAt.IsZero() && !job.EndsAt.IsZero() && job.EndsAt.Before(job.StartsAt) {
		return models.OperationalJob{}, ErrValidation
	}
	if err := s.store.updateJob(actorID, job); err != nil {
		return models.OperationalJob{}, err
	}
	return job, nil
}

// isJobStatusTransitionValid mirrors a typical scheduling lifecycle:
// planned <-> scheduled -> in_progress -> finished, plus canceled from any
// non-terminal state.
func isJobStatusTransitionValid(from, to models.JobStatus) bool {
	if from == to {
		return true
	}
	if to == models.JobStatusCanceled &&
		from != models.JobStatusFinished &&
		from != models.JobStatusCanceled {
		return true
	}
	switch from {
	case models.JobStatusPlanned:
		return to == models.JobStatusScheduled || to == models.JobStatusInProgress
	case models.JobStatusScheduled:
		return to == models.JobStatusPlanned || to == models.JobStatusInProgress
	case models.JobStatusInProgress:
		return to == models.JobStatusFinished
	default:
		return false
	}
}

func (s *Service) DeleteJob(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.deleteJob(actorID, id)
}

// ---------------------------------------------------------------------------
// publications (news / activities / events)
// ---------------------------------------------------------------------------

func (s *Service) ListPublications(filter PublicationFilter) ([]PublicationListItem, error) {
	return s.store.getPublications(filter)
}

// PublicationDetail is a discriminated union for the publication detail page.
// Exactly one of News/Activity/Event is non-nil based on Kind.
type PublicationDetail struct {
	Kind     PublicationKind
	News     *models.Publication
	Activity *models.Activity
	Event    *models.Event
}

// GetPublication tries each typed table until one matches the id.
func (s *Service) GetPublication(id uuid.UUID) (PublicationDetail, error) {
	if id == uuid.Nil {
		return PublicationDetail{}, ErrValidation
	}
	if news, err := s.store.getNewsByID(id); err == nil {
		return PublicationDetail{Kind: PublicationKindNews, News: &news}, nil
	}
	if act, err := s.store.getActivityByID(id); err == nil {
		return PublicationDetail{Kind: PublicationKindActivity, Activity: &act}, nil
	}
	if evt, err := s.store.getEventByID(id); err == nil {
		return PublicationDetail{Kind: PublicationKindEvent, Event: &evt}, nil
	}
	return PublicationDetail{}, ErrNotFound
}

func (s *Service) CreateNews(actorID uuid.UUID, req CreateNewsRequest) (models.Publication, error) {
	if actorID == uuid.Nil {
		return models.Publication{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models.Publication{}, ErrValidation
	}
	if req.State == "" {
		req.State = models.PublicationStateDraft
	}
	pub := models.Publication{
		Post: models.Post{
			Title:       req.Title,
			Description: req.Description,
			State:       req.State,
			AuthorID:    actorID,
		},
	}
	if err := s.store.createPublication(actorID, pub); err != nil {
		return models.Publication{}, err
	}
	return pub, nil
}

func (s *Service) ArchiveNews(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.archiveNews(actorID, id)
}

func (s *Service) UpdateNewsState(actorID, id uuid.UUID, state models.PublicationState) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil || state == "" {
		return ErrValidation
	}
	return s.store.updateNewsState(actorID, id, state)
}

func (s *Service) CreateActivity(actorID uuid.UUID, req CreateActivityRequest) (models.Activity, error) {
	if actorID == uuid.Nil {
		return models.Activity{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models.Activity{}, ErrValidation
	}
	if req.Capacity < 1 {
		return models.Activity{}, ErrValidation
	}
	if req.State == "" {
		req.State = models.PublicationStateDraft
	}
	act := models.Activity{
		Post: models.Post{
			Title:       req.Title,
			Description: req.Description,
			State:       req.State,
			AuthorID:    actorID,
		},
		SharedAreaID: req.SharedAreaID,
		Capacity:     req.Capacity,
	}
	if err := s.store.createActivity(actorID, act); err != nil {
		return models.Activity{}, err
	}
	return act, nil
}

func (s *Service) ArchiveActivity(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.archiveActivity(actorID, id)
}

func (s *Service) UpdateActivityState(actorID, id uuid.UUID, state models.PublicationState) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil || state == "" {
		return ErrValidation
	}
	return s.store.updateActivityState(actorID, id, state)
}

func (s *Service) CreateEvent(actorID uuid.UUID, req CreateEventRequest) (models.Event, error) {
	if actorID == uuid.Nil {
		return models.Event{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models.Event{}, ErrValidation
	}
	if req.StartsAt.IsZero() || req.EndsAt.IsZero() || req.EndsAt.Before(req.StartsAt) {
		return models.Event{}, ErrValidation
	}
	if req.State == "" {
		req.State = models.PublicationStateDraft
	}
	evt := models.Event{
		Post: models.Post{
			Title:       req.Title,
			Description: req.Description,
			State:       req.State,
			AuthorID:    actorID,
		},
		SharedAreaID: req.SharedAreaID,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
	}
	if err := s.store.createEvent(actorID, evt); err != nil {
		return models.Event{}, err
	}
	return evt, nil
}

func (s *Service) ArchiveEvent(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.archiveEvent(actorID, id)
}

func (s *Service) UpdateEventState(actorID, id uuid.UUID, state models.PublicationState) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil || state == "" {
		return ErrValidation
	}
	return s.store.updateEventState(actorID, id, state)
}

// ---------------------------------------------------------------------------
// housing — buildings / flats / shared areas / rooms
// ---------------------------------------------------------------------------

// HousingOverview bundles every top-level housing entity for the housing index page.
type HousingOverview struct {
	Buildings   []models.Building
	Flats       []models.Flat
	SharedAreas []models.SharedArea
	Rooms       []models.Room
}

func (s *Service) GetHousingOverview() (HousingOverview, error) {
	bs, err := s.store.getBuilding(BuildingFilter{})
	if err != nil {
		return HousingOverview{}, err
	}
	fs, err := s.store.getFlat(FlatFilter{})
	if err != nil {
		return HousingOverview{}, err
	}
	sas, err := s.store.getSharedArea(SharedAreaFilter{})
	if err != nil {
		return HousingOverview{}, err
	}
	rs, err := s.store.getRoom(RoomFilter{})
	if err != nil {
		return HousingOverview{}, err
	}
	return HousingOverview{
		Buildings:   bs,
		Flats:       fs,
		SharedAreas: sas,
		Rooms:       rs,
	}, nil
}

func (s *Service) GetBuilding(id uuid.UUID) (models.Building, error) {
	if id == uuid.Nil {
		return models.Building{}, ErrValidation
	}
	return s.store.getBuildingByID(id)
}

func (s *Service) GetFlat(id uuid.UUID) (models.Flat, error) {
	if id == uuid.Nil {
		return models.Flat{}, ErrValidation
	}
	return s.store.getFlatByID(id)
}

func (s *Service) GetSharedArea(id uuid.UUID) (models.SharedArea, error) {
	if id == uuid.Nil {
		return models.SharedArea{}, ErrValidation
	}
	return s.store.getSharedAreaByID(id)
}

func (s *Service) GetRoom(id uuid.UUID) (models.Room, error) {
	if id == uuid.Nil {
		return models.Room{}, ErrValidation
	}
	return s.store.getRoomByID(id)
}

// ---------------------------------------------------------------------------
// room assignments
// ---------------------------------------------------------------------------

func (s *Service) AssignRoom(actorID uuid.UUID, req AssignRoomRequest) (models.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return models.RoomAssignment{}, ErrUnauthorized
	}
	if req.TenantID == uuid.Nil || req.RoomID == uuid.Nil {
		return models.RoomAssignment{}, ErrValidation
	}
	tenant, err := s.store.getTenantByID(req.TenantID)
	if err != nil {
		return models.RoomAssignment{}, err
	}
	if !tenant.IsActive {
		return models.RoomAssignment{}, ErrStudentStatusInvalid
	}
	room, err := s.store.getRoomByID(req.RoomID)
	if err != nil {
		return models.RoomAssignment{}, err
	}
	active := 0
	for _, a := range room.Assignments {
		if a.EndedAt == nil {
			active++
		}
	}
	if active >= room.Capacity {
		return models.RoomAssignment{}, ErrCapacityConflict
	}
	effectiveAt := time.Now()
	if req.EffectiveAt != nil {
		effectiveAt = *req.EffectiveAt
	}
	assignment := models.RoomAssignment{
		TenantID:    req.TenantID,
		RoomID:      req.RoomID,
		EffectiveAt: effectiveAt,
	}
	if err := s.store.createRoomAssignment(actorID, assignment); err != nil {
		return models.RoomAssignment{}, err
	}
	return assignment, nil
}

func (s *Service) MassAssignRooms(actorID uuid.UUID, req MassAssignRequest) ([]models.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return nil, ErrUnauthorized
	}
	if len(req.Assignments) == 0 {
		return nil, ErrValidation
	}
	out := make([]models.RoomAssignment, 0, len(req.Assignments))
	for _, a := range req.Assignments {
		assignment, err := s.AssignRoom(actorID, a)
		if err != nil {
			return out, err
		}
		out = append(out, assignment)
	}
	return out, nil
}

func (s *Service) UpdateAssignment(actorID uuid.UUID, req UpdateAssignmentRequest) (models.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return models.RoomAssignment{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models.RoomAssignment{}, ErrValidation
	}
	a, err := s.store.getRoomAssignmentByID(req.ID)
	if err != nil {
		return models.RoomAssignment{}, err
	}
	if req.RoomID != nil {
		a.RoomID = *req.RoomID
	}
	if req.EffectiveAt != nil {
		a.EffectiveAt = *req.EffectiveAt
	}
	if req.EndedAt != nil {
		a.EndedAt = req.EndedAt
	}
	if err := s.store.updateRoomAssignment(actorID, a); err != nil {
		return models.RoomAssignment{}, err
	}
	return a, nil
}

func (s *Service) DeleteAssignment(actorID, id uuid.UUID) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}
	if id == uuid.Nil {
		return ErrValidation
	}
	return s.store.deleteRoomAssignment(actorID, id)
}

// ---------------------------------------------------------------------------
// users (registration / update)
// ---------------------------------------------------------------------------

func (s *Service) RegisterUser(actorID uuid.UUID, req RegisterUserRequest) (models.User, error) {
	if actorID == uuid.Nil {
		return models.User{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Nickname) == "" ||
		strings.TrimSpace(req.Password) == "" {
		return models.User{}, ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	role, err := s.store.getRoleByID(req.RoleID)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		Name:         req.Name,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Nickname:     strings.TrimSpace(req.Nickname),
		PasswordHash: string(hash),
		AvatarURL:    req.AvatarURL,
		PhotoUrl:     req.PhotoURL,
		RoleID:       role.ID,
	}
	if role.Name == "tenant" {
		tenant := models.Tenant{
			UserID:      &user.ID,
			StudentCode: req.StudentCode,
			Degree:      req.Degree,
			Faculty:     req.Faculty,
			Age:         req.Age,
			Sex:         req.Sex,
			Nationality: req.Nationality,
		}
		if err := s.store.createTenant(actorID, user, tenant); err != nil {
			return models.User{}, err
		}
	} else {
		if err := s.store.registerUser(actorID, user); err != nil {
			return models.User{}, err
		}
	}
	return user, nil
}

func (s *Service) UpdateUser(actorID uuid.UUID, req UpdateUserRequest) (models.User, error) {
	if actorID == uuid.Nil {
		return models.User{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models.User{}, ErrValidation
	}
	user := models.User{
		BaseModel: models.BaseModel{ID: req.ID},
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Nickname != nil {
		user.Nickname = strings.TrimSpace(*req.Nickname)
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, err
		}
		user.PasswordHash = string(hash)
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.PhotoURL != nil {
		user.PhotoUrl = *req.PhotoURL
	}
	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}
	if err := s.store.updateUser(actorID, user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

// ---------------------------------------------------------------------------
// form lookups (select dropdowns)
// ---------------------------------------------------------------------------

func (s *Service) ListRoles() ([]models.Role, error) {
	return s.store.getRoles()
}

func (s *Service) ListBuildings() ([]models.Building, error) {
	return s.store.getBuilding(BuildingFilter{})
}

func (s *Service) ListFlats() ([]models.Flat, error) {
	return s.store.getFlat(FlatFilter{})
}

func (s *Service) ListRooms() ([]models.Room, error) {
	return s.store.getRoom(RoomFilter{})
}

func (s *Service) ListSharedAreas() ([]models.SharedArea, error) {
	return s.store.getSharedArea(SharedAreaFilter{})
}

func (s *Service) ListActiveTenants() ([]models.Tenant, error) {
	active := true
	return s.store.getTenants(TenantFilter{IsActive: &active})
}

// ---------------------------------------------------------------------------
// audit
// ---------------------------------------------------------------------------

func (s *Service) ListAuditLogs(filter AuditLogFilter) ([]models.Audit, error) {
	return s.store.getAuditLogs(filter)
}
