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

	pages := e.Group("/chat/view")
	pages.GET("", h.workspacePage)
	pages.GET("/conversations", h.conversationsPartial)
	pages.GET("/profile", h.profilePage)
	pages.POST("/profile", h.profileSaveView)
	pages.GET("/groups/new", h.groupNewPage)
	pages.POST("/groups", h.groupCreateView)
	pages.POST("/direct", h.openDirectView)
	pages.GET("/rooms/:id", h.roomPage)
	pages.GET("/rooms/:id/messages", h.messagesPartial)
	pages.POST("/rooms/:id/messages", h.sendMessageView)
	pages.POST("/rooms/:id/read", h.markReadView)
	pages.POST("/rooms/:id/members", h.roomMembersView)
	pages.POST("/rooms/:id/leave", h.leaveGroupView)
}

// NewServiceFromDB exposes the chat service for cross-module hooks (e.g. assignment sync).
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewStore(db))
}
