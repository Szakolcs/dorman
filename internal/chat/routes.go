package chat

import (
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	chat := e.Group("/api/chat")
	chat.Use(auth, middleware.RequireRole("tenant", "dev"))
	chat.GET("", h.chatPage)
	chat.GET("/", h.chatPage)
	chat.GET("/view", h.chatViewPage)
	chat.GET("/view/", h.chatViewPage)
	chat.GET("/view/messages", h.chatMessagesFragment)
	chat.GET("/view/messages/:id", h.chatMessageDetailPage)
	chat.POST("/view/messages/:id/reactions", h.toggleMessageReactionView)
	chat.POST("/view/messages/:id/attendance", h.setMessageAttendanceView)

}

// NewServiceFromDB exposes the chat service for cross-module hooks (e.g. assignment sync).
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewStore(db))
}
