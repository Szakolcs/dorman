package doorman

import (
	adm "dorm-man/internal/models/administration"
	crosscutting "dorm-man/internal/models/cross-cutting"
	models "dorm-man/internal/models/doorman"
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
	GuestID    uuid.UUID          `json:"guest_id"`
	HostTenant *crosscutting.User `json:"host_tenant_id"`
	GuestName  string             `json:"guest_name"`
	IDNotes    string             `json:"id_notes"`
}

type GuestRegisterRequest struct {
	crosscutting.BaseModel
	HostTenant *crosscutting.User `json:"host_tenant_id"`
	GuestName  string             `json:"guest_name"`
	IDNotes    string             `json:"id_notes"`
}

type GuestRegisterResponse struct {
	Success bool `json:"success"`
}

type TenantData struct {
	Tenant     *adm.Tenant `json:"tenant"`
	CheckedIn  time.Time   `json:"checked_in"`
	CheckedOut time.Time   `json:"checked_out"`
}
type TenantAccessRequest struct {
	TenantID     uuid.UUID           `json:"tenant_id"`
	AccessStatus models.AccessStatus `json:"accessstatus"`
}

type TenantAccessResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type GuestFilter struct {
	TenantID uuid.UUID           `json:"tenant_id"`
	From     time.Time           `json:"from"`
	To       time.Time           `json:"to"`
	Status   models.AccessStatus `json:"status"`
}

type TenantAccessFilter struct {
	TenantID uuid.UUID           `json:"tenant_id"`
	From     time.Time           `json:"from"`
	To       time.Time           `json:"to"`
	Status   models.AccessStatus `json:"status"`
}
