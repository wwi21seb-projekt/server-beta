package initializers

import (
	"fmt"
	"github.com/marcbudd/server-beta/internal/models"
)

// SyncDatabase synchronizes the database tables with the model definitions
func SyncDatabase() {
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.VerificationToken{},
	}

	for _, model := range modelsToMigrate {
		if err := DB.AutoMigrate(model); err != nil {
			panic(fmt.Sprintf("Failed to auto-migrate %T: %v", model, err))
		}
	}

	fmt.Println("Synchronizing database successful...")
}
