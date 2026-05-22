package forum

import (
	"dorm-man/internal/administration"

	"gorm.io/gorm"
)

type Store interface {
}

type GormStore struct {
	db      *gorm.DB
	adminDB administration.Store
}

func NewStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db, adminDB: administration.NewStore(db)}
}
