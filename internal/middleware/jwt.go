package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// Authenticate parses and validates the JWT stored in the "jwt" cookie.
// Valid tokens are stored on the context under the key "user" (echo-jwt convention).
func Authenticate(sessionSecret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(sessionSecret),
		TokenLookup: "cookie:jwt",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return jwt.MapClaims{}
		},
	})
}
