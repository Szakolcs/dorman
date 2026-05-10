package router

import (
	"dorm-man/internal/controllers"

	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	homeController := controllers.NewHomeController()

	e.GET("/health", homeController.Health)
	e.GET("/", homeController.Index)
	e.GET("/about", homeController.About)
}
