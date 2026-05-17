package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

// ChatMessage is a persisted chat message in a room.
type ChatMessage struct {
	platform.BaseModel
	RoomID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_chat_message_idempotent,priority:1;index"`
	AuthorTenantID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_chat_message_idempotent,priority:2;index"`
	Body            string    `gorm:"type:text;not null"`
	EditedAt        *time.Time
	ClientMessageID *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_chat_message_idempotent,priority:3"`

	Room         ChatRoom   `gorm:"foreignKey:RoomID;references:ID"`
	AuthorTenant adm.Tenant `gorm:"foreignKey:AuthorTenantID;references:ID"`
}
