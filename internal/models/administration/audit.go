package models

import (
	"time"

	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeFailure AuditOutcome = "failure"
)

type AuditEvent struct {
	platform.BaseModel
	ActorUserID *uuid.UUID   `gorm:"type:uuid;index"`
	Action      string       `gorm:"not null;index"`
	TargetType  string       `gorm:"not null;index"`
	TargetID    string       `gorm:"index"`
	Outcome     AuditOutcome `gorm:"type:varchar(20);not null;index"`
	Metadata    string       `gorm:"type:text"`
	OccurredAt  time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index"`

	ActorUser *User `gorm:"foreignKey:ActorUserID;references:ID"`
}
