package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadConfig loads .env file
func LoadConfig() {
	// Try to load .env file from current directory
	err := godotenv.Load(".env")
	if err != nil {
		// If not found, try to load from project root
		err = godotenv.Load()
		if err != nil {
			log.Println(".env file not found, using system environment variables")
		}
	}
}
