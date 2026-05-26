package models

// All returns all model pointers in migration order
// (FK dependencies first). Activity and Event live in the forum package.
func All() []any {
	return []any{
		&Building{},
		&Flat{},
		&Room{},
		&SharedArea{},
		&Tenant{},
		&RoomAssignment{},
		&InventoryItem{},
		&OperationalJob{},
		&Audit{},
		&ChatRoom{},
		&Membership{},
		&Message{},
		&Ignore{},
		&User{},
		&Role{},
		&Permission{},
		&Operation{},
		&TenantEntry{},
		&GuestEntry{},
		&Publication{},
		&Activity{},
		&Event{},
		&EventAttendance{},
		&EventComment{},
		&ActivityBooking{},
		&Ticket{},
		&StatusChange{},
	}
}
