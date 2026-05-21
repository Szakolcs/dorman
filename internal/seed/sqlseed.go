package seed

import (
	"embed"
	"fmt"
	"os"

	"gorm.io/gorm"
)

//go:embed data/data.sql
var seedDataFS embed.FS

// DefaultSQLPath is the bundled fabricate export used by cmd/seed.
const DefaultSQLPath = "data/data.sql"

// LoadSQL reads seed SQL from path, or the embedded bundle when path is empty.
func LoadSQL(path string) ([]byte, error) {
	if path == "" || path == DefaultSQLPath {
		b, err := seedDataFS.ReadFile("data/data.sql")
		if err != nil {
			return nil, fmt.Errorf("read embedded seed sql: %w", err)
		}
		if len(b) == 0 {
			return nil, fmt.Errorf("embedded seed SQL is empty")
		}
		return b, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seed sql %q: %w", path, err)
	}
	return b, nil
}

// ExecSQL runs a PostgreSQL script (multiple statements). The fabricate export sets
// session_replication_role = replica so FK order is relaxed during load.
func ExecSQL(db *gorm.DB, sql []byte) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if _, err := sqlDB.Exec(string(sql)); err != nil {
		return fmt.Errorf("execute seed sql: %w", err)
	}
	return nil
}

// RunSQLSeed loads and executes the fabricate SQL export, then normalizes roles and
// ensures development login accounts exist.
func RunSQLSeed(db *gorm.DB, sqlPath string) error {
	b, err := LoadSQL(sqlPath)
	if err != nil {
		return err
	}
	if err := ExecSQL(db, b); err != nil {
		return err
	}
	if err := NormalizeRoleNames(db); err != nil {
		return err
	}
	return EnsureDevAccounts(db)
}
