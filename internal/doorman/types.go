package doorman

import (
	"dorm-man/internal/models"
	"errors"
	"time"

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

type GuestData struct {
	GuestID    uuid.UUID    `json:"guest_id"`
	HostTenant *models.User `json:"host_tenant"`
	GuestName  string       `json:"guest_name"`
	IDNotes    string       `json:"id_notes"`
}

type GuestRegisterRequest struct {
	HostTenantID uuid.UUID `json:"host_tenant_id" form:"host_tenant_id"`
	GuestName    string    `json:"guest_name"     form:"guest_name"`
	IDNotes      string    `json:"id_notes"       form:"id_notes"`
}

type GuestRegisterResponse struct {
	Success bool `json:"success"`
}

type TenantData struct {
	Tenant     *models.Tenant `json:"tenant"`
	CheckedIn  time.Time      `json:"checked_in"`
	CheckedOut time.Time      `json:"checked_out"`
}

type TenantAccessRequest struct {
	TenantID     uuid.UUID           `json:"tenant_id"     query:"id"`
	AccessStatus models.AccessStatus `json:"access_status" query:"direction"`
}

type TenantAccessResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type GuestFilter struct {
	TenantID uuid.UUID           `json:"tenant_id" query:"tenant_id"`
	From     time.Time           `json:"from"      query:"from"`
	To       time.Time           `json:"to"        query:"to"`
	Status   models.AccessStatus `json:"status"    query:"status"`
}

type TenantAccessFilter struct {
	TenantID uuid.UUID           `json:"tenant_id" query:"tenant_id"`
	From     time.Time           `json:"from"      query:"from"`
	To       time.Time           `json:"to"        query:"to"`
	Status   models.AccessStatus `json:"status"    query:"status"`
}
