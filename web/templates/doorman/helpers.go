package doorman

import (
	"time"

	dm "dorm-man/internal/models/doorman"
)

func packageTenantLabel(p dm.Package) string {
	if p.Tenant != nil && p.Tenant.Name != "" {
		return p.Tenant.Name
	}
	if p.TenantID != nil {
		return p.TenantID.String()
	}
	return p.RecipientLabel
}

func guestHostName(v dm.GuestVisit) string {
	if v.HostTenant != nil && v.HostTenant.Name != "" {
		return v.HostTenant.Name
	}
	return v.HostTenantID.String()
}

func optionalTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("2006-01-02 15:04")
}

func entryTokenTenant(t dm.TenantEntryToken) string {
	if t.Tenant != nil && t.Tenant.Name != "" {
		return t.Tenant.Name
	}
	return t.TenantID.String()
}

func accessEventTenant(e dm.AccessEvent) string {
	if e.Tenant != nil && e.Tenant.Name != "" {
		return e.Tenant.Name
	}
	if e.TenantID != nil {
		return e.TenantID.String()
	}
	return "—"
}

func loanItemName(loan dm.ItemLoan) string {
	if loan.InventoryItem != nil && loan.InventoryItem.Name != "" {
		return loan.InventoryItem.Name
	}
	return loan.InventoryItemID.String()
}

func loanTenantName(loan dm.ItemLoan) string {
	if loan.Tenant != nil && loan.Tenant.Name != "" {
		return loan.Tenant.Name
	}
	return loan.TenantID.String()
}

func loanStatus(loan dm.ItemLoan) string {
	if loan.ReturnedAt != nil {
		return "returned"
	}
	return "open"
}
