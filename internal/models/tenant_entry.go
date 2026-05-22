package models

import (
	"time"

	"github.com/google/uuid"
)

type TenantEntry struct {
	BaseModel
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	Status      AccessStatus `gorm:"type:varchar(20);not null;default:'scheduled';index"`
	TimeOfEntry time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}
