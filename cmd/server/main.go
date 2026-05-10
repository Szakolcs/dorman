package main

import (
	"log"

	"dorm-man/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatalf("failed to bootstrap app: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
