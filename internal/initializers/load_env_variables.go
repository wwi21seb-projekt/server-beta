package initializers

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// LoadEnvVariables loads necessary environment variables (DB, Mail Server, etc.) from .env file
func LoadEnvVariables() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env-file")
	}

	fmt.Println("Environment variables loaded...")
}
