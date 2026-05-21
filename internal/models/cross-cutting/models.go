package cross_cutting

func All() []any {
	return []any{
		&User{},
		&Role{},
		&Permission{},
		&Operation{},
	}
}
