package administration

import (
	"errors"
	"fmt"
	"time"

	models "dorm-man/internal/models/administration"
	forummodels "dorm-man/internal/models/forum"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ResolvePrincipal(actorID uuid.UUID) (Principal, error) {
	return s.store.LoadPrincipal(actorID)
}

func (s *Service) ListTenants(filter TenantListFilter) ([]models.Tenant, error) {
	return s.store.ListTenants(filter)
}

func (s *Service) RegisterTenant(principal Principal, tenant models.Tenant) (models.Tenant, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Tenant{}, ErrUnauthorized
	}
	if tenant.StudentCode == "" || tenant.Name == "" {
		return models.Tenant{}, ErrValidation
	}
	if tenant.IsActive && tenant.Nationality == nil {
		return models.Tenant{}, ErrStudentStatusInvalid
	}

	created, err := s.store.CreateTenant(tenant)
	if err != nil {
		return models.Tenant{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "tenant.register", "tenant", created.ID, models.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) SetTenantActive(principal Principal, tenantID uuid.UUID, active bool) (models.Tenant, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Tenant{}, ErrUnauthorized
	}
	tenant, err := s.store.UpdateTenantStatus(tenantID, active)
	if err != nil {
		return models.Tenant{}, err
	}
	action := "tenant.deactivate"
	if active {
		action = "tenant.activate"
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, action, "tenant", tenant.ID, models.AuditOutcomeSuccess))
	return tenant, nil
}

func (s *Service) GetTenant(tenantID uuid.UUID) (models.Tenant, error) {
	return s.store.GetTenant(tenantID)
}

func (s *Service) ListRooms(filter RoomListFilter) ([]models.Room, error) {
	rooms, err := s.store.ListRooms(filter)
	if err != nil {
		return nil, err
	}
	if filter.State == "" {
		return rooms, nil
	}

	filtered := make([]models.Room, 0, len(rooms))
	for _, room := range rooms {
		occupancy, err := s.store.GetActiveAssignmentCount(room.ID)
		if err != nil {
			return nil, err
		}
		switch filter.State {
		case "occupied":
			if occupancy > 0 {
				filtered = append(filtered, room)
			}
		case "available":
			if occupancy < int64(room.Capacity) {
				filtered = append(filtered, room)
			}
		case "maintenance-needed":
			// Placeholder for future rule integration, currently inferred by low-quality inventory.
			hasDamagedInventory := false
			for _, item := range room.InventoryItems {
				if item.Condition == models.InventoryConditionDamaged || item.Condition == models.InventoryConditionBroken {
					hasDamagedInventory = true
					break
				}
			}
			if hasDamagedInventory {
				filtered = append(filtered, room)
			}
		default:
			filtered = append(filtered, room)
		}
	}
	return filtered, nil
}

func (s *Service) GetRoom(roomID uuid.UUID) (models.Room, error) {
	return s.store.GetRoom(roomID)
}

func (s *Service) AssignTenant(principal Principal, tenantID, roomID uuid.UUID) (models.RoomAssignment, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.RoomAssignment{}, ErrUnauthorized
	}

	room, err := s.store.GetRoom(roomID)
	if err != nil {
		return models.RoomAssignment{}, err
	}

	tenant, err := s.store.GetTenant(tenantID)
	if err != nil {
		return models.RoomAssignment{}, err
	}
	if !tenant.IsActive {
		return models.RoomAssignment{}, ErrValidation
	}

	var created models.RoomAssignment
	err = s.store.WithTx(func(tx *gorm.DB) error {
		count, countErr := s.store.GetActiveAssignmentCount(room.ID)
		if countErr != nil {
			return countErr
		}
		if count >= int64(room.Capacity) {
			return ErrCapacityConflict
		}

		if closeErr := s.store.CloseActiveAssignmentByTenant(tx, tenant.ID, time.Now().UTC()); closeErr != nil {
			return closeErr
		}

		assignment := models.RoomAssignment{
			TenantID:        tenant.ID,
			RoomID:          room.ID,
			EffectiveAt:     time.Now().UTC(),
			CreatedByUserID: principal.UserID,
		}
		var createErr error
		created, createErr = s.store.CreateAssignment(tx, assignment)
		return createErr
	})
	if err != nil {
		return models.RoomAssignment{}, err
	}

	_ = s.store.CreateAudit(newAudit(principal.UserID, "room.assignment", "room_assignment", created.ID, models.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) GenerateAllocationPlan() (AssignmentPlan, error) {
	tenants, err := s.store.ListUnassignedActiveTenants()
	if err != nil {
		return AssignmentPlan{}, err
	}
	rooms, err := s.store.ListAllRoomsForPlanning()
	if err != nil {
		return AssignmentPlan{}, err
	}

	type roomCapacity struct {
		id       uuid.UUID
		capacity int64
		used     int64
	}
	roomCaps := make([]roomCapacity, 0, len(rooms))
	for _, room := range rooms {
		used, countErr := s.store.GetActiveAssignmentCount(room.ID)
		if countErr != nil {
			return AssignmentPlan{}, countErr
		}
		roomCaps = append(roomCaps, roomCapacity{id: room.ID, capacity: int64(room.Capacity), used: used})
	}

	items := make([]AssignmentPlanItem, 0, len(tenants))
	roomIdx := 0
	for _, tenant := range tenants {
		for roomIdx < len(roomCaps) && roomCaps[roomIdx].used >= roomCaps[roomIdx].capacity {
			roomIdx++
		}
		if roomIdx >= len(roomCaps) {
			break
		}
		items = append(items, AssignmentPlanItem{TenantID: tenant.ID, RoomID: roomCaps[roomIdx].id})
		roomCaps[roomIdx].used++
	}

	return AssignmentPlan{Items: items}, nil
}

func (s *Service) ApproveAllocationPlan(principal Principal, plan AssignmentPlan) error {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return ErrUnauthorized
	}
	return s.store.WithTx(func(tx *gorm.DB) error {
		for _, item := range plan.Items {
			count, err := s.store.GetActiveAssignmentCount(item.RoomID)
			if err != nil {
				return err
			}
			room, err := s.store.GetRoom(item.RoomID)
			if err != nil {
				return err
			}
			if count >= int64(room.Capacity) {
				return fmt.Errorf("%w: room %s", ErrCapacityConflict, room.ID)
			}
			if err := s.store.CloseActiveAssignmentByTenant(tx, item.TenantID, time.Now().UTC()); err != nil {
				return err
			}
			_, err = s.store.CreateAssignment(tx, models.RoomAssignment{
				TenantID:        item.TenantID,
				RoomID:          item.RoomID,
				EffectiveAt:     time.Now().UTC(),
				CreatedByUserID: principal.UserID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) CreateInventoryItem(principal Principal, item models.InventoryItem) (models.InventoryItem, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector) {
		return models.InventoryItem{}, ErrUnauthorized
	}
	if item.LocationType == models.InventoryLocationRoom && item.RoomID == nil {
		return models.InventoryItem{}, ErrValidation
	}
	if item.LocationType == models.InventoryLocationSharedArea && item.BuildingID == nil && item.FlatID == nil {
		return models.InventoryItem{}, ErrValidation
	}
	created, err := s.store.CreateInventoryItem(item)
	if err != nil {
		return models.InventoryItem{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "inventory.create", "inventory_item", created.ID, models.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) UpdateInventoryStatus(principal Principal, id uuid.UUID, status models.InventoryStatus, condition models.InventoryCondition, withdrawDate *time.Time) (models.InventoryItem, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector) {
		return models.InventoryItem{}, ErrUnauthorized
	}
	item, err := s.store.UpdateInventoryStatus(id, status, condition, withdrawDate)
	if err != nil {
		return models.InventoryItem{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "inventory.status_change", "inventory_item", item.ID, models.AuditOutcomeSuccess))
	return item, nil
}

func (s *Service) ListInventory() ([]models.InventoryItem, error) {
	return s.store.ListInventory()
}

func (s *Service) CreateMaintenanceTicket(principal Principal, ticket models.MaintenanceTicket) (models.MaintenanceTicket, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector) {
		return models.MaintenanceTicket{}, ErrUnauthorized
	}
	ticket.Status = models.MaintenanceStatusReported
	created, err := s.store.CreateMaintenanceTicket(ticket)
	if err != nil {
		return models.MaintenanceTicket{}, err
	}
	_ = s.store.CreateTicketStatusChange(models.TicketStatusChange{
		TicketID:    created.ID,
		ActorUserID: principal.UserID,
		FromStatus:  nil,
		ToStatus:    models.MaintenanceStatusReported,
		Note:        "ticket created",
	})
	return created, nil
}

func (s *Service) ApproveMaintenanceTicket(principal Principal, ticketID uuid.UUID, assigneeID *uuid.UUID, dueAt *time.Time) (models.MaintenanceTicket, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector) {
		return models.MaintenanceTicket{}, ErrUnauthorized
	}
	ticket, err := s.store.GetMaintenanceTicket(ticketID)
	if err != nil {
		return models.MaintenanceTicket{}, err
	}
	from := ticket.Status
	if from != models.MaintenanceStatusReported {
		return models.MaintenanceTicket{}, ErrStateTransition
	}
	ticket.Status = models.MaintenanceStatusInProgress
	ticket.AssigneeUserID = assigneeID
	ticket.DueAt = dueAt
	updated, err := s.store.UpdateMaintenanceTicket(ticket)
	if err != nil {
		return models.MaintenanceTicket{}, err
	}
	_ = s.store.CreateTicketStatusChange(models.TicketStatusChange{
		TicketID:    updated.ID,
		ActorUserID: principal.UserID,
		FromStatus:  &from,
		ToStatus:    updated.Status,
		Note:        "ticket approved",
	})
	return updated, nil
}

func (s *Service) TransitionMaintenanceTicket(principal Principal, ticketID uuid.UUID, toStatus models.MaintenanceStatus, note string) (models.MaintenanceTicket, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector) {
		return models.MaintenanceTicket{}, ErrUnauthorized
	}
	ticket, err := s.store.GetMaintenanceTicket(ticketID)
	if err != nil {
		return models.MaintenanceTicket{}, err
	}
	if !isAllowedTransition(ticket.Status, toStatus) {
		return models.MaintenanceTicket{}, ErrStateTransition
	}
	from := ticket.Status
	ticket.Status = toStatus
	if toStatus == models.MaintenanceStatusResolved || toStatus == models.MaintenanceStatusClosed {
		now := time.Now().UTC()
		ticket.ClosedAt = &now
	}
	updated, err := s.store.UpdateMaintenanceTicket(ticket)
	if err != nil {
		return models.MaintenanceTicket{}, err
	}
	_ = s.store.CreateTicketStatusChange(models.TicketStatusChange{
		TicketID:    updated.ID,
		ActorUserID: principal.UserID,
		FromStatus:  &from,
		ToStatus:    toStatus,
		Note:        note,
	})
	return updated, nil
}

func (s *Service) ListMaintenanceTickets(filter TicketListFilter) ([]models.MaintenanceTicket, error) {
	return s.store.ListMaintenanceTickets(filter)
}

func (s *Service) CreateOperationalJob(principal Principal, job models.OperationalJob, allowConflict bool) (CreateJobResult, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleDirector) {
		return CreateJobResult{}, ErrUnauthorized
	}
	if !job.EndsAt.After(job.StartsAt) {
		return CreateJobResult{}, ErrValidation
	}
	conflicts, err := s.store.ListJobConflicts(job.AssigneeUserID, job.StartsAt, job.EndsAt)
	if err != nil {
		return CreateJobResult{}, err
	}
	mapped := make([]JobConflict, 0, len(conflicts))
	for _, conflict := range conflicts {
		mapped = append(mapped, JobConflict{
			JobID:    conflict.ID,
			StartsAt: conflict.StartsAt,
			EndsAt:   conflict.EndsAt,
		})
	}
	if len(mapped) > 0 && !allowConflict {
		return CreateJobResult{Conflicts: mapped}, ErrConcurrencyConflict
	}
	job.CreatedByUserID = &principal.UserID
	created, err := s.store.CreateOperationalJob(job)
	if err != nil {
		return CreateJobResult{}, err
	}
	return CreateJobResult{Job: created, Conflicts: mapped}, nil
}

func (s *Service) ListOperationalJobs(filter JobListFilter) ([]models.OperationalJob, error) {
	return s.store.ListOperationalJobs(filter)
}

func (s *Service) CreateNews(principal Principal, input NewsUpsertInput) (forummodels.ForumPost, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return forummodels.ForumPost{}, ErrUnauthorized
	}
	post := forummodels.ForumPost{
		AuthorUserID: principal.UserID,
		Kind:         forummodels.ForumPostKindOfficialNews,
		Source:       forummodels.ForumPostSourceAdministration,
		State:        input.State,
		Title:        input.Title,
		Body:         input.Body,
		Tags:         input.Tags,
		PublishedAt:  input.PublishDate,
	}
	if post.State == "" {
		post.State = forummodels.ForumPostStateDraft
	}
	created, err := s.store.CreateForumPost(post)
	if err != nil {
		return forummodels.ForumPost{}, err
	}
	if created.State == forummodels.ForumPostStatePublished {
		_ = s.store.CreateAudit(newAudit(principal.UserID, "news.publish", "forum_post", created.ID, models.AuditOutcomeSuccess))
	}
	return created, nil
}

func (s *Service) PublishNews(principal Principal, id uuid.UUID) (forummodels.ForumPost, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return forummodels.ForumPost{}, ErrUnauthorized
	}
	post, err := s.store.GetForumPost(id)
	if err != nil {
		return forummodels.ForumPost{}, err
	}
	now := time.Now().UTC()
	post.State = forummodels.ForumPostStatePublished
	post.PublishedAt = &now
	updated, err := s.store.UpdateForumPost(post)
	if err != nil {
		return forummodels.ForumPost{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "news.publish", "forum_post", updated.ID, models.AuditOutcomeSuccess))
	return updated, nil
}

func (s *Service) ListNews() ([]forummodels.ForumPost, error) {
	return s.store.ListForumPosts()
}

func (s *Service) CreateActivity(principal Principal, input ActivityUpsertInput) (models.Activity, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Activity{}, ErrUnauthorized
	}
	activity := models.Activity{
		Title:       input.Title,
		Description: input.Description,
		BuildingID:  input.BuildingID,
		Location:    input.Location,
		Capacity:    input.Capacity,
		State:       input.State,
	}
	if activity.State == "" {
		activity.State = models.PublicationStateDraft
	}
	created, err := s.store.CreateActivity(activity)
	if err != nil {
		return models.Activity{}, err
	}
	if created.State == models.PublicationStatePublished {
		_ = s.store.CreateAudit(newAudit(principal.UserID, "activity.publish", "activity", created.ID, models.AuditOutcomeSuccess))
	}
	return created, nil
}

func (s *Service) PublishActivity(principal Principal, id uuid.UUID) (models.Activity, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Activity{}, ErrUnauthorized
	}
	activity, err := s.store.GetActivity(id)
	if err != nil {
		return models.Activity{}, err
	}
	activity.State = models.PublicationStatePublished
	updated, err := s.store.UpdateActivity(activity)
	if err != nil {
		return models.Activity{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "activity.publish", "activity", updated.ID, models.AuditOutcomeSuccess))
	return updated, nil
}

func (s *Service) ListActivities() ([]models.Activity, error) {
	return s.store.ListActivities()
}

func (s *Service) CreateEvent(principal Principal, input EventUpsertInput) (models.Event, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Event{}, ErrUnauthorized
	}
	if !input.EndsAt.After(input.StartsAt) {
		return models.Event{}, ErrValidation
	}
	event := models.Event{
		Title:           input.Title,
		Description:     input.Description,
		BuildingID:      input.BuildingID,
		OrganizerUserID: input.OrganizerUserID,
		Location:        input.Location,
		StartsAt:        input.StartsAt,
		EndsAt:          input.EndsAt,
		Capacity:        input.Capacity,
		State:           input.State,
	}
	if event.State == "" {
		event.State = models.PublicationStateDraft
	}
	created, err := s.store.CreateEvent(event)
	if err != nil {
		return models.Event{}, err
	}
	if created.State == models.PublicationStatePublished {
		_ = s.store.CreateAudit(newAudit(principal.UserID, "event.publish", "event", created.ID, models.AuditOutcomeSuccess))
	}
	return created, nil
}

func (s *Service) UpdateEventState(principal Principal, eventID uuid.UUID, state models.PublicationState) (models.Event, error) {
	if !hasAnyRole(principal, models.RoleAdministrator, models.RoleOfficeWorker) {
		return models.Event{}, ErrUnauthorized
	}
	switch state {
	case models.PublicationStateDraft, models.PublicationStatePublished, models.PublicationStateArchived, models.PublicationStatePostponed, models.PublicationStateCanceled:
	default:
		return models.Event{}, ErrValidation
	}

	event, err := s.store.GetEvent(eventID)
	if err != nil {
		return models.Event{}, err
	}
	event.State = state
	updated, err := s.store.UpdateEvent(event)
	if err != nil {
		return models.Event{}, err
	}
	if state == models.PublicationStatePublished || state == models.PublicationStatePostponed || state == models.PublicationStateCanceled {
		_ = s.store.CreateAudit(newAudit(principal.UserID, "event.state_change", "event", updated.ID, models.AuditOutcomeSuccess))
	}
	return updated, nil
}

func (s *Service) ListEvents() ([]models.Event, error) {
	return s.store.ListEvents()
}

func hasAnyRole(principal Principal, required ...models.RoleName) bool {
	for _, role := range principal.Roles {
		for _, req := range required {
			if role == req {
				return true
			}
		}
	}
	return false
}

func newAudit(actorID uuid.UUID, action, targetType string, targetID uuid.UUID, outcome models.AuditOutcome) models.AuditEvent {
	return models.AuditEvent{
		ActorUserID: &actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Outcome:     outcome,
		OccurredAt:  time.Now().UTC(),
	}
}

func isAllowedTransition(from, to models.MaintenanceStatus) bool {
	allowed := map[models.MaintenanceStatus]map[models.MaintenanceStatus]struct{}{
		models.MaintenanceStatusReported: {
			models.MaintenanceStatusInProgress: {},
			models.MaintenanceStatusDuplicate:  {},
		},
		models.MaintenanceStatusInProgress: {
			models.MaintenanceStatusHalted:   {},
			models.MaintenanceStatusResolved: {},
		},
		models.MaintenanceStatusHalted: {
			models.MaintenanceStatusInProgress: {},
			models.MaintenanceStatusClosed:     {},
		},
		models.MaintenanceStatusResolved: {
			models.MaintenanceStatusClosed: {},
		},
	}
	next, ok := allowed[from]
	if !ok {
		return false
	}
	_, ok = next[to]
	return ok
}

func classifyError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrValidation), errors.Is(err, ErrStudentStatusInvalid):
		return "validation_error"
	case errors.Is(err, ErrCapacityConflict):
		return "capacity_conflict"
	case errors.Is(err, ErrStateTransition):
		return "state_transition_invalid"
	case errors.Is(err, ErrUnauthorized):
		return "authorization_denied"
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case errors.Is(err, ErrConcurrencyConflict):
		return "concurrency_conflict"
	default:
		return "internal_error"
	}
}
