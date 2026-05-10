package models

// All returns doorman-owned model pointers in migration order (FK dependencies first).
func All() []any {
	return []any{
		&TenantEntryToken{},
		&Package{},
		&PackageNotification{},
		&GuestVisit{},
		&GuestAccessEvent{},
		&AccessEvent{},
		&ItemLoan{},
	}
}
