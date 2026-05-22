package forum

// All returns forum-owned model pointers in migration order
// (FK dependencies first).
func All() []any {
	return []any{
		&Publication{},
		&Activity{},
		&Event{},
		&EventAttendance{},
		&EventComment{},
		&ActivityBooking{},
	}
}
