package chat

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) writeError(c echo.Context, err error) error {
	status := http.StatusInternalServerError
	category := "internal_error"

	switch {
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
		category = "validation_error"
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusForbidden
		category = "authorization_denied"
	case errors.Is(err, ErrNotAMember):
		status = http.StatusForbidden
		category = "not_a_member"
	case errors.Is(err, ErrNotGroupOwner), errors.Is(err, ErrFlatMembership), errors.Is(err, ErrRoomKindMismatch):
		status = http.StatusForbidden
		category = "authorization_denied"
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		category = "not_found"
	case errors.Is(err, ErrTenantNotFound):
		status = http.StatusNotFound
		category = "tenant_not_found"
	}

	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"category": category,
			"message":  err.Error(),
		},
	})
}
