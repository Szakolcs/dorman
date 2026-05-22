package cross_cutting

import (
	"dorm-man/internal/config"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(e *echo.Echo, db *gorm.DB, cfg config.Config) {
	store := NewStore(db)
	service := NewService(store, cfg.SessionSecret)
	handler := NewHandler(service, cfg.CookieSecure)

	e.GET("/", handler.loginPage)
	home := e.Group("/home")
	home.POST("/login", handler.login)
}
