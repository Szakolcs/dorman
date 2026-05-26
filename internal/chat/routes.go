package chat

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes is deprecated; tenant chat lives under /tenant/chat via tenantportal.
func RegisterRoutes(e *echo.Echo, db *gorm.DB, auth echo.MiddlewareFunc) {}

// NewServiceFromDB exposes the chat service for cross-module hooks (e.g. assignment sync).
func NewServiceFromDB(db *gorm.DB) *Service {
	return NewService(NewStore(db))
}
