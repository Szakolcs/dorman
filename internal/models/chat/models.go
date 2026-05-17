package models

// All returns chat-owned model pointers in migration order (FK dependencies first).
func All() []any {
	return []any{
		&ChatRoom{},
		&ChatTenantProfile{},
		&ChatMessage{},
		&ChatRoomMember{},
		&ChatMembershipSyncLog{},
	}
}
