package models

import (
	"dorm-man/internal/models/chat"
	"time"

	"github.com/google/uuid"
)

// Membership links a Tenant to a Room. Chats are tenant-only, so the
// participant pointer is always *administration.Tenant. LastReadAt powers
// per-tenant unread counts.
type Membership struct {
	BaseModel
	RoomID     uuid.UUID           `gorm:"type:uuid;not null;index;uniqueIndex:idx_room_tenant_membership"`
	TenantID   uuid.UUID           `gorm:"type:uuid;not null;index;uniqueIndex:idx_room_tenant_membership"`
	Role       chat.MembershipRole `gorm:"type:varchar(20);not null;default:'member';index"`
	JoinedAt   time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastReadAt *time.Time          `gorm:"index"`

	Room   *Room   `gorm:"foreignKey:RoomID;references:ID"`
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}
