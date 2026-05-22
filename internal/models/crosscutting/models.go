package crosscutting

func All() []any {
	return []any{
		&User{},
		&Role{},
		&Permission{},
		&Operation{},
	}
}
