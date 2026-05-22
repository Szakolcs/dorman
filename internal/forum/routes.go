package forum

import (
	"dorm-man/internal/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	forum := e.Group("/api/forum")
	forum.Use(middleware.RequireRole("admin", "dev"))
	forum.GET("", h.feedPage)
	forum.GET("/feed", h.feedListFragment)
	forum.POST("/posts", h.createCommunityPostView)
	forum.GET("/posts/:id", h.postPage)
	forum.POST("/posts/:id/comments", h.createCommentView)
	forum.POST("/posts/:id/reactions", h.togglePostReactionView)
	forum.POST("/posts/:id/attendance", h.setAttendanceView)
	forum.POST("/polls/:id/votes", h.castPollVoteView)
}
