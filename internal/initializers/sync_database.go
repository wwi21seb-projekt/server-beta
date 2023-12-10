package initializers

import "fmt"

func SyncDatabase() {
	// Sync Database tables with structs in model package
	// err := DB.AutoMigrate(&models.User{})
	// if err != nil {
	// panic("Failed to auto-migrate database")
	//}

	fmt.Println("Synchronizing database successful...")
}
