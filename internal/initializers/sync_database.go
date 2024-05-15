package initializers

import (
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

// SyncDatabase synchronizes the database tables with the model definitions and creates extensions if necessary
func SyncDatabase() {
	// Create extensions
	extensions := []string{"pg_trgm", "fuzzystrmatch"} // needed for levenshtein distance search
	for _, ext := range extensions {
		err := DB.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", ext)).Error
		if err != nil {
			panic(fmt.Sprintf("Failed to create extension %s: %v", ext, err))
		}
	}

	// Migrate models
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Location{},
		&models.ActivationToken{},
		&models.Post{},
		&models.Hashtag{},
		&models.Subscription{},
		&models.Notification{},
	}

	for _, model := range modelsToMigrate {
		if err := DB.AutoMigrate(model); err != nil {
			panic(fmt.Sprintf("Failed to auto-migrate %T: %v", model, err))
		}
	}

	//DB.Exec("ALTER TABLE activation_tokens ADD FOREIGN KEY (username) REFERENCES users(username)")
	//DB.Exec("ALTER TABLE posts ADD FOREIGN KEY (username) REFERENCES users(username)")
	//DB.Exec("ALTER TABLE subscriptions ADD FOREIGN KEY (follower) REFERENCES users(username)")
	//DB.Exec("ALTER TABLE subscriptions ADD FOREIGN KEY (following) REFERENCES users(username)")
	//DB.Exec("ALTER TABLE notifications ADD FOREIGN KEY (forUsername) REFERENCES users(username)")
	//DB.Exec("ALTER TABLE notifications ADD FOREIGN KEY (fromUsername) REFERENCES users(username)")

	fmt.Println("Synchronizing database successful...")
}
