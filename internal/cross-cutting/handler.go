package cross_cutting

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	loginComponent "dorm-man/web/templates/crosscutting"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

	cookie := new(http.Cookie)
	cookie.Name = "jwt"
	cookie.Value = res.Token
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode

	c.SetCookie(cookie)
	switch res.RedirectTo {
	case "dev":
		c.Response().Header().Set("HX-Redirect", "/dev")
	case "administrator":
		c.Response().Header().Set("HX-Redirect", "/admin/dashboard")

	case "doorman":
		c.Response().Header().Set("HX-Redirect", "/doorman")
	case "maintainer":
		c.Response().Header().Set("HX-Redirect", "/maintenance")
	case "tenant":
		c.Response().Header().Set("HX-Redirect", "/tenant")
	default:
		return h.writeError(c, ErrNoModuleRoute)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) loginPage(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return loginComponent.LoginPage(c.QueryParam("error")).Render(c.Request().Context(), c.Response().Writer)
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
