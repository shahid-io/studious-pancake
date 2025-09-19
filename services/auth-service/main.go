package main

import (
	"log"
	"net/http"

	"github.com/shahid-io/studious-pancake/libs/domain/user"
	"github.com/shahid-io/studious-pancake/pkg/config"
	"github.com/shahid-io/studious-pancake/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database with retry
	db := database.Connect(cfg.DatabaseURL)

	// Auto-migrate User model
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Start HTTP server
	addr := ":" + cfg.AppPort
	log.Printf("Auth-Service running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
