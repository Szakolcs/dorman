package doorman

import (
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	doorman := e.Group("/doorman")
	doorman.Use(auth, middleware.RequireRole("doorman", "admin", "dev"))
	doorman.GET("", h.dashboardPage)
	doorman.GET("/", h.dashboardPage)
	doorman.GET("/list/tenants", h.listTenantAccess)
	doorman.GET("/access/tenant", h.createTenantAccess)
	doorman.GET("/guests", h.listGuest)
	doorman.POST("/register/guest", h.registerGuest)
	doorman.DELETE("/leave/guest/:guestID", h.deleteGuest)
}
