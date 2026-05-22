package doorman

import (
	adm "dorm-man/internal/models"

	"github.com/google/uuid"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}
func (s *Service) GetTenants() ([]adm.Tenant, error) {
	return s.store.getTenants()
}

func (s *Service) ListTenantAccess(filter TenantAccessFilter) ([]adm.TenantEntry, error) {
	return s.store.ListTenantAccess(filter)
}

func (s *Service) ListGuests(filter GuestFilter) ([]adm.GuestEntry, error) {
	return s.store.ListGuests(filter)
}

func (s *Service) RegisterGuest(req GuestRegisterRequest) (*GuestRegisterResponse, error) {
	if req.HostTenantID == uuid.Nil {
		return nil, ErrTenantNotFound
	}
	if req.GuestName == "" {
		return nil, ErrValidation
	}
	guest := GuestData{
		HostTenant: &adm.User{
			BaseModel: adm.BaseModel{ID: req.HostTenantID},
		},
		GuestName: req.GuestName,
		IDNotes:   req.IDNotes,
	}
	err := s.store.RegisterGuest(guest)
	if err != nil {
		return nil, err
	}
	return &GuestRegisterResponse{
		Success: true,
	}, nil
}

func (s *Service) DeleteGuest(guestID uuid.UUID) error {
	if guestID == uuid.Nil {
		return ErrValidation
	}
	return s.store.DeleteGuest(guestID)
}

func (s *Service) RegisterTenantAccess(req TenantAccessRequest) (*TenantAccessResponse, error) {
	if req.TenantID == uuid.Nil {
		return nil, ErrValidation
	}
	err := s.store.RegisterTenantAccess(
		req.TenantID,
		req.AccessStatus,
	)
	if err != nil {
		return nil, err
	}
	return &TenantAccessResponse{
		TenantID: req.TenantID,
	}, nil
}
