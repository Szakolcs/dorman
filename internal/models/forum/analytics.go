package models

import (
	platform "dorm-man/internal/models/cross-cutting"
	"time"

	adm "dorm-man/internal/models/administration"

	"github.com/google/uuid"
)

// ForumPostView: set ViewerUserID for staff, ViewerTenantID for tenant sessions (at least one in application logic).
type ForumPostView struct {
	platform.BaseModel
	PostID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	ViewerUserID   *uuid.UUID `gorm:"type:uuid;index"`
	ViewerTenantID *uuid.UUID `gorm:"type:uuid;index"`
	ViewedAt       time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index"`

	Post         ForumPost   `gorm:"foreignKey:PostID;references:ID"`
	ViewerUser   *adm.User   `gorm:"foreignKey:ViewerUserID;references:ID"`
	ViewerTenant *adm.Tenant `gorm:"foreignKey:ViewerTenantID;references:ID"`
}
