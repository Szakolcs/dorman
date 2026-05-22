package administration

import (
	models2 "dorm-man/internal/models"
	"strings"
	"time"

	"dorm-man/internal/models/forum"

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
	RecentAudits  []models2.Audit
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
	planned := models2.JobStatusPlanned
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

func (s *Service) ListTenants(filter TenantFilter) ([]models2.Tenant, error) {
	return s.store.getTenants(filter)
}

func (s *Service) GetTenant(id uuid.UUID) (models2.Tenant, error) {
	if id == uuid.Nil {
		return models2.Tenant{}, ErrValidation
	}
	return s.store.getTenantByID(id)
}

// ---------------------------------------------------------------------------
// inventory
// ---------------------------------------------------------------------------

func (s *Service) ListInventory(filter InventoryFilter) ([]models2.InventoryItem, error) {
	return s.store.getInventory(filter)
}

func (s *Service) GetInventoryItem(id uuid.UUID) (models2.InventoryItem, error) {
	if id == uuid.Nil {
		return models2.InventoryItem{}, ErrValidation
	}
	return s.store.getInventoryByID(id)
}

func (s *Service) CreateInventoryItem(actorID uuid.UUID, req CreateInventoryItemRequest) (models2.InventoryItem, error) {
	if actorID == uuid.Nil {
		return models2.InventoryItem{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Name) == "" {
		return models2.InventoryItem{}, ErrValidation
	}
	if req.PurchaseDate.IsZero() {
		req.PurchaseDate = time.Now()
	}
	if req.Condition == "" {
		req.Condition = models2.InventoryConditionGood
	}
	if req.Status == "" {
		req.Status = models2.InventoryStatusInUse
	}
	item := models2.InventoryItem{
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
		return models2.InventoryItem{}, err
	}
	return item, nil
}

func (s *Service) UpdateInventoryItemStatus(actorID uuid.UUID, req UpdateInventoryStatusRequest) (models2.InventoryItem, error) {
	if actorID == uuid.Nil {
		return models2.InventoryItem{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models2.InventoryItem{}, ErrValidation
	}
	item, err := s.store.getInventoryByID(req.ID)
	if err != nil {
		return models2.InventoryItem{}, err
	}
	if !isInventoryStatusTransitionValid(item.Status, req.Status) {
		return models2.InventoryItem{}, ErrStateTransition
	}
	now := time.Now()
	item.Status = req.Status
	if req.Condition != nil {
		item.Condition = *req.Condition
	}
	switch req.Status {
	case models2.InventoryStatusInUse:
		if item.InUseDate == nil {
			item.InUseDate = &now
		}
	case models2.InventoryStatusWithdrawn, models2.InventoryStatusDestroyed:
		item.WithdrawDate = &now
	}
	if err := s.store.updateInventoryItem(actorID, item); err != nil {
		return models2.InventoryItem{}, err
	}
	return item, nil
}

// isInventoryStatusTransitionValid encodes the lifecycle: in_stock -> in_use,
// in_stock|in_use -> withdrawn, anything -> destroyed. Withdrawn/destroyed are
// terminal.
func isInventoryStatusTransitionValid(from, to models2.InventoryStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case models2.InventoryStatusInStock:
		return to == models2.InventoryStatusInUse ||
			to == models2.InventoryStatusWithdrawn ||
			to == models2.InventoryStatusDestroyed
	case models2.InventoryStatusInUse:
		return to == models2.InventoryStatusInStock ||
			to == models2.InventoryStatusWithdrawn ||
			to == models2.InventoryStatusDestroyed
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

func (s *Service) ListJobs(filter JobsFilter) ([]models2.OperationalJob, error) {
	return s.store.getJobs(filter)
}

func (s *Service) GetJob(id uuid.UUID) (models2.OperationalJob, error) {
	if id == uuid.Nil {
		return models2.OperationalJob{}, ErrValidation
	}
	return s.store.getJobByID(id)
}

func (s *Service) CreateJob(actorID uuid.UUID, req CreateJobRequest) (models2.OperationalJob, error) {
	if actorID == uuid.Nil {
		return models2.OperationalJob{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models2.OperationalJob{}, ErrValidation
	}
	if !req.StartsAt.IsZero() && !req.EndsAt.IsZero() && req.EndsAt.Before(req.StartsAt) {
		return models2.OperationalJob{}, ErrValidation
	}
	if req.Priority == "" {
		req.Priority = models2.JobPriorityMedium
	}
	if req.Status == "" {
		req.Status = models2.JobStatusPlanned
	}
	job := models2.OperationalJob{
		Title:       req.Title,
		Description: req.Description,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Priority:    req.Priority,
		Status:      req.Status,
	}
	if err := s.store.createJob(actorID, job); err != nil {
		return models2.OperationalJob{}, err
	}
	return job, nil
}

func (s *Service) UpdateJob(actorID uuid.UUID, req UpdateJobRequest) (models2.OperationalJob, error) {
	if actorID == uuid.Nil {
		return models2.OperationalJob{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models2.OperationalJob{}, ErrValidation
	}
	job, err := s.store.getJobByID(req.ID)
	if err != nil {
		return models2.OperationalJob{}, err
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
			return models2.OperationalJob{}, ErrStateTransition
		}
		job.Status = *req.Status
	}
	if !job.StartsAt.IsZero() && !job.EndsAt.IsZero() && job.EndsAt.Before(job.StartsAt) {
		return models2.OperationalJob{}, ErrValidation
	}
	if err := s.store.updateJob(actorID, job); err != nil {
		return models2.OperationalJob{}, err
	}
	return job, nil
}

// isJobStatusTransitionValid mirrors a typical scheduling lifecycle:
// planned <-> scheduled -> in_progress -> finished, plus canceled from any
// non-terminal state.
func isJobStatusTransitionValid(from, to models2.JobStatus) bool {
	if from == to {
		return true
	}
	if to == models2.JobStatusCanceled &&
		from != models2.JobStatusFinished &&
		from != models2.JobStatusCanceled {
		return true
	}
	switch from {
	case models2.JobStatusPlanned:
		return to == models2.JobStatusScheduled || to == models2.JobStatusInProgress
	case models2.JobStatusScheduled:
		return to == models2.JobStatusPlanned || to == models2.JobStatusInProgress
	case models2.JobStatusInProgress:
		return to == models2.JobStatusFinished
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

func (s *Service) ListPublications(filter PublicationFilter) ([]models2.Publication, error) {
	return s.store.getPublications(filter)
}

// PublicationDetail is a discriminated union for the publication detail page.
// Exactly one of News/Activity/Event is non-nil based on Kind.
type PublicationDetail struct {
	Kind     PublicationKind
	News     *models2.Publication
	Activity *models2.Activity
	Event    *models2.Event
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

func (s *Service) CreateNews(actorID uuid.UUID, req CreateNewsRequest) (models2.Publication, error) {
	if actorID == uuid.Nil {
		return models2.Publication{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models2.Publication{}, ErrValidation
	}
	if req.State == "" {
		req.State = forum.PublicationStateDraft
	}
	pub := models2.Publication{
		Post: models2.Post{
			Title:       req.Title,
			Description: req.Description,
			State:       req.State,
			AuthorID:    actorID,
		},
	}
	if err := s.store.createPublication(actorID, pub); err != nil {
		return models2.Publication{}, err
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

func (s *Service) CreateActivity(actorID uuid.UUID, req CreateActivityRequest) (models2.Activity, error) {
	if actorID == uuid.Nil {
		return models2.Activity{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models2.Activity{}, ErrValidation
	}
	if req.Capacity < 1 {
		return models2.Activity{}, ErrValidation
	}
	if req.State == "" {
		req.State = forum.PublicationStateDraft
	}
	act := models2.Activity{
		Post: models2.Post{
			Title:       req.Title,
			Description: req.Description,
			State:       req.State,
			AuthorID:    actorID,
		},
		SharedAreaID: req.SharedAreaID,
		Capacity:     req.Capacity,
	}
	if err := s.store.createActivity(actorID, act); err != nil {
		return models2.Activity{}, err
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

func (s *Service) CreateEvent(actorID uuid.UUID, req CreateEventRequest) (models2.Event, error) {
	if actorID == uuid.Nil {
		return models2.Event{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return models2.Event{}, ErrValidation
	}
	if req.StartsAt.IsZero() || req.EndsAt.IsZero() || req.EndsAt.Before(req.StartsAt) {
		return models2.Event{}, ErrValidation
	}
	if req.State == "" {
		req.State = forum.PublicationStateDraft
	}
	evt := models2.Event{
		Post: models2.Post{
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
		return models2.Event{}, err
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

// ---------------------------------------------------------------------------
// housing — buildings / flats / shared areas / rooms
// ---------------------------------------------------------------------------

// HousingOverview bundles every top-level housing entity for the housing index page.
type HousingOverview struct {
	Buildings   []models2.Building
	Flats       []models2.Flat
	SharedAreas []models2.SharedArea
	Rooms       []models2.Room
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

func (s *Service) GetBuilding(id uuid.UUID) (models2.Building, error) {
	if id == uuid.Nil {
		return models2.Building{}, ErrValidation
	}
	return s.store.getBuildingByID(id)
}

func (s *Service) GetFlat(id uuid.UUID) (models2.Flat, error) {
	if id == uuid.Nil {
		return models2.Flat{}, ErrValidation
	}
	return s.store.getFlatByID(id)
}

func (s *Service) GetSharedArea(id uuid.UUID) (models2.SharedArea, error) {
	if id == uuid.Nil {
		return models2.SharedArea{}, ErrValidation
	}
	return s.store.getSharedAreaByID(id)
}

func (s *Service) GetRoom(id uuid.UUID) (models2.Room, error) {
	if id == uuid.Nil {
		return models2.Room{}, ErrValidation
	}
	return s.store.getRoomByID(id)
}

// ---------------------------------------------------------------------------
// room assignments
// ---------------------------------------------------------------------------

func (s *Service) AssignRoom(actorID uuid.UUID, req AssignRoomRequest) (models2.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return models2.RoomAssignment{}, ErrUnauthorized
	}
	if req.TenantID == uuid.Nil || req.RoomID == uuid.Nil {
		return models2.RoomAssignment{}, ErrValidation
	}
	tenant, err := s.store.getTenantByID(req.TenantID)
	if err != nil {
		return models2.RoomAssignment{}, err
	}
	if !tenant.IsActive {
		return models2.RoomAssignment{}, ErrStudentStatusInvalid
	}
	room, err := s.store.getRoomByID(req.RoomID)
	if err != nil {
		return models2.RoomAssignment{}, err
	}
	active := 0
	for _, a := range room.Assignments {
		if a.EndedAt == nil {
			active++
		}
	}
	if active >= room.Capacity {
		return models2.RoomAssignment{}, ErrCapacityConflict
	}
	effectiveAt := time.Now()
	if req.EffectiveAt != nil {
		effectiveAt = *req.EffectiveAt
	}
	assignment := models2.RoomAssignment{
		TenantID:    req.TenantID,
		RoomID:      req.RoomID,
		EffectiveAt: effectiveAt,
	}
	if err := s.store.createRoomAssignment(actorID, assignment); err != nil {
		return models2.RoomAssignment{}, err
	}
	return assignment, nil
}

func (s *Service) MassAssignRooms(actorID uuid.UUID, req MassAssignRequest) ([]models2.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return nil, ErrUnauthorized
	}
	if len(req.Assignments) == 0 {
		return nil, ErrValidation
	}
	out := make([]models2.RoomAssignment, 0, len(req.Assignments))
	for _, a := range req.Assignments {
		assignment, err := s.AssignRoom(actorID, a)
		if err != nil {
			return out, err
		}
		out = append(out, assignment)
	}
	return out, nil
}

func (s *Service) UpdateAssignment(actorID uuid.UUID, req UpdateAssignmentRequest) (models2.RoomAssignment, error) {
	if actorID == uuid.Nil {
		return models2.RoomAssignment{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models2.RoomAssignment{}, ErrValidation
	}
	a, err := s.store.getRoomAssignmentByID(req.ID)
	if err != nil {
		return models2.RoomAssignment{}, err
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
		return models2.RoomAssignment{}, err
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

func (s *Service) RegisterUser(actorID uuid.UUID, req RegisterUserRequest) (models2.User, error) {
	if actorID == uuid.Nil {
		return models2.User{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Nickname) == "" ||
		strings.TrimSpace(req.Password) == "" {
		return models2.User{}, ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models2.User{}, err
	}
	user := models2.User{
		Name:         req.Name,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Nickname:     strings.TrimSpace(req.Nickname),
		PasswordHash: string(hash),
		AvatarURL:    req.AvatarURL,
		PhotoUrl:     req.PhotoURL,
		RoleID:       req.RoleID.String(),
	}
	if err := s.store.registerUser(actorID, user); err != nil {
		return models2.User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUser(actorID uuid.UUID, req UpdateUserRequest) (models2.User, error) {
	if actorID == uuid.Nil {
		return models2.User{}, ErrUnauthorized
	}
	if req.ID == uuid.Nil {
		return models2.User{}, ErrValidation
	}
	user := models2.User{
		BaseModel: models2.BaseModel{ID: req.ID},
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
			return models2.User{}, err
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
		user.RoleID = req.RoleID.String()
	}
	if err := s.store.updateUser(actorID, user); err != nil {
		return models2.User{}, err
	}
	return user, nil
}

// ---------------------------------------------------------------------------
// audit
// ---------------------------------------------------------------------------

func (s *Service) ListAuditLogs(filter AuditLogFilter) ([]models2.Audit, error) {
	return s.store.getAuditLogs(filter)
}
