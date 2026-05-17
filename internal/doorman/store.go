package doorman

import (
	"errors"
	"fmt"
	"time"

	"dorm-man/internal/administration"

	adm "dorm-man/internal/models/administration"
	dm "dorm-man/internal/models/doorman"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	LoadPrincipal(userID uuid.UUID) (administration.Principal, error)
	GetTenant(tenantID uuid.UUID) (adm.Tenant, error)
	GetInventoryItem(id uuid.UUID) (adm.InventoryItem, error)
	CreateAudit(event adm.AuditEvent) error
	Transaction(fn func(tx *gorm.DB) error) error

	CreatePackage(p dm.Package) (dm.Package, error)
	UpdatePackage(pkg dm.Package) (dm.Package, error)
	GetPackage(id uuid.UUID) (dm.Package, error)
	ListPackages(filter PackageListFilter) ([]dm.Package, error)

	CreatePackageNotification(n dm.PackageNotification) (dm.PackageNotification, error)
	UpdatePackageNotification(n dm.PackageNotification) (dm.PackageNotification, error)

	CreateGuestVisit(v dm.GuestVisit) (dm.GuestVisit, error)
	UpdateGuestVisit(v dm.GuestVisit) (dm.GuestVisit, error)
	GetGuestVisit(id uuid.UUID) (dm.GuestVisit, error)
	CreateGuestAccessEvent(e dm.GuestAccessEvent) (dm.GuestAccessEvent, error)
	ListGuestVisits(filter GuestVisitListFilter) ([]dm.GuestVisit, error)

	CreateTenantEntryToken(t dm.TenantEntryToken) (dm.TenantEntryToken, error)
	GetTenantEntryTokenByPublicRef(ref string) (dm.TenantEntryToken, error)
	CreateAccessEvent(e dm.AccessEvent) (dm.AccessEvent, error)
	ListAccessEvents(filter AccessEventListFilter) ([]dm.AccessEvent, error)

	CreateItemLoan(loan dm.ItemLoan) (dm.ItemLoan, error)
	UpdateItemLoan(loan dm.ItemLoan) (dm.ItemLoan, error)
	GetItemLoan(id uuid.UUID) (dm.ItemLoan, error)
	CountOpenLoansForInventory(inventoryItemID uuid.UUID) (int64, error)
	ListItemLoans(filter ItemLoanListFilter) ([]dm.ItemLoan, error)
}

type GormStore struct {
	db      *gorm.DB
	housing administration.Store
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, housing: administration.NewStore(db)}
}

func (s *GormStore) LoadPrincipal(userID uuid.UUID) (administration.Principal, error) {
	return s.housing.LoadPrincipal(userID)
}

func (s *GormStore) GetTenant(tenantID uuid.UUID) (adm.Tenant, error) {
	return s.housing.GetTenant(tenantID)
}

func (s *GormStore) GetInventoryItem(id uuid.UUID) (adm.InventoryItem, error) {
	var item adm.InventoryItem
	err := s.db.First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return adm.InventoryItem{}, ErrNotFound
	}
	return item, err
}

func (s *GormStore) CreateAudit(event adm.AuditEvent) error {
	return s.housing.CreateAudit(event)
}

func (s *GormStore) Transaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

func (s *GormStore) CreatePackage(p dm.Package) (dm.Package, error) {
	if err := s.db.Create(&p).Error; err != nil {
		return dm.Package{}, err
	}
	return p, nil
}

func (s *GormStore) UpdatePackage(pkg dm.Package) (dm.Package, error) {
	if err := s.db.Save(&pkg).Error; err != nil {
		return dm.Package{}, err
	}
	return pkg, nil
}

func (s *GormStore) GetPackage(id uuid.UUID) (dm.Package, error) {
	var p dm.Package
	err := s.db.Preload("Tenant").
		Preload("RegisteredByUser").
		Preload("PickedUpByUser").
		First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm.Package{}, ErrNotFound
	}
	return p, err
}

func (s *GormStore) ListPackages(filter PackageListFilter) ([]dm.Package, error) {
	var list []dm.Package
	q := s.db.Model(&dm.Package{})
	switch filter.Status {
	case "pending":
		q = q.Where("status <> ?", dm.PackageStatusPickedUp)
	case string(dm.PackageStatusReceived), string(dm.PackageStatusNotified), string(dm.PackageStatusPickedUp):
		q = q.Where("status = ?", dm.PackageStatus(filter.Status))
	default:
		if filter.Status != "" {
			q = q.Where("status = ?", filter.Status)
		}
	}
	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	err := q.Order("received_at DESC").Limit(limit).Offset(offset).
		Preload("Tenant").
		Find(&list).Error
	return list, err
}

func (s *GormStore) CreatePackageNotification(n dm.PackageNotification) (dm.PackageNotification, error) {
	if err := s.db.Create(&n).Error; err != nil {
		return dm.PackageNotification{}, err
	}
	return n, nil
}

func (s *GormStore) UpdatePackageNotification(n dm.PackageNotification) (dm.PackageNotification, error) {
	if err := s.db.Save(&n).Error; err != nil {
		return dm.PackageNotification{}, err
	}
	return n, nil
}

func (s *GormStore) CreateGuestVisit(v dm.GuestVisit) (dm.GuestVisit, error) {
	if err := s.db.Create(&v).Error; err != nil {
		return dm.GuestVisit{}, err
	}
	return v, nil
}

func (s *GormStore) UpdateGuestVisit(v dm.GuestVisit) (dm.GuestVisit, error) {
	if err := s.db.Save(&v).Error; err != nil {
		return dm.GuestVisit{}, err
	}
	return v, nil
}

func (s *GormStore) GetGuestVisit(id uuid.UUID) (dm.GuestVisit, error) {
	var v dm.GuestVisit
	err := s.db.Preload("HostTenant").Preload("AccessEvents").
		First(&v, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm.GuestVisit{}, ErrNotFound
	}
	return v, err
}

func (s *GormStore) CreateGuestAccessEvent(e dm.GuestAccessEvent) (dm.GuestAccessEvent, error) {
	if err := s.db.Create(&e).Error; err != nil {
		return dm.GuestAccessEvent{}, err
	}
	return e, nil
}

func (s *GormStore) ListGuestVisits(filter GuestVisitListFilter) ([]dm.GuestVisit, error) {
	var list []dm.GuestVisit
	q := s.db.Model(&dm.GuestVisit{})
	if filter.OnDate != nil {
		d := filter.OnDate.UTC()
		start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		end := start.Add(24 * time.Hour)
		q = q.Where("valid_from < ? AND valid_to >= ?", end, start)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	err := q.Order("valid_from ASC").
		Preload("HostTenant").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	return list, err
}

func (s *GormStore) CreateTenantEntryToken(t dm.TenantEntryToken) (dm.TenantEntryToken, error) {
	if err := s.db.Create(&t).Error; err != nil {
		return dm.TenantEntryToken{}, err
	}
	return t, nil
}

func (s *GormStore) GetTenantEntryTokenByPublicRef(ref string) (dm.TenantEntryToken, error) {
	var t dm.TenantEntryToken
	err := s.db.Preload("Tenant").
		First(&t, "public_ref = ?", ref).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm.TenantEntryToken{}, ErrNotFound
	}
	return t, err
}

func (s *GormStore) CreateAccessEvent(e dm.AccessEvent) (dm.AccessEvent, error) {
	if err := s.db.Create(&e).Error; err != nil {
		return dm.AccessEvent{}, fmt.Errorf("create access event: %w", err)
	}
	return e, nil
}

func (s *GormStore) ListAccessEvents(filter AccessEventListFilter) ([]dm.AccessEvent, error) {
	var list []dm.AccessEvent
	q := s.db.Model(&dm.AccessEvent{})
	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.From != nil {
		q = q.Where("occurred_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("occurred_at <= ?", *filter.To)
	}
	if filter.Outcome != "" {
		q = q.Where("outcome = ?", filter.Outcome)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	err := q.Order("occurred_at DESC").
		Preload("Tenant").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	return list, err
}

func (s *GormStore) CreateItemLoan(loan dm.ItemLoan) (dm.ItemLoan, error) {
	if err := s.db.Create(&loan).Error; err != nil {
		return dm.ItemLoan{}, err
	}
	return loan, nil
}

func (s *GormStore) UpdateItemLoan(loan dm.ItemLoan) (dm.ItemLoan, error) {
	if err := s.db.Save(&loan).Error; err != nil {
		return dm.ItemLoan{}, err
	}
	return loan, nil
}

func (s *GormStore) GetItemLoan(id uuid.UUID) (dm.ItemLoan, error) {
	var loan dm.ItemLoan
	err := s.db.Preload("Tenant").
		Preload("InventoryItem").
		Preload("CheckedOutByUser").
		Preload("ReturnedByUser").
		First(&loan, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm.ItemLoan{}, ErrNotFound
	}
	return loan, err
}

func (s *GormStore) CountOpenLoansForInventory(inventoryItemID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.Model(&dm.ItemLoan{}).
		Where("inventory_item_id = ? AND returned_at IS NULL", inventoryItemID).
		Count(&n).Error
	return n, err
}

func (s *GormStore) ListItemLoans(filter ItemLoanListFilter) ([]dm.ItemLoan, error) {
	var list []dm.ItemLoan
	q := s.db.Model(&dm.ItemLoan{})
	if filter.OpenOnly || filter.OverdueOnly {
		q = q.Where("returned_at IS NULL")
	}
	if filter.OverdueOnly {
		q = q.Where("expected_return_at IS NOT NULL AND expected_return_at < ?", time.Now().UTC())
	}
	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	err := q.Order("checked_out_at DESC").
		Preload("Tenant").
		Preload("InventoryItem").
		Limit(limit).
		Offset(offset).
		Find(&list).Error
	return list, err
}
