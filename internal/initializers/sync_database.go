package initializers

import (
	"fmt"
	"github.com/marcbudd/server-beta/internal/models"
)

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
