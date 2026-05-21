package models

import (
	platform "dorm-man/internal/models/cross-cutting"
	"time"

	adm "dorm-man/internal/models/administration"

	"github.com/google/uuid"
)

type ChatRoomMemberRole string

const (
	ChatRoomMemberRoleMember ChatRoomMemberRole = "member"
	ChatRoomMemberRoleOwner  ChatRoomMemberRole = "owner"
)

type ChatRoomMemberSource string

const (
	ChatRoomMemberSourceDerived ChatRoomMemberSource = "derived"
	ChatRoomMemberSourceInvited ChatRoomMemberSource = "invited"
	ChatRoomMemberSourceCreated ChatRoomMemberSource = "created"
)

// ChatRoomMember tracks tenant membership in a room, read cursor, and leave state.
type ChatRoomMember struct {
	platform.BaseModel
	RoomID            uuid.UUID            `gorm:"type:uuid;not null;index;uniqueIndex:idx_chat_room_member_active,where:left_at IS NULL"`
	TenantID          uuid.UUID            `gorm:"type:uuid;not null;index;uniqueIndex:idx_chat_room_member_active,where:left_at IS NULL"`
	Role              ChatRoomMemberRole   `gorm:"type:varchar(20);not null;default:'member'"`
	JoinedAt          time.Time            `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LeftAt            *time.Time           `gorm:"index"`
	LastReadMessageID *uuid.UUID           `gorm:"type:uuid;index"`
	Source            ChatRoomMemberSource `gorm:"type:varchar(20);not null;default:'derived'"`

	Room            ChatRoom     `gorm:"foreignKey:RoomID;references:ID"`
	Tenant          adm.Tenant   `gorm:"foreignKey:TenantID;references:ID"`
	LastReadMessage *ChatMessage `gorm:"foreignKey:LastReadMessageID;references:ID"`
}
