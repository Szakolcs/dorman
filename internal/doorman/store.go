package doorman

import (
	"dorm-man/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	ListTenantAccess(filter TenantAccessFilter) ([]models.TenantEntry, error)
	ListGuests(filter GuestFilter) ([]models.GuestEntry, error)
	RegisterGuest(guest GuestData) error
	DeleteGuest(guestID uuid.UUID) error
	RegisterTenantAccess(tenantID uuid.UUID, direction models.AccessStatus) error
	getTenants() ([]models.Tenant, error)
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) getTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.db.Find(&tenants).Error
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *GormStore) ListTenantAccess(filter TenantAccessFilter) ([]models.TenantEntry, error) {
	query := s.db.Model(&models.TenantEntry{})
	if filter.TenantID != uuid.Nil {
		query = query.Where(
			"user_id = ?",
			filter.TenantID,
		)
	}
	if !filter.From.IsZero() {
		query = query.Where(
			"time_of_entry >= ?",
			filter.From,
		)
	}
	if !filter.To.IsZero() {
		query = query.Where(
			"time_of_entry <= ?",
			filter.To,
		)
	}
	if filter.Status != "" {
		query = query.Where(
			"status = ?",
			filter.Status,
		)
	}
	var entries []models.TenantEntry
	err := query.
		Preload("User").
		Order("time_of_entry DESC").
		Find(&entries).
		Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *GormStore) ListGuests(filter GuestFilter) ([]models.GuestEntry, error) {
	query := s.db.Model(&models.GuestEntry{})
	if filter.TenantID != uuid.Nil {
		query = query.Where(
			"host_tenant_id = ?",
			filter.TenantID,
		)
	}
	if !filter.From.IsZero() {
		query = query.Where(
			"created_at >= ?",
			filter.From,
		)
	}
	if !filter.To.IsZero() {
		query = query.Where(
			"created_at <= ?",
			filter.To,
		)
	}
	if filter.Status != "" {
		query = query.Where(
			"status = ?",
			filter.Status,
		)
	}
	var guests []models.GuestEntry
	err := query.
		Preload("HostTenant").
		Order("created_at DESC").
		Find(&guests).
		Error
	if err != nil {
		return nil, err
	}
	return guests, nil
}

func (s *GormStore) RegisterGuest(guest GuestData) error {
	entry := models.GuestEntry{
		HostTenantID: guest.HostTenant.ID,
		GuestName:    guest.GuestName,
		IDNotes:      guest.IDNotes,
	}

	return s.db.Create(&entry).Error
}

func (s *GormStore) DeleteGuest(guestID uuid.UUID) error {
	result := s.db.Delete(&models.GuestEntry{}, "id = ?", guestID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *GormStore) RegisterTenantAccess(tenantID uuid.UUID, accessStatus models.AccessStatus) error {
	entry := models.TenantEntry{
		UserID:      tenantID,
		Status:      accessStatus,
		TimeOfEntry: time.Now(),
	}
	return s.db.Create(&entry).Error
}
