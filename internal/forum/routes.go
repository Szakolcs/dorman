package forum

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	api := e.Group("/api/forum")

	api.GET("/posts", h.listFeed)
	api.POST("/posts", h.createCommunityPost)
	api.GET("/posts/:id", h.getPost)
	api.PATCH("/posts/:id", h.updateCommunityPost)
	api.POST("/posts/:id/archive", h.archiveCommunityPost)

	api.GET("/posts/:id/comments", h.listComments)
	api.POST("/posts/:id/comments", h.createComment)
	api.PATCH("/posts/:id/comments/:commentId", h.updateComment)
	api.DELETE("/posts/:id/comments/:commentId", h.deleteComment)

	api.POST("/posts/:id/reactions", h.togglePostReaction)
	api.PUT("/posts/:id/attendance", h.setAttendance)

	api.GET("/posts/:id/updates", h.listOrganizerUpdates)
	api.POST("/posts/:id/updates", h.addOrganizerUpdate)

	api.POST("/polls", h.createPoll)
	api.POST("/polls/:id/votes", h.castPollVote)
	api.GET("/polls/:id/results", h.getPollResults)

	api.POST("/reactions", h.toggleReaction)
	api.POST("/moderation", h.moderate)
}
