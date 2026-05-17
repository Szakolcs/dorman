package main

import (
	"flag"
	"log"
	"os"

	"dorm-man/internal/config"
	"dorm-man/internal/seed"
	models "dorm-man/internal/models/administration"
	chatModels "dorm-man/internal/models/chat"
	doormanModels "dorm-man/internal/models/doorman"
	forumModels "dorm-man/internal/models/forum"

	"gorm.io/gorm"
)

func main() {
	reset := flag.Bool("reset", false, "truncate all tables before seeding")
	seedNum := flag.Uint64("seed", 42, "RNG seed for reproducible data")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable"
	}

	db, err := config.OpenDB(dbURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if *reset {
		log.Println("truncating existing data...")
		if err := seed.TruncateAll(db); err != nil {
			log.Fatalf("truncate: %v", err)
		}
	}

	log.Println("seeding database (see docs/database/seed.md)...")
	if err := seed.New(db, *seedNum).Run(); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("seed complete")
}

func migrate(db *gorm.DB) error {
	toMigrate := append([]any(nil), models.All()...)
	toMigrate = append(toMigrate, forumModels.All()...)
	toMigrate = append(toMigrate, chatModels.All()...)
	toMigrate = append(toMigrate, doormanModels.All()...)
	return db.AutoMigrate(toMigrate...)
}
