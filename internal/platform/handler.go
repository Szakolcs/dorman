package platform

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func (h *Handler) login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return h.writeError(c, ErrValidation)
	}

	res, err := h.service.Login(req)
	if err != nil {
		return h.writeError(c, err)
	}

	if c.QueryParam("redirect") == "true" {
		return c.Redirect(http.StatusSeeOther, res.RedirectTo)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": res})
}

func (h *Handler) about(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"data": h.service.About()})
}

func (h *Handler) writeError(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	category := "internal_error"

	switch {
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
		category = "validation_error"
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrInactiveUser):
		status = http.StatusUnauthorized
		category = "authentication_failed"
	case errors.Is(err, ErrNoModuleRoute):
		status = http.StatusForbidden
		category = "authorization_denied"
	}

	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}
