package platform

import (
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("inactive user")
	ErrValidation         = errors.New("validation error")
	ErrNoModuleRoute      = errors.New("no module route for user role")
)

type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type LoginResponse struct {
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	RedirectTo  string    `json:"redirect_to"`
	LastLoginAt time.Time `json:"last_login_at"`
	PrimaryRole string    `json:"primary_role"`
}

type AboutInfo struct {
	Title             string   `json:"title"`
	History           string   `json:"history"`
	StudentLife       string   `json:"student_life"`
	UsefulInformation []string `json:"useful_information"`
}
