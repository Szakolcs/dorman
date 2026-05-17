package doorman

import (
	"errors"
	"time"

	"dorm-man/internal/pagination"
	adm "dorm-man/internal/models/administration"
	dm "dorm-man/internal/models/doorman"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized              = errors.New("authorization denied")
	ErrNotFound                  = errors.New("not found")
	ErrValidation                = errors.New("validation error")
	ErrOutsideVisitWindow        = errors.New("outside_visit_window")
	ErrPackageInvalidTransition  = errors.New("package_invalid_transition")
	ErrLoanItemUnavailable       = errors.New("loan_item_unavailable")
	ErrTenantNotFound            = errors.New("tenant_not_found")
	ErrGuestVisitStateTransition = errors.New("guest_visit_state_invalid")
	ErrLoanAlreadyReturned       = errors.New("loan_already_returned")
)

type PackageRegisterInput struct {
	RecipientLabel string     `json:"recipient_label"`
	Description    string     `json:"description"`
	TenantID       *uuid.UUID `json:"tenant_id"`
	ReceivedAt     *time.Time `json:"received_at"`
}

type PackageNotifyInput struct {
	Channel dm.PackageNotificationChannel `json:"channel"`
}

type TransitionPackageBody struct {
	To dm.PackageStatus `json:"to"`
}

type GuestVisitCreateInput struct {
	HostTenantID uuid.UUID `json:"host_tenant_id"`
	GuestName    string    `json:"guest_name"`
	IDNotes      string    `json:"id_notes"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
}

type QRValidateBody struct {
	PublicRef string `json:"public_ref"`
}

type ManualAccessBody struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type TenantEntryTokenCreateInput struct {
	TenantID  uuid.UUID  `json:"tenant_id"`
	PublicRef string     `json:"public_ref"` // optional; server generates UUID string if empty
	ExpiresAt *time.Time `json:"expires_at"`
}

type CheckoutLoanBody struct {
	TenantID         uuid.UUID  `json:"tenant_id"`
	InventoryItemID  uuid.UUID  `json:"inventory_item_id"`
	ExpectedReturnAt *time.Time `json:"expected_return_at"`
	Notes            string     `json:"notes"`
}

type PackageListFilter struct {
	Status   string // received|notified|picked_up|pending (queued = not picked up)
	TenantID *uuid.UUID
	pagination.Params
}

type GuestVisitListFilter struct {
	OnDate *time.Time // local date compare in UTC midnight window
	pagination.Params
}

type AccessEventListFilter struct {
	TenantID *uuid.UUID
	From     *time.Time
	To       *time.Time
	Outcome  dm.AccessEventOutcome
	pagination.Params
}

type TokenListFilter struct {
	pagination.Params
}

type ItemLoanListFilter struct {
	OpenOnly    bool
	OverdueOnly bool
	TenantID    *uuid.UUID
	pagination.Params
}

type QRValidationResult struct {
	Outcome dm.AccessEventOutcome `json:"outcome"`
	Reason  string                `json:"reason,omitempty"`
	Tenant  *adm.Tenant           `json:"tenant,omitempty"`
	Event   dm.AccessEvent        `json:"event"`
}
