package initializers

import (
	"fmt"
	"github.com/marcbudd/server-beta/internal/models"
	"gorm.io/driver/sqlite"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // Global variable for database

// ConnectToDb can be called after program start to connect to database
func ConnectToDb() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSL_MODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	DB, err = gorm.Open(postgres.Open(dsn))

	if err != nil {
		panic("Failed to connect to db")
	}

	fmt.Println("Connection to database successful...")

}

// CloseDbConnection can be called when program execution is stopped, to close database connection
func CloseDbConnection() {
	if DB != nil {
		db, err := DB.DB()
		if err != nil {
			panic("Failed to close db connection")
		}

		err = db.Close()
		if err != nil {
			panic("Failed to close db connection")
		}
	}

	fmt.Println("Connection to database closed...")
}

// CreateMockDB creates sets global DB variable to in-memory db for testing
func CreateMockDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect in-memory database")
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		panic("failed to migrate database")
	}

	DB = db
}
