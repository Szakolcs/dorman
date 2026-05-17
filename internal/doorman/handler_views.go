package doorman

import (
	"dorm-man/internal/pagination"
	doormanviews "dorm-man/web/templates/doorman"

	"github.com/labstack/echo/v4"
)

func (h *Handler) packagesPage(c echo.Context) error {
	params := pageParams(c)
	packages, total, err := h.service.ListPackages(PackageListFilter{Params: params})
	if err != nil {
		return err
	}
	tenants, err := h.service.ListActiveTenants()
	if err != nil {
		return err
	}
	return renderComponent(c, doormanviews.PackagesPage(doormanviews.PackagesPageData{
		Packages:   packages,
		Tenants:    tenants,
		Pagination: pagination.NewMeta(params, total),
	}))
}

func (h *Handler) guestsPage(c echo.Context) error {
	params := pageParams(c)
	visits, total, err := h.service.ListGuestVisits(GuestVisitListFilter{Params: params})
	if err != nil {
		return err
	}
	tenants, err := h.service.ListActiveTenants()
	if err != nil {
		return err
	}
	return renderComponent(c, doormanviews.GuestsPage(doormanviews.GuestsPageData{
		Visits:     visits,
		Tenants:    tenants,
		Pagination: pagination.NewMeta(params, total),
	}))
}

func (h *Handler) accessPage(c echo.Context) error {
	tokenParams := pageParamsNamed(c, "tokens_page", "tokens_page_size")
	eventParams := pageParamsNamed(c, "events_page", "events_page_size")
	tokens, tokenTotal, err := h.service.ListTenantEntryTokens(TokenListFilter{Params: tokenParams})
	if err != nil {
		return err
	}
	events, eventTotal, err := h.service.ListAccessEvents(AccessEventListFilter{Params: eventParams})
	if err != nil {
		return err
	}
	tenants, err := h.service.ListActiveTenants()
	if err != nil {
		return err
	}
	return renderComponent(c, doormanviews.AccessPage(doormanviews.AccessPageData{
		Tokens:           tokens,
		TokensPagination: pagination.NewMeta(tokenParams, tokenTotal),
		TokensPreserve:   queryPreserve(c, "tokens_page", "tokens_page_size"),
		Events:           events,
		EventsPagination: pagination.NewMeta(eventParams, eventTotal),
		EventsPreserve:   queryPreserve(c, "events_page", "events_page_size"),
		Tenants:          tenants,
	}))
}

func (h *Handler) lendingPage(c echo.Context) error {
	params := pageParams(c)
	loans, total, err := h.service.ListItemLoans(ItemLoanListFilter{Params: params})
	if err != nil {
		return err
	}
	tenants, err := h.service.ListActiveTenants()
	if err != nil {
		return err
	}
	inventory, err := h.service.ListLendableInventory()
	if err != nil {
		return err
	}
	return renderComponent(c, doormanviews.LendingPage(doormanviews.LendingPageData{
		Loans:      loans,
		Tenants:    tenants,
		Inventory:  inventory,
		Pagination: pagination.NewMeta(params, total),
	}))
}
