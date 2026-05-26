package maintenance

import (
	"dorm-man/internal/models"

	"github.com/google/uuid"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListTickets(filter TicketFilter) ([]models.Ticket, error) {
	return s.store.ListTickets(filter)
}

func (s *Service) ListTicketsByUser(userID uuid.UUID, filter TicketFilter) ([]models.Ticket, error) {
	if userID == uuid.Nil {
		return nil, ErrValidation
	}
	return s.store.ListTicketsByUser(userID, filter)
}

func (s *Service) ListTicketsForTenantFlat(userID uuid.UUID, filter TicketFilter) ([]models.Ticket, error) {
	if userID == uuid.Nil {
		return nil, ErrValidation
	}
	location, err := s.store.GetTenantLocation(userID)
	if err != nil {
		return nil, err
	}
	if location.FlatID == nil {
		return []models.Ticket{}, nil
	}
	filter.FlatID = location.FlatID
	return s.store.ListTickets(filter)
}

func (s *Service) TenantCanAccessTicket(userID uuid.UUID, ticket models.Ticket) (bool, error) {
	if userID == uuid.Nil || ticket.FlatID == nil {
		return false, nil
	}
	location, err := s.store.GetTenantLocation(userID)
	if err != nil {
		return false, err
	}
	if location.FlatID == nil {
		return false, nil
	}
	return *ticket.FlatID == *location.FlatID, nil
}

func (s *Service) GetTenantLocation(userID uuid.UUID) (TenantLocation, error) {
	return s.store.GetTenantLocation(userID)
}

func (s *Service) GetTicketByID(id uuid.UUID) (models.Ticket, error) {
	if id == uuid.Nil {
		return models.Ticket{}, ErrValidation
	}
	return s.store.GetTicketByID(id)
}

func (s *Service) CreateTicket(actorID uuid.UUID, req CreateTicketRequest) (models.Ticket, error) {
	if req.Description == "" {
		return models.Ticket{}, ErrValidation
	}
	if req.Category == "" || req.Severity == "" || req.Impact == "" {
		return models.Ticket{}, ErrValidation
	}

	location, err := s.store.GetTenantLocation(actorID)
	if err != nil {
		return models.Ticket{}, err
	}

	flatID := req.FlatID
	if flatID == nil {
		flatID = location.FlatID
	}

	ticket := models.Ticket{
		FlatID:          flatID,
		Category:        req.Category,
		Severity:        req.Severity,
		Impact:          req.Impact,
		Status:          models.StatusReported,
		Description:     req.Description,
		CreatedByUserID: actorID,
	}
	if err := s.store.CreateTicket(actorID, ticket); err != nil {
		return models.Ticket{}, err
	}
	return s.store.GetTicketByID(ticket.ID)
}

func (s *Service) UpdateTicket(actorID uuid.UUID, req UpdateTicketRequest) (models.Ticket, error) {
	if req.TicketID == uuid.Nil {
		return models.Ticket{}, ErrValidation
	}
	if req.Status == "" {
		return models.Ticket{}, ErrValidation
	}

	existing, err := s.store.GetTicketByID(req.TicketID)
	if err != nil {
		return models.Ticket{}, err
	}

	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Severity != "" {
		existing.Severity = req.Severity
	}
	if req.Impact != "" {
		existing.Impact = req.Impact
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	existing.Status = req.Status

	if err := s.store.UpdateTicket(actorID, existing, req.Note); err != nil {
		return models.Ticket{}, err
	}
	return s.store.GetTicketByID(existing.ID)
}

func (s *Service) DeleteTicket(actorID, ticketID uuid.UUID) error {
	if ticketID == uuid.Nil {
		return ErrValidation
	}
	return s.store.DeleteTicket(actorID, ticketID)
}

func (s *Service) DeleteOwnTicket(actorID, ticketID uuid.UUID) error {
	if ticketID == uuid.Nil {
		return ErrValidation
	}
	ticket, err := s.store.GetTicketByID(ticketID)
	if err != nil {
		return err
	}
	if ticket.CreatedByUserID != actorID {
		return ErrUnauthorized
	}
	if ticket.Status != models.StatusReported {
		return ErrValidation
	}
	return s.store.DeleteTicket(actorID, ticketID)
}
