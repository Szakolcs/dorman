package platform

import (
	"errors"
	"strings"
	"time"

	models "dorm-man/internal/models/administration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	GetUserByEmail(email string) (models.User, error)
	UpdateLastLogin(userID uuid.UUID, at time.Time) error
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := s.db.
		Preload("UserRoles", "revoked_at IS NULL").
		Preload("UserRoles.Role").
		Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrInvalidCredentials
	}
	return user, err
}

func (s *GormStore) UpdateLastLogin(userID uuid.UUID, at time.Time) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}
