package models

// All returns all model pointers in migration order.
func All() []any {
	return []any{
		&User{},
		&Role{},
		&UserRole{},
		&Building{},
		&Flat{},
		&Room{},
		&Tenant{},
		&RoomAssignment{},
		&InventoryItem{},
		&MaintenanceTicket{},
		&TicketStatusChange{},
		&OperationalJob{},
		&Activity{},
		&Event{},
		&AuditEvent{},
	}
}
