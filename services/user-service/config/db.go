package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Email     string `gorm:"unique;not null;type:varchar(255)"`
	Password  string `gorm:"not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

func InitDB() {
	var err error
	err = godotenv.Load("../.env")

	if err != nil {
		log.Println("Warning: Could not load .env file, relying on environment variables.")
		fmt.Println("------------------------------------------------>")
	}

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("DB_URL environment variable is not explicitly set or empty")
		fmt.Println("------------------------------------------------>")
	}

	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		fmt.Println("------------------------------------------------>")
	}

	fmt.Println("Connected to PostgreSQL database completely via GORM!")
	fmt.Println("------------------------------------------------>")

	err = DB.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate users table: %v", err)
		fmt.Println("------------------------------------------------>")
	}

	fmt.Println("Users schema automatically synced.")
	fmt.Println("------------------------------------------------>")
}
