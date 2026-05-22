package tenantportal

import (
	"dorm-man/internal/administration"
	"dorm-man/internal/chat"
	"dorm-man/internal/maintenance"
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {
	adminStore := administration.NewStore(db)
	adminSvc := administration.NewService(adminStore)
	maintSvc := maintenance.NewService(maintenance.NewStore(db))
	chatSvc := chat.NewService(chat.NewStore(db))
	h := NewHandler(adminSvc, maintSvc, chatSvc)

	tenant := e.Group("/tenant")
	tenant.Use(auth, middleware.RequireRole("tenant", "dev"))
	tenant.GET("", h.publicationsPage)
	tenant.GET("/", h.publicationsPage)
	tenant.GET("/publications/:id", h.publicationDetailPage)
	tenant.POST("/publications/activities/:id/book", h.bookActivity)
	tenant.POST("/publications/activities/:id/cancel", h.cancelActivityBooking)
	tenant.POST("/publications/events/:id/intent", h.setEventIntent)
	tenant.GET("/tickets", h.ticketsPage)
	tenant.POST("/tickets", h.createTicket)
	tenant.GET("/tickets/:id", h.ticketDetailPage)
	tenant.DELETE("/tickets/:id", h.deleteTicket)
	tenant.GET("/chat", h.chatPage)
	tenant.GET("/chat/sidebar", h.chatSidebarFragment)
	tenant.GET("/chat/panel", h.chatPanelFragment)
	tenant.GET("/chat/messages", h.chatMessagesFragment)
	tenant.POST("/chat/messages", h.sendChatMessage)
	tenant.POST("/chat/direct", h.openDirectChat)
	tenant.POST("/chat/groups", h.createGroupChat)
}
