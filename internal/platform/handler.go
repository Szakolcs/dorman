package platform

import (
	"errors"
	"net/http"
	"net/url"

	platformviews "dorm-man/web/templates/platform"

	"github.com/a-h/templ"
	"github.com/google/uuid"
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
		if c.QueryParam("redirect") == "true" {
			q := url.Values{}
			q.Set("error", err.Error())
			return c.Redirect(http.StatusSeeOther, "/?"+q.Encode())
		}
		return h.writeError(c, err)
	}

	if c.QueryParam("redirect") == "true" {
		userID, parseErr := uuid.Parse(res.UserID)
		if parseErr != nil {
			return h.writeError(c, ErrValidation)
		}
		SetSession(c, userID)
		return c.Redirect(http.StatusSeeOther, res.RedirectTo)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": res})
}

func (h *Handler) aboutAPI(c echo.Context) error {
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

func renderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) loginPage(c echo.Context) error {
	return renderComponent(c, platformviews.LoginPage(c.QueryParam("error")))
}

func (h *Handler) aboutPage(c echo.Context) error {
	info := h.service.About()
	return renderComponent(c, platformviews.AboutPage(platformviews.AboutPageData{
		Title:             info.Title,
		History:           info.History,
		StudentLife:       info.StudentLife,
		UsefulInformation: info.UsefulInformation,
	}))
}
