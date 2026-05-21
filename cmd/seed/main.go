package main

import (
	"flag"
	"log"
	"os"

	"dorm-man/internal/config"
	models "dorm-man/internal/models/administration"
	chatModels "dorm-man/internal/models/chat"
	doormanModels "dorm-man/internal/models/doorman"
	forumModels "dorm-man/internal/models/forum"
	"dorm-man/internal/seed"

	"gorm.io/gorm"
)

func main() {
	legacy := flag.Bool("legacy", false, "use the older Go programmatic seeder instead of SQL")
	sqlPath := flag.String("sql", seed.DefaultSQLPath, "path to fabricate SQL export (default: embedded data/data.sql)")
	reset := flag.Bool("reset", false, "with --legacy only: truncate before Go seed (SQL export truncates itself)")
	seedNum := flag.Uint64("seed", 42, "with --legacy only: RNG seed")
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

	if *legacy {
		runLegacy(db, *reset, *seedNum)
		return
	}

	log.Printf("seeding from SQL (%s)...", *sqlPath)
	if err := seed.RunSQLSeed(db, *sqlPath); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("seed complete (SQL + dev accounts)")
	log.Printf("dev login: %s / %s", seed.DevAdminEmail, seed.DevPassword)
}

func runLegacy(db *gorm.DB, reset bool, seedNum uint64) {
	if reset {
		log.Println("truncating existing data...")
		if err := seed.TruncateAll(db); err != nil {
			log.Fatalf("truncate: %v", err)
		}
	}
	log.Println("seeding database (legacy Go seeder)...")
	if err := seed.New(db, seedNum).Run(); err != nil {
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
