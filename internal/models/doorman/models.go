package doorman

// All returns doorman-owned model pointers in migration order (FK dependencies first).
func All() []any {
	return []any{
		&TenantEntry{},
		&GuestEntry{},
	}
}
