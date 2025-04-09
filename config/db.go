package config

import (
	"fmt"
	"log"
	"os"

	"learningfilesharing/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	_ = godotenv.Load() // Optional: allow fallback to env vars directly

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connected ✅")
	DB = db

	// Auto-migrate all necessary models
	err = db.AutoMigrate(&models.User{}, &models.File{})
	if err != nil {
		log.Fatal("Auto migration failed:", err)
	}
}
