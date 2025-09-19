package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
	var db *gorm.DB
	var err error

	maxRetries := 10
	retryInterval := 3 * time.Second

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("Database connected [Posgres]")
			return db
		}

		log.Printf("[attempt %d/%d] failed to connect to database: %v", i, maxRetries, err)
		time.Sleep(retryInterval)
	}

	log.Fatal("[error] failed to initialize database after multiple attempts:", err)
	return nil
}
