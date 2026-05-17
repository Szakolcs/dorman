package platform

import (
	"crypto/subtle"
	"strings"
	"time"

	models "dorm-man/internal/models/administration"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Login(input LoginRequest) (LoginResponse, error) {
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		return LoginResponse{}, ErrValidation
	}

	user, err := s.store.GetUserByEmail(input.Email)
	if err != nil {
		return LoginResponse{}, err
	}
	if !user.IsActive {
		return LoginResponse{}, ErrInactiveUser
	}

	// Current scaffold keeps password verification simple until dedicated auth integration is added.
	if subtle.ConstantTimeCompare([]byte(user.PasswordHash), []byte(input.Password)) != 1 {
		return LoginResponse{}, ErrInvalidCredentials
	}

	redirectTo, role, err := resolveModuleRedirect(user.UserRoles)
	if err != nil {
		return LoginResponse{}, err
	}

	now := time.Now().UTC()
	if err := s.store.UpdateLastLogin(user.ID, now); err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		UserID:      user.ID.String(),
		Name:        user.Name,
		Email:       user.Email,
		RedirectTo:  redirectTo,
		LastLoginAt: now,
		PrimaryRole: string(role),
	}, nil
}

func (s *Service) About() AboutInfo {
	return AboutInfo{
		Title:       "Dormitory Overview",
		History:     "Our dormitory supports university students with safe, shared housing and a long tradition of community programs.",
		StudentLife: "Residents participate in study groups, sports, social events, and peer support activities across the semester.",
		UsefulInformation: []string{
			"Office hours are published each semester by the administration office.",
			"Maintenance requests can be submitted through the maintenance workflow.",
			"Room assignments and move-in details are announced before each semester.",
			"Official news and events are communicated through the tenant-facing forum module.",
		},
	}
}

func resolveModuleRedirect(userRoles []models.UserRole) (string, models.RoleName, error) {
	for _, ur := range userRoles {
		switch ur.Role.Name {
		case models.RoleAdministrator, models.RoleOfficeWorker, models.RoleDirector:
			return "/administration", ur.Role.Name, nil
		}
	}
	for _, ur := range userRoles {
		switch ur.Role.Name {
		case models.RoleDoorman:
			return "/doorman", ur.Role.Name, nil
		case models.RoleTenant:
			return "/forum/view", ur.Role.Name, nil
		}
	}
	return "", "", ErrNoModuleRoute
}
