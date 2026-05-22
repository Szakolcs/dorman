package cross_cutting

import (
	"errors"
	"strings"
	"time"

	models "dorm-man/internal/models/crosscutting"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store interface {
	GetUserByNickname(nickname string) (models.User, error)
	UpdateLastLogin(userID uuid.UUID, at time.Time) error
}
type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) GetUserByNickname(nickname string) (models.User, error) {
	var user models.User

	err := s.db.
		Preload("Role").
		Preload("Role.Permission").
		Preload("Role.Permission.Operation").
		Where("LOWER(nickname) = ?", strings.ToLower(strings.TrimSpace(nickname))).
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
