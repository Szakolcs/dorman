package chat

// All returns chat-owned model pointers in migration order
// (FK dependencies first).
func All() []any {
	return []any{
		&Room{},
		&Membership{},
		&Message{},
		&Ignore{},
	}
}
