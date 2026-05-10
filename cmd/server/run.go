package main

import (
	"fmt"

	"dorm-man/internal/app"
)

func run() error {
	server, err := app.New()
	if err != nil {
		return fmt.Errorf("bootstrap app: %w", err)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}
