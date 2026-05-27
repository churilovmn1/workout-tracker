package main

import (
	"log"
	"os"

	"github.com/churilovmn1/workout-tracker/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	_ = cfg
	log.Printf("server starting on port %s", os.Getenv("PORT"))
}
