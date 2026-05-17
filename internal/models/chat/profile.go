package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

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
