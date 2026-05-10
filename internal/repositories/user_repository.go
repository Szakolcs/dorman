package repositories

import (
	"dorm-man/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}
