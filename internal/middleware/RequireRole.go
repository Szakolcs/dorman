package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw, ok := c.Get("user").(*jwt.Token)
			if !ok || raw == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			claims, ok := raw.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			role, ok := claims["role"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "invalid role")
			}

			for _, allowed := range roles {
				if role == allowed {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "access denied")
		}
	}
}
