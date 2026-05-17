package chat

import (
	"errors"

	adm "dorm-man/internal/models/administration"
	cm "dorm-man/internal/models/chat"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized     = errors.New("authorization denied")
	ErrNotFound         = errors.New("not found")
	ErrValidation       = errors.New("validation error")
	ErrNotAMember       = errors.New("not_a_member")
	ErrNotGroupOwner    = errors.New("not_group_owner")
	ErrFlatMembership   = errors.New("flat_membership_managed")
	ErrTenantNotFound   = errors.New("tenant_not_found")
	ErrRoomKindMismatch = errors.New("room_kind_mismatch")
)

// TenantPrincipal is the chat actor (tenant session or dev stub via headers).
type TenantPrincipal struct {
	TenantID uuid.UUID
	UserID   *uuid.UUID
}

type SendMessageInput struct {
	Body            string     `json:"body"`
	ClientMessageID *uuid.UUID `json:"client_message_id"`
}

type CreateGroupInput struct {
	Title           string      `json:"title"`
	AvatarStorageKey string     `json:"avatar_storage_key"`
	MemberTenantIDs []uuid.UUID `json:"member_tenant_ids"`
}

type GroupMemberInput struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type OpenDirectInput struct {
	OtherTenantID uuid.UUID `json:"other_tenant_id"`
}

type UpdateProfileInput struct {
	Nickname         *string `json:"nickname"`
	Bio              *string `json:"bio"`
	AvatarStorageKey *string `json:"avatar_storage_key"`
}

type MarkReadInput struct {
	MessageID *uuid.UUID `json:"message_id"`
}

type MessageListFilter struct {
	BeforeMessageID *uuid.UUID
	Limit             int
}

type ConversationSummary struct {
	Room                     cm.ChatRoom `json:"room"`
	UnreadCount              int64       `json:"unread_count"`
	DisplayTitle             string      `json:"display_title"`
	DisplayAvatarStorageKey  string      `json:"display_avatar_storage_key,omitempty"`
	OtherTenantID            *uuid.UUID  `json:"other_tenant_id,omitempty"`
}

type RoomDetail struct {
	Room    cm.ChatRoom        `json:"room"`
	Members []MemberView       `json:"members"`
	Self    MemberView         `json:"self"`
}

type MemberView struct {
	Member      cm.ChatRoomMember `json:"member"`
	DisplayName string            `json:"display_name"`
	Tenant      adm.Tenant        `json:"tenant"`
}

type TenantProfileView struct {
	Profile     *cm.ChatTenantProfile `json:"profile,omitempty"`
	Tenant      adm.Tenant            `json:"tenant"`
	DisplayName string                `json:"display_name"`
}
