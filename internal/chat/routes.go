package chat

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	api := e.Group("/api/chat")

	api.GET("/conversations", h.listConversations)
	api.GET("/profile", h.getProfile)
	api.PUT("/profile", h.updateProfile)

	api.POST("/direct", h.openDirect)
	api.POST("/groups", h.createGroup)

	api.GET("/rooms/:id", h.getRoom)
	api.GET("/rooms/:id/messages", h.listMessages)
	api.POST("/rooms/:id/messages", h.sendMessage)
	api.POST("/rooms/:id/read", h.markRead)
	api.POST("/rooms/:id/members", h.addGroupMember)
	api.DELETE("/rooms/:id/members/:tenantId", h.removeGroupMember)
	api.POST("/rooms/:id/leave", h.leaveGroup)
}

// NewServiceFromDB exposes the chat service for cross-module hooks (e.g. assignment sync).
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewStore(db))
}
