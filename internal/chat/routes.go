package chat

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	svc := NewService(store)
	h := NewHandler(svc)

	chat := e.Group("/api/chat")

}

// NewServiceFromDB exposes the chat service for cross-module hooks (e.g. assignment sync).
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewStore(db))
}
