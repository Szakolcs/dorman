package maintenance

import (
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	staff := e.Group("/maintenance")
	staff.Use(auth, middleware.RequireRole("maintainer", "admin", "dev"))
	staff.GET("", h.dashboardPage)
	staff.GET("/", h.dashboardPage)
	staff.GET("/tickets", h.ticketsPage)
	staff.GET("/tickets/:id", h.ticketDetailPage)
	staff.PUT("/tickets/:id", h.updateTicket)
	staff.DELETE("/tickets/:id", h.deleteTicket)
}
