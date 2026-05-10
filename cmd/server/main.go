package main

import (
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Printf("server startup failed: %v", err)
		os.Exit(1)
	}
}
