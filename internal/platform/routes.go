package platform

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	service := NewService(store)
	handler := &Handler{service: service}

	api := e.Group("/home")
	api.POST("/login", handler.login)
	api.GET("/about", handler.about)

	e.GET("/about", handler.about)
}
