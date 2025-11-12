package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadConfig loads .env file
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system environment variables")
	}
}

