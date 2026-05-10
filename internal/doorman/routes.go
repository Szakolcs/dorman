package doorman

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	pages := e.Group("/doorman")
	pages.GET("", h.dashboardPage)
	pages.GET("/packages", h.packagesPage)
	pages.GET("/guests", h.guestsPage)
	pages.GET("/access", h.accessPage)
	pages.GET("/lending", h.lendingPage)

	api := e.Group("/api/doorman")

	api.POST("/packages", h.registerPackage)
	api.GET("/packages", h.listPackages)
	api.GET("/packages/:id", h.getPackage)
	api.POST("/packages/:id/transition", h.transitionPackage)
	api.POST("/packages/:id/pickup", h.pickupPackage)
	api.POST("/packages/:id/notify", h.notifyPackage)

	api.POST("/guest-visits", h.createGuestVisit)
	api.GET("/guest-visits", h.listGuestVisits)
	api.GET("/guest-visits/:id", h.getGuestVisit)
	api.POST("/guest-visits/:id/check-in", h.guestCheckIn)
	api.POST("/guest-visits/:id/check-out", h.guestCheckOut)

	api.POST("/tenant-entry-tokens", h.issueTenantEntryToken)

	api.POST("/access/validate-qr", h.validateEntryQR)
	api.POST("/access/validate-manual", h.validateEntryManual)
	api.GET("/access/events", h.listAccessEvents)

	api.POST("/loans", h.checkoutLoan)
	api.GET("/loans", h.listLoans)
	api.POST("/loans/:id/return", h.returnLoan)
}
