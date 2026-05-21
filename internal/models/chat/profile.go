package models

import (
	platform "dorm-man/internal/models/cross-cutting"
	"time"

	adm "dorm-man/internal/models/administration"

	"github.com/google/uuid"
)

// ChatTenantProfile holds per-tenant chat display overrides (nickname, bio, avatar).
type ChatTenantProfile struct {
	platform.BaseModel
	TenantID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Nickname         *string   `gorm:"type:varchar(200)"`
	Bio              string    `gorm:"type:text"`
	AvatarStorageKey string    `gorm:"type:varchar(512)"`
	ProfileUpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Tenant adm.Tenant `gorm:"foreignKey:TenantID;references:ID"`
}
