package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomAssignment struct {
	BaseModel
	TenantID        uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_active_tenant_assignment,where:ended_at IS NULL"`
	RoomID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	EffectiveAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	EndedAt         *time.Time `gorm:"index"`
	CreatedByUserID uuid.UUID  `gorm:"type:uuid;not null;index"`

	Tenant        Tenant `gorm:"foreignKey:TenantID;references:ID"`
	Room          Room   `gorm:"foreignKey:RoomID;references:ID"`
	CreatedByUser User   `gorm:"foreignKey:CreatedByUserID;references:ID"`
}
