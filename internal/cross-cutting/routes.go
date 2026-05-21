package cross_cutting

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB) {
	store := NewStore(db)
	service := NewService(store)
	handler := NewHandler(service)

	api := e.Group("/home")
	e.GET("/", handler.loginPage)
	api.POST("/login", handler.login)

}
