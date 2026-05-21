package models

// All returns all model pointers in migration order.
func All() []any {
	return []any{
		&Building{},
		&Flat{},
		&Room{},
		&Tenant{},
		&RoomAssignment{},
		&InventoryItem{},
		&OperationalJob{},
		&Activity{},
		&Event{},
	}
}
