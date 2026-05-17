package pagination

import "gorm.io/gorm"

// Scope returns a GORM scope applying limit and offset.
func Scope(p Params) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p.PageSize <= 0 {
			return db
		}
		return db.Limit(p.Limit()).Offset(p.Offset())
	}
}
