package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dbPath string, models ...interface{}) {
	if dbPath == "" {
		dbPath = "data/app.db"
	}
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("cannot create data dir: %v", err)
		}
	}
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := database.AutoMigrate(models...); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}
	DB = database
	log.Printf("database initialized at %s", dbPath)
}

func MustGet() *gorm.DB {
	if DB == nil {
		panic("database not initialized")
	}
	return DB
}

func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get raw db: %w", err)
	}
	return sqlDB.Close()
}
