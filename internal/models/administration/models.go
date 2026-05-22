package administration

// All returns all administration-owned model pointers in migration order
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
	}
}
