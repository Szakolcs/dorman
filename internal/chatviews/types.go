package chatviews

import (
	adm "dorm-man/internal/models/administration"
	cm "dorm-man/internal/models/chat"

	"github.com/google/uuid"
)

type ConversationSummary struct {
	Room                    cm.ChatRoom `json:"room"`
	UnreadCount             int64       `json:"unread_count"`
	DisplayTitle            string      `json:"display_title"`
	DisplayAvatarStorageKey string      `json:"display_avatar_storage_key,omitempty"`
	OtherTenantID           *uuid.UUID  `json:"other_tenant_id,omitempty"`
}

type RoomDetail struct {
	Room    cm.ChatRoom  `json:"room"`
	Members []MemberView `json:"members"`
	Self    MemberView   `json:"self"`
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
