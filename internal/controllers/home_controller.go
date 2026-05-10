package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HomeController struct{}

func NewHomeController() *HomeController {
	return &HomeController{}
}

func (h *HomeController) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HomeController) Index(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h *HomeController) About(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}
