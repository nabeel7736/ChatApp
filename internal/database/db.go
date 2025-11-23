package database

import (
	"chatapp/internal/config"
	"chatapp/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPass, cfg.DBSSL,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	return db
}

func Migrate(db *gorm.DB) error {
	// auto migrate models
	return db.AutoMigrate(
		&models.User{},
		&models.Message{},
		&models.Room{},
	)
}
