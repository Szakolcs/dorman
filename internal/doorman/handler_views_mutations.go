package doorman

import (
	"net/http"
	"strings"
	"time"

	dm "dorm-man/internal/models/doorman"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func redirectDoormanView(c echo.Context, path string, err error) error {
	if err == nil {
		return c.Redirect(http.StatusSeeOther, path)
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, path+"?error="+err.Error())
}

func (h *Handler) registerPackageView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/packages", err)
	}
	input := PackageRegisterInput{
		RecipientLabel: c.FormValue("recipient_label"),
		Description:    c.FormValue("description"),
	}
	if v := strings.TrimSpace(c.FormValue("tenant_id")); v != "" {
		id, parseErr := uuid.Parse(v)
		if parseErr != nil {
			return redirectDoormanView(c, "/doorman/packages", ErrValidation)
		}
		input.TenantID = &id
	}
	_, err = h.service.RegisterPackage(principal, input)
	return redirectDoormanView(c, "/doorman/packages", err)
}

func (h *Handler) pickupPackageView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/packages", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/packages", ErrValidation)
	}
	_, err = h.service.ConfirmPackagePickup(principal, id)
	return redirectDoormanView(c, "/doorman/packages", err)
}

func (h *Handler) notifyPackageView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/packages", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/packages", ErrValidation)
	}
	_, err = h.service.NotifyPackageTenant(principal, id, PackageNotifyInput{
		Channel: dm.PackageNotificationChannel(c.FormValue("channel")),
	})
	return redirectDoormanView(c, "/doorman/packages", err)
}

func (h *Handler) createGuestVisitView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", err)
	}
	hostID, err := uuid.Parse(c.FormValue("host_tenant_id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", ErrValidation)
	}
	validFrom, err := time.Parse(time.RFC3339, c.FormValue("valid_from"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", ErrValidation)
	}
	validTo, err := time.Parse(time.RFC3339, c.FormValue("valid_to"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", ErrValidation)
	}
	_, err = h.service.RegisterGuestVisit(principal, GuestVisitCreateInput{
		HostTenantID: hostID,
		GuestName:    c.FormValue("guest_name"),
		IDNotes:      c.FormValue("id_notes"),
		ValidFrom:    validFrom,
		ValidTo:      validTo,
	})
	return redirectDoormanView(c, "/doorman/guests", err)
}

func (h *Handler) guestCheckInView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", ErrValidation)
	}
	_, err = h.service.GuestCheckIn(principal, id)
	return redirectDoormanView(c, "/doorman/guests", err)
}

func (h *Handler) guestCheckOutView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/guests", ErrValidation)
	}
	_, err = h.service.GuestCheckOut(principal, id)
	return redirectDoormanView(c, "/doorman/guests", err)
}

func (h *Handler) issueTenantEntryTokenView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/access", err)
	}
	tenantID, err := uuid.Parse(c.FormValue("tenant_id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/access", ErrValidation)
	}
	input := TenantEntryTokenCreateInput{
		TenantID:  tenantID,
		PublicRef: c.FormValue("public_ref"),
	}
	if v := strings.TrimSpace(c.FormValue("expires_at")); v != "" {
		t, parseErr := time.Parse(time.RFC3339, v)
		if parseErr != nil {
			return redirectDoormanView(c, "/doorman/access", ErrValidation)
		}
		input.ExpiresAt = &t
	}
	_, err = h.service.IssueTenantEntryToken(principal, input)
	return redirectDoormanView(c, "/doorman/access", err)
}

func (h *Handler) checkoutLoanView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/lending", err)
	}
	tenantID, err := uuid.Parse(c.FormValue("tenant_id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/lending", ErrValidation)
	}
	itemID, err := uuid.Parse(c.FormValue("inventory_item_id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/lending", ErrValidation)
	}
	input := CheckoutLoanBody{
		TenantID:        tenantID,
		InventoryItemID: itemID,
		Notes:           c.FormValue("notes"),
	}
	if v := strings.TrimSpace(c.FormValue("expected_return_at")); v != "" {
		t, parseErr := time.Parse(time.RFC3339, v)
		if parseErr != nil {
			return redirectDoormanView(c, "/doorman/lending", ErrValidation)
		}
		input.ExpectedReturnAt = &t
	}
	_, err = h.service.CheckoutItem(principal, input)
	return redirectDoormanView(c, "/doorman/lending", err)
}

func (h *Handler) returnLoanView(c echo.Context) error {
	principal, err := h.staffActor(c)
	if err != nil {
		return redirectDoormanView(c, "/doorman/lending", err)
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return redirectDoormanView(c, "/doorman/lending", ErrValidation)
	}
	_, err = h.service.ReturnItem(principal, id)
	return redirectDoormanView(c, "/doorman/lending", err)
}
