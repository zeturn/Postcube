package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/WorkPlace/Postcube/backend/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "postcube.db"
	}

	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("failed to create db directory: %v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Question{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	DB = db
	log.Printf("database ready at %s", dbPath)
}
