package models

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func ConnectDatabase() {

	database, err := gorm.Open(sqlite.Open(getEnv("DATABASE_URL", "test.db")), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	err = database.AutoMigrate(&Book{})
	if err != nil {
		return
	}

	DB = database
}
