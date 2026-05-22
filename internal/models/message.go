package models

import (
	"github.com/google/uuid"
)

// Message is a single chat message in a Room. Only tenants can post, so
// SenderTenantID points at administration.Tenant. The (RoomID, CreatedAt)
// composite index is the natural one for paginating a chat backlog.
type Message struct {
	BaseModel
	RoomID         uuid.UUID `gorm:"type:uuid;not null;index:idx_room_created;index"`
	SenderTenantID uuid.UUID `gorm:"type:uuid;not null;index"`
	Body           string    `gorm:"type:text;not null"`

	ChatRoom *ChatRoom `gorm:"foreignKey:RoomID;references:ID"`
	Sender   *Tenant   `gorm:"foreignKey:SenderTenantID;references:ID"`
}
