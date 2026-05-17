package platform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const sessionCookieName = "dm_session"

var sessionSecret []byte

// ConfigureSession sets the HMAC secret used to sign session cookies.
func ConfigureSession(secret string) {
	if strings.TrimSpace(secret) == "" {
		secret = "dev-session-secret-change-in-production"
	}
	sessionSecret = []byte(secret)
}

// SetSession stores the authenticated user id in a signed HTTP-only cookie.
func SetSession(c echo.Context, userID uuid.UUID) {
	payload := fmt.Sprintf("%s|%d", userID.String(), time.Now().Add(7*24*time.Hour).Unix())
	sig := sign(payload)
	value := base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + sig))
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
}

// ClearSession removes the session cookie.
func ClearSession(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// ActorUserID resolves the acting user from header, form, query, then session cookie.
func ActorUserID(c echo.Context) (uuid.UUID, bool) {
	for _, raw := range []string{
		c.Request().Header.Get("X-Actor-User-ID"),
		c.FormValue("actor_user_id"),
		c.QueryParam("actor_user_id"),
	} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err == nil {
			return id, true
		}
	}
	return sessionUserID(c)
}

// ActorUserIDString is the string form for templates and dev actor bars.
func ActorUserIDString(c echo.Context) string {
	if id, ok := ActorUserID(c); ok {
		return id.String()
	}
	return ""
}

func sessionUserID(c echo.Context) (uuid.UUID, bool) {
	if len(sessionSecret) == 0 {
		return uuid.UUID{}, false
	}
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return uuid.UUID{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return uuid.UUID{}, false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return uuid.UUID{}, false
	}
	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(sign(payload)), []byte(parts[2])) {
		return uuid.UUID{}, false
	}
	userParts := strings.Split(parts[0], "|")
	if len(userParts) != 2 {
		return uuid.UUID{}, false
	}
	exp, err := strconv.ParseInt(userParts[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(userParts[0])
	if err != nil {
		return uuid.UUID{}, false
	}
	return id, true
}

func sign(payload string) string {
	mac := hmac.New(sha256.New, sessionSecret)
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
