package chat

import (
	"dorm-man/internal/administration"

	"gorm.io/gorm"
)

type Store interface {
}

type GormStore struct {
	db      *gorm.DB
	housing administration.Store
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, housing: administration.NewStore(db)}
}
