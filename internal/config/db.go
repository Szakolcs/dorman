package config

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDB(databaseURL string) (*gorm.DB, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database url is required")
	}
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}
