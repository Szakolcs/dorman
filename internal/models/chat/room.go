package models

import (
	"time"

	adm "dorm-man/internal/models/administration"
	platform "dorm-man/internal/models/platform"

	"github.com/google/uuid"
)

type ChatRoomKind string

const (
	ChatRoomKindFlat   ChatRoomKind = "flat"
	ChatRoomKindDirect ChatRoomKind = "direct"
	ChatRoomKindGroup  ChatRoomKind = "group"
)

// ChatRoom is a conversation container: flat (system), direct (pair), or group (user-created).
type ChatRoom struct {
	platform.BaseModel
	Kind               ChatRoomKind `gorm:"type:varchar(20);not null;index"`
	Title              string       `gorm:"not null;default:''"`
	AvatarStorageKey   string       `gorm:"type:varchar(512)"`
	FlatID             *uuid.UUID   `gorm:"type:uuid;uniqueIndex"`
	TenantLowID        *uuid.UUID   `gorm:"type:uuid;uniqueIndex:idx_chat_direct_pair,priority:1"`
	TenantHighID       *uuid.UUID   `gorm:"type:uuid;uniqueIndex:idx_chat_direct_pair,priority:2"`
	LastMessageAt      *time.Time   `gorm:"index"`
	LastMessagePreview string       `gorm:"type:varchar(500)"`

	Flat       *adm.Flat        `gorm:"foreignKey:FlatID;references:ID"`
	TenantLow  *adm.Tenant      `gorm:"foreignKey:TenantLowID;references:ID"`
	TenantHigh *adm.Tenant      `gorm:"foreignKey:TenantHighID;references:ID"`
	Members    []ChatRoomMember `gorm:"foreignKey:RoomID"`
	Messages   []ChatMessage    `gorm:"foreignKey:RoomID"`
}
