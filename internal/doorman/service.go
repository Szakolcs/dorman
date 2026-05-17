package doorman

import (
	"errors"
	"strings"
	"time"

	"dorm-man/internal/administration"

	adm "dorm-man/internal/models/administration"
	dm "dorm-man/internal/models/doorman"

	"github.com/google/uuid"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ResolvePrincipal(actorID uuid.UUID) (administration.Principal, error) {
	return s.store.LoadPrincipal(actorID)
}

func (s *Service) RequireStaffDoormanPrincipal(p administration.Principal) error {
	if !hasStaffDoorman(p) {
		return ErrUnauthorized
	}
	return nil
}

func (s *Service) RegisterPackage(principal administration.Principal, in PackageRegisterInput) (dm.Package, error) {
	if !hasStaffDoorman(principal) {
		return dm.Package{}, ErrUnauthorized
	}
	if strings.TrimSpace(in.RecipientLabel) == "" {
		return dm.Package{}, ErrValidation
	}
	if in.TenantID != nil {
		if _, err := s.store.GetTenant(*in.TenantID); err != nil {
			if errors.Is(err, administration.ErrNotFound) {
				return dm.Package{}, ErrTenantNotFound
			}
			return dm.Package{}, err
		}
	}

	recvAt := time.Now().UTC()
	if in.ReceivedAt != nil && !in.ReceivedAt.IsZero() {
		recvAt = in.ReceivedAt.UTC()
	}
	pkg := dm.Package{
		RecipientLabel:     strings.TrimSpace(in.RecipientLabel),
		Description:        in.Description,
		TenantID:           in.TenantID,
		Status:             dm.PackageStatusReceived,
		ReceivedAt:         recvAt,
		RegisteredByUserID: &principal.UserID,
	}

	created, err := s.store.CreatePackage(pkg)
	if err != nil {
		return dm.Package{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "package.create", "package", created.ID, adm.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) ListPackages(filter PackageListFilter) ([]dm.Package, int64, error) {
	return s.store.ListPackages(filter)
}

func (s *Service) GetPackage(id uuid.UUID) (dm.Package, error) {
	return s.store.GetPackage(id)
}

func (s *Service) TransitionPackage(principal administration.Principal, packageID uuid.UUID, to dm.PackageStatus) (dm.Package, error) {
	if !hasStaffDoorman(principal) {
		return dm.Package{}, ErrUnauthorized
	}
	pkg, err := s.store.GetPackage(packageID)
	if err != nil {
		return dm.Package{}, err
	}
	if !canTransitionPackage(pkg.Status, to) {
		return dm.Package{}, ErrPackageInvalidTransition
	}
	pkg.Status = to
	updated, err := s.store.UpdatePackage(pkg)
	if err != nil {
		return dm.Package{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "package.status_change", "package", updated.ID, adm.AuditOutcomeSuccess))
	return updated, nil
}

func (s *Service) ConfirmPackagePickup(principal administration.Principal, packageID uuid.UUID) (dm.Package, error) {
	if !hasStaffDoorman(principal) {
		return dm.Package{}, ErrUnauthorized
	}
	pkg, err := s.store.GetPackage(packageID)
	if err != nil {
		return dm.Package{}, err
	}
	if pkg.Status == dm.PackageStatusPickedUp {
		return dm.Package{}, ErrPackageInvalidTransition
	}
	now := time.Now().UTC()
	pkg.Status = dm.PackageStatusPickedUp
	pkg.PickedUpAt = &now
	tmp := principal.UserID
	pkg.PickedUpByUserID = &tmp
	updated, err := s.store.UpdatePackage(pkg)
	if err != nil {
		return dm.Package{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "package.pickup", "package", updated.ID, adm.AuditOutcomeSuccess))
	return updated, nil
}

func (s *Service) NotifyPackageTenant(principal administration.Principal, packageID uuid.UUID, input PackageNotifyInput) (dm.PackageNotification, error) {
	if !hasStaffDoorman(principal) {
		return dm.PackageNotification{}, ErrUnauthorized
	}
	ch := input.Channel
	if ch != dm.PackageNotificationChannelInApp && ch != dm.PackageNotificationChannelEmail {
		ch = dm.PackageNotificationChannelInApp
	}
	pkg, err := s.store.GetPackage(packageID)
	if err != nil {
		return dm.PackageNotification{}, err
	}
	if pkg.TenantID == nil {
		return dm.PackageNotification{}, ErrValidation
	}
	if pkg.Status == dm.PackageStatusPickedUp {
		return dm.PackageNotification{}, ErrPackageInvalidTransition
	}

	now := time.Now().UTC()
	n := dm.PackageNotification{
		PackageID: pkg.ID,
		TenantID:  *pkg.TenantID,
		Channel:   ch,
		Status:    dm.PackageNotificationStatusSent,
		SentAt:    &now,
	}
	created, err := s.store.CreatePackageNotification(n)
	if err != nil {
		return dm.PackageNotification{}, err
	}

	if pkg.Status == dm.PackageStatusReceived {
		pkg.Status = dm.PackageStatusNotified
		if _, err := s.store.UpdatePackage(pkg); err != nil {
			return created, err
		}
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "package.notification", "package", pkg.ID, adm.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) RegisterGuestVisit(principal administration.Principal, in GuestVisitCreateInput) (dm.GuestVisit, error) {
	if !hasStaffDoorman(principal) {
		return dm.GuestVisit{}, ErrUnauthorized
	}
	if strings.TrimSpace(in.GuestName) == "" {
		return dm.GuestVisit{}, ErrValidation
	}
	if !in.ValidTo.After(in.ValidFrom) {
		return dm.GuestVisit{}, ErrValidation
	}
	if _, err := s.store.GetTenant(in.HostTenantID); err != nil {
		if errors.Is(err, administration.ErrNotFound) {
			return dm.GuestVisit{}, ErrTenantNotFound
		}
		return dm.GuestVisit{}, err
	}

	v := dm.GuestVisit{
		HostTenantID:       in.HostTenantID,
		GuestName:          strings.TrimSpace(in.GuestName),
		IDNotes:            in.IDNotes,
		ValidFrom:          in.ValidFrom.UTC(),
		ValidTo:            in.ValidTo.UTC(),
		Status:             dm.GuestVisitStatusScheduled,
		RegisteredByUserID: &principal.UserID,
	}
	return s.store.CreateGuestVisit(v)
}

func (s *Service) ListGuestVisits(filter GuestVisitListFilter) ([]dm.GuestVisit, int64, error) {
	return s.store.ListGuestVisits(filter)
}

func (s *Service) GetGuestVisit(id uuid.UUID) (dm.GuestVisit, error) {
	return s.store.GetGuestVisit(id)
}

func (s *Service) GuestCheckIn(principal administration.Principal, visitID uuid.UUID) (dm.GuestVisit, error) {
	if !hasStaffDoorman(principal) {
		return dm.GuestVisit{}, ErrUnauthorized
	}
	v, err := s.store.GetGuestVisit(visitID)
	if err != nil {
		return dm.GuestVisit{}, err
	}
	now := time.Now().UTC()
	if guestOutsideWindow(now, v.ValidFrom, v.ValidTo) {
		reason := dm.GuestDenialReasonOutsideVisitWindow
		_, _ = s.store.CreateGuestAccessEvent(dm.GuestAccessEvent{
			GuestVisitID: visitID,
			ActorUserID:  principal.UserID,
			EventType:    dm.GuestAccessEventDenied,
			OccurredAt:   now,
			Reason:       &reason,
		})
		_ = s.store.CreateAudit(newAudit(principal.UserID, "guest.access_denied", "guest_visit", v.ID, adm.AuditOutcomeFailure))
		return dm.GuestVisit{}, ErrOutsideVisitWindow
	}
	switch v.Status {
	case dm.GuestVisitStatusScheduled:
		// proceed
	default:
		return dm.GuestVisit{}, ErrGuestVisitStateTransition
	}
	v.Status = dm.GuestVisitStatusCheckedIn
	updated, err := s.store.UpdateGuestVisit(v)
	if err != nil {
		return dm.GuestVisit{}, err
	}
	_, err = s.store.CreateGuestAccessEvent(dm.GuestAccessEvent{
		GuestVisitID: visitID,
		ActorUserID:  principal.UserID,
		EventType:    dm.GuestAccessEventCheckIn,
		OccurredAt:   now,
	})
	if err != nil {
		return dm.GuestVisit{}, err
	}
	return updated, nil
}

func (s *Service) GuestCheckOut(principal administration.Principal, visitID uuid.UUID) (dm.GuestVisit, error) {
	if !hasStaffDoorman(principal) {
		return dm.GuestVisit{}, ErrUnauthorized
	}
	v, err := s.store.GetGuestVisit(visitID)
	if err != nil {
		return dm.GuestVisit{}, err
	}
	if v.Status != dm.GuestVisitStatusCheckedIn {
		return dm.GuestVisit{}, ErrGuestVisitStateTransition
	}
	now := time.Now().UTC()
	v.Status = dm.GuestVisitStatusCheckedOut
	updated, err := s.store.UpdateGuestVisit(v)
	if err != nil {
		return dm.GuestVisit{}, err
	}
	_, err = s.store.CreateGuestAccessEvent(dm.GuestAccessEvent{
		GuestVisitID: visitID,
		ActorUserID:  principal.UserID,
		EventType:    dm.GuestAccessEventCheckOut,
		OccurredAt:   now,
	})
	if err != nil {
		return dm.GuestVisit{}, err
	}
	return updated, nil
}

func (s *Service) IssueTenantEntryToken(principal administration.Principal, in TenantEntryTokenCreateInput) (dm.TenantEntryToken, error) {
	if !hasStaffDoorman(principal) {
		return dm.TenantEntryToken{}, ErrUnauthorized
	}
	if _, err := s.store.GetTenant(in.TenantID); err != nil {
		if errors.Is(err, administration.ErrNotFound) {
			return dm.TenantEntryToken{}, ErrTenantNotFound
		}
		return dm.TenantEntryToken{}, err
	}
	ref := strings.TrimSpace(in.PublicRef)
	if ref == "" {
		ref = uuid.NewString()
	}
	t := dm.TenantEntryToken{
		TenantID:  in.TenantID,
		PublicRef: ref,
		ExpiresAt: in.ExpiresAt,
	}
	return s.store.CreateTenantEntryToken(t)
}

func (s *Service) ValidateEntryByQR(principal administration.Principal, publicRef string) (QRValidationResult, error) {
	if !hasStaffDoorman(principal) {
		return QRValidationResult{}, ErrUnauthorized
	}
	actor := principal.UserID
	now := time.Now().UTC()
	ref := strings.TrimSpace(publicRef)
	if ref == "" {
		return s.appendDenied(actor, dm.AccessEventSourceQRScan, nil, nil, dm.AccessDenialReasonInvalid)
	}

	token, err := s.store.GetTenantEntryTokenByPublicRef(ref)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return s.appendDenied(actor, dm.AccessEventSourceQRScan, nil, nil, dm.AccessDenialReasonInvalid)
		}
		return QRValidationResult{}, err
	}
	tid := token.TenantID

	if token.RevokedAt != nil {
		return s.appendDenied(actor, dm.AccessEventSourceQRScan, &tid, &token.ID, dm.AccessDenialReasonRevoked)
	}
	if token.ExpiresAt != nil && now.After(*token.ExpiresAt) {
		return s.appendDenied(actor, dm.AccessEventSourceQRScan, &tid, &token.ID, dm.AccessDenialReasonExpired)
	}

	tenantPtr, tenantErr := s.store.GetTenant(tid)
	if tenantErr != nil {
		tiduuid := tid
		return s.appendDenied(actor, dm.AccessEventSourceQRScan, &tiduuid, &token.ID, dm.AccessDenialReasonInvalid)
	}

	ev := dm.AccessEvent{
		TenantID:    &tid,
		ActorUserID: &actor,
		TokenID:     &token.ID,
		Outcome:     dm.AccessEventOutcomeGranted,
		Source:      dm.AccessEventSourceQRScan,
		OccurredAt:  now,
	}
	saved, err := s.store.CreateAccessEvent(ev)
	if err != nil {
		return QRValidationResult{}, err
	}
	return QRValidationResult{Outcome: dm.AccessEventOutcomeGranted, Tenant: &tenantPtr, Event: saved}, nil
}

func (s *Service) ValidateEntryManual(principal administration.Principal, tenantID uuid.UUID) (QRValidationResult, error) {
	if !hasStaffDoorman(principal) {
		return QRValidationResult{}, ErrUnauthorized
	}
	actor := principal.UserID
	tenant, err := s.store.GetTenant(tenantID)
	if err != nil {
		if errors.Is(err, administration.ErrNotFound) {
			return s.appendDenied(actor, dm.AccessEventSourceManual, nil, nil, dm.AccessDenialReasonInvalid)
		}
		return QRValidationResult{}, err
	}
	tid := tenant.ID
	ev := dm.AccessEvent{
		TenantID:    &tid,
		ActorUserID: &actor,
		Outcome:     dm.AccessEventOutcomeGranted,
		Source:      dm.AccessEventSourceManual,
		OccurredAt:  time.Now().UTC(),
	}
	saved, err := s.store.CreateAccessEvent(ev)
	if err != nil {
		return QRValidationResult{}, err
	}
	return QRValidationResult{Outcome: dm.AccessEventOutcomeGranted, Tenant: &tenant, Event: saved}, nil
}

func (s *Service) appendDenied(actor uuid.UUID, source dm.AccessEventSource, tenantID *uuid.UUID, tokenID *uuid.UUID, reason string) (QRValidationResult, error) {
	r := reason
	ev := dm.AccessEvent{
		TenantID:    tenantID,
		ActorUserID: &actor,
		TokenID:     tokenID,
		Outcome:     dm.AccessEventOutcomeDenied,
		Reason:      &r,
		Source:      source,
		OccurredAt:  time.Now().UTC(),
	}
	saved, err := s.store.CreateAccessEvent(ev)
	if err != nil {
		return QRValidationResult{}, err
	}
	_ = s.store.CreateAudit(newAudit(actor, "access.denied", "access_event", saved.ID, adm.AuditOutcomeFailure))
	return QRValidationResult{Outcome: dm.AccessEventOutcomeDenied, Reason: reason, Event: saved}, nil
}

func (s *Service) ListAccessEvents(filter AccessEventListFilter) ([]dm.AccessEvent, int64, error) {
	return s.store.ListAccessEvents(filter)
}

func (s *Service) CheckoutItem(principal administration.Principal, body CheckoutLoanBody) (dm.ItemLoan, error) {
	if !hasStaffDoorman(principal) {
		return dm.ItemLoan{}, ErrUnauthorized
	}
	item, err := s.store.GetInventoryItem(body.InventoryItemID)
	if err != nil {
		return dm.ItemLoan{}, err
	}
	if item.Status == adm.InventoryStatusWithdrawn || item.Status == adm.InventoryStatusDestroyed {
		return dm.ItemLoan{}, ErrLoanItemUnavailable
	}
	count, err := s.store.CountOpenLoansForInventory(body.InventoryItemID)
	if err != nil {
		return dm.ItemLoan{}, err
	}
	if count > 0 {
		return dm.ItemLoan{}, ErrLoanItemUnavailable
	}
	tid := body.TenantID
	if _, err := s.store.GetTenant(tid); err != nil {
		if errors.Is(err, administration.ErrNotFound) {
			return dm.ItemLoan{}, ErrTenantNotFound
		}
		return dm.ItemLoan{}, err
	}

	loan := dm.ItemLoan{
		TenantID:           tid,
		InventoryItemID:    body.InventoryItemID,
		CheckedOutByUserID: principal.UserID,
		CheckedOutAt:       time.Now().UTC(),
		ExpectedReturnAt:   body.ExpectedReturnAt,
		Notes:              body.Notes,
	}
	created, err := s.store.CreateItemLoan(loan)
	if err != nil {
		return dm.ItemLoan{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "loan.checkout", "item_loan", created.ID, adm.AuditOutcomeSuccess))
	return created, nil
}

func (s *Service) ReturnItem(principal administration.Principal, loanID uuid.UUID) (dm.ItemLoan, error) {
	if !hasStaffDoorman(principal) {
		return dm.ItemLoan{}, ErrUnauthorized
	}
	loan, err := s.store.GetItemLoan(loanID)
	if err != nil {
		return dm.ItemLoan{}, err
	}
	if loan.ReturnedAt != nil {
		return dm.ItemLoan{}, ErrLoanAlreadyReturned
	}
	now := time.Now().UTC()
	loan.ReturnedAt = &now
	tmp := principal.UserID
	loan.ReturnedByUserID = &tmp
	updated, err := s.store.UpdateItemLoan(loan)
	if err != nil {
		return dm.ItemLoan{}, err
	}
	_ = s.store.CreateAudit(newAudit(principal.UserID, "loan.return", "item_loan", updated.ID, adm.AuditOutcomeSuccess))
	return updated, nil
}

func (s *Service) ListItemLoans(filter ItemLoanListFilter) ([]dm.ItemLoan, int64, error) {
	return s.store.ListItemLoans(filter)
}

func (s *Service) ListTenantEntryTokens(filter TokenListFilter) ([]dm.TenantEntryToken, int64, error) {
	return s.store.ListTenantEntryTokens(filter)
}

func (s *Service) ListActiveTenants() ([]adm.Tenant, error) {
	return s.store.ListActiveTenants()
}

func (s *Service) ListLendableInventory() ([]adm.InventoryItem, error) {
	return s.store.ListLendableInventory()
}

func hasStaffDoorman(p administration.Principal) bool {
	for _, role := range p.Roles {
		switch role {
		case adm.RoleDoorman, adm.RoleAdministrator, adm.RoleOfficeWorker, adm.RoleDirector:
			return true
		}
	}
	return false
}

func guestOutsideWindow(now, from, to time.Time) bool {
	return now.Before(from) || now.After(to)
}

func canTransitionPackage(from, to dm.PackageStatus) bool {
	if from == to {
		return true
	}
	edges := map[dm.PackageStatus][]dm.PackageStatus{
		dm.PackageStatusReceived: {dm.PackageStatusNotified, dm.PackageStatusPickedUp},
		dm.PackageStatusNotified: {dm.PackageStatusPickedUp},
	}
	for _, next := range edges[from] {
		if next == to {
			return true
		}
	}
	return false
}

func newAudit(actorID uuid.UUID, action, targetType string, targetID uuid.UUID, outcome adm.AuditOutcome) adm.AuditEvent {
	return adm.AuditEvent{
		ActorUserID: &actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Outcome:     outcome,
		OccurredAt:  time.Now().UTC(),
	}
}
