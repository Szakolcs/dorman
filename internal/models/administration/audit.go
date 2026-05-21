package models

import (
	crosscutting "dorm-man/internal/models/cross-cutting"

	"github.com/google/uuid"
)

type Audit struct {
	crosscutting.BaseModel
	TableName string    `gorm:"not null"`
	Operation string    `gorm:"not null"`
	OldData   string    `gorm:"type:jsonb"`
	NewData   string    `gorm:"type:jsonb"`
	ChangedAt string    `gorm:"not null"`
	ChangedBy uuid.UUID `gorm:"type: uuid;"`

	User crosscutting.User `gorm:"foreignKey:ChangedBy;references:ID"`
}
