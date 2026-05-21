package models

import (
	crosscutting "dorm-man/internal/models/cross-cutting"
	"time"

	"github.com/google/uuid"
)

type AccessStatus string

const (
	CheckedIn  AccessStatus = "checked_in"
	CheckedOut AccessStatus = "checked_out"
)

type TenantEntry struct {
	crosscutting.BaseModel
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	Status      AccessStatus `gorm:"type:varchar(20);not null;default:'scheduled';index"`
	TimeOfEntry time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`

	User *crosscutting.User `gorm:"foreignKey:HostTenantID;references:ID"`
}
