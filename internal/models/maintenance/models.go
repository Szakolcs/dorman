package maintenance

// All returns maintenance-owned model pointers in migration order.
func All() []any {
	return []any{
		&Ticket{},
		&StatusChange{},
	}
}
