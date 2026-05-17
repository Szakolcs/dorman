package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ChatMembershipSyncEventType string

const (
	ChatMembershipSyncEventAssignmentCreated ChatMembershipSyncEventType = "room_assignment.created"
	ChatMembershipSyncEventAssignmentEnded   ChatMembershipSyncEventType = "room_assignment.ended"
	ChatMembershipSyncEventFlatChanged       ChatMembershipSyncEventType = "tenant.flat_changed"
)

// ChatMembershipSyncLog records assignment-driven sync runs for support replay.
type ChatMembershipSyncLog struct {
	platform.BaseModel
	TenantID  uuid.UUID                   `gorm:"type:uuid;not null;index"`
	FlatID    uuid.UUID                   `gorm:"type:uuid;not null;index"`
	EventType ChatMembershipSyncEventType `gorm:"type:varchar(80);not null;index"`
	AppliedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Payload   []byte    `gorm:"type:jsonb"`

	Tenant adm.Tenant `gorm:"foreignKey:TenantID;references:ID"`
	Flat   adm.Flat   `gorm:"foreignKey:FlatID;references:ID"`
}
