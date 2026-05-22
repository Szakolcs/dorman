package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"dorm-man/internal/config"
	"dorm-man/internal/models"
	"dorm-man/internal/seed"
)

func main() {
	reset := flag.Bool("reset", false, "truncate all application tables before seeding")
	flag.Parse()

	cfg := config.Load()
	db, err := config.OpenDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	toMigrate := models.All()
	if err := db.AutoMigrate(toMigrate...); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	if err := seed.Run(db, *reset); err != nil {
		log.Fatalf("seed: %v", err)
	}

	fmt.Fprintln(os.Stdout, "database seeded successfully")
}
