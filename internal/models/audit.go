package models

import (
	"github.com/google/uuid"
)

type Audit struct {
	TableName string     `gorm:"not null"`
	Operation string     `gorm:"not null"`
	OldData   string     `gorm:"type:jsonb"`
	NewData   string     `gorm:"type:jsonb"`
	ChangedAt string     `gorm:"not null"`
	ChangedBy *uuid.UUID `gorm:"type:uuid"`

	User User `gorm:"foreignKey:ChangedBy;references:ID"`
}
