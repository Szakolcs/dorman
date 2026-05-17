package doorman

import (
	adm "dorm-man/internal/models/administration"
	"dorm-man/internal/pagination"
	dm "dorm-man/internal/models/doorman"
)

type PackagesPageData struct {
	Packages   []dm.Package
	Tenants    []adm.Tenant
	Pagination pagination.Meta
}

type GuestsPageData struct {
	Visits     []dm.GuestVisit
	Tenants    []adm.Tenant
	Pagination pagination.Meta
}

type AccessPageData struct {
	Tokens           []dm.TenantEntryToken
	TokensPagination pagination.Meta
	TokensPreserve   map[string]string
	Events           []dm.AccessEvent
	EventsPagination pagination.Meta
	EventsPreserve   map[string]string
	Tenants          []adm.Tenant
}

type LendingPageData struct {
	Loans      []dm.ItemLoan
	Tenants    []adm.Tenant
	Inventory  []adm.InventoryItem
	Pagination pagination.Meta
}
