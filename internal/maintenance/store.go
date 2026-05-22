package maintenance

import (
	"dorm-man/internal/administration"
	adm "dorm-man/internal/models/administration"
	models "dorm-man/internal/models/maintenance"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	ListTickets(filter TicketFilter) ([]models.Ticket, error)
	ListTicketsByUser(userID uuid.UUID, filter TicketFilter) ([]models.Ticket, error)
	GetTicketByID(id uuid.UUID) (models.Ticket, error)
	CreateTicket(actorID uuid.UUID, ticket models.Ticket) error
	UpdateTicket(actorID uuid.UUID, ticket models.Ticket, note string) error
	DeleteTicket(actorID, id uuid.UUID) error
	GetTenantLocation(userID uuid.UUID) (TenantLocation, error)
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func applyTicketFilters(q *gorm.DB, filter TicketFilter) *gorm.DB {
	if filter.FlatID != nil {
		q = q.Where("flat_id = ?", filter.FlatID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if !filter.From.IsZero() {
		q = q.Where("created_at >= ?", filter.From)
	}
	if !filter.To.IsZero() {
		q = q.Where("created_at <= ?", filter.To)
	}
	return q
}

func (s *GormStore) ListTickets(filter TicketFilter) ([]models.Ticket, error) {
	q := applyTicketFilters(s.db.Model(&models.Ticket{}), filter)
	var tickets []models.Ticket
	err := q.
		Preload("CreatedByUser").
		Order("created_at DESC").
		Find(&tickets).
		Error
	return tickets, err
}

func (s *GormStore) ListTicketsByUser(userID uuid.UUID, filter TicketFilter) ([]models.Ticket, error) {
	q := applyTicketFilters(
		s.db.Model(&models.Ticket{}).Where("created_by_user_id = ?", userID),
		filter,
	)
	var tickets []models.Ticket
	err := q.
		Preload("CreatedByUser").
		Order("created_at DESC").
		Find(&tickets).
		Error
	return tickets, err
}

func (s *GormStore) GetTicketByID(id uuid.UUID) (models.Ticket, error) {
	var ticket models.Ticket
	err := s.db.
		Preload("CreatedByUser").
		Preload("StatusTransitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&ticket, "id = ?", id).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Ticket{}, ErrNotFound
		}
		return models.Ticket{}, err
	}
	return ticket, nil
}

func (s *GormStore) CreateTicket(actorID uuid.UUID, ticket models.Ticket) error {
	return administration.AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}
		change := models.StatusChange{
			TicketID: ticket.ID,
			ToStatus: ticket.Status,
			Note:     "Ticket created",
		}
		return tx.Create(&change).Error
	})
}

func (s *GormStore) UpdateTicket(actorID uuid.UUID, ticket models.Ticket, note string) error {
	return administration.AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		var existing models.Ticket
		if err := tx.First(&existing, "id = ?", ticket.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNotFound
			}
			return err
		}

		fromStatus := existing.Status
		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		if ticket.Status != fromStatus {
			change := models.StatusChange{
				TicketID:   ticket.ID,
				FromStatus: &fromStatus,
				ToStatus:   ticket.Status,
				Note:       note,
			}
			if err := tx.Create(&change).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GormStore) DeleteTicket(actorID, id uuid.UUID) error {
	return administration.AuditedTransaction(s.db, actorID, func(tx *gorm.DB) error {
		result := tx.Delete(&models.Ticket{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *GormStore) GetTenantLocation(userID uuid.UUID) (TenantLocation, error) {
	var tenant adm.Tenant
	err := s.db.
		Preload("RoomAssignments", "ended_at IS NULL").
		Preload("RoomAssignments.Room").
		Where("user_id = ?", userID).
		First(&tenant).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return TenantLocation{}, nil
		}
		return TenantLocation{}, err
	}

	for _, assignment := range tenant.RoomAssignments {
		if assignment.EndedAt != nil {
			continue
		}
		var room adm.Room
		if err := s.db.First(&room, "id = ?", assignment.RoomID).Error; err != nil {
			continue
		}
		roomID := assignment.RoomID
		flatID := room.FlatID
		return TenantLocation{
			RoomID: &roomID,
			FlatID: &flatID,
		}, nil
	}
	return TenantLocation{}, nil
}
