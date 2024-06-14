package routines

import (
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/initializers"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"time"
)

// StartDailyRoutines can be called when starting the server to ensure all the routines run daily at 3 AM
// called with: `go StartDailyRoutines()`
func StartDailyRoutines() {
	// Arrange
	userRepo := repositories.NewUserRepository(initializers.DB)
	validator := utils.NewValidator()
	fileSystem := repositories.NewFileSystem()
	imageService := services.NewImageService(fileSystem, validator)

	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, time.Local)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour) // If it is already past 3 AM, add 24 hours to nextRun
		}

		timer := time.NewTimer(time.Until(nextRun))
		<-timer.C

		// Will be called daily at 3 AM to delete users that did not verify their email address
		DeleteUnactivatedUsers(userRepo, imageService)
	}
}

// DeleteUnactivatedUsers deletes all users that have not activated their account within 7 days
func DeleteUnactivatedUsers(userRepo repositories.UserRepositoryInterface, imageService services.ImageServiceInterface) {
	fmt.Println("Delete unactivated users...")

	// Get all unactivated users
	users, err := userRepo.GetUnactivatedUsers()
	if err != nil {
		fmt.Println("Failed loading users from database: ", err)
		return
	}

	// Delete users
	counter := 0
	for _, user := range users {
		if user.CreatedAt.Add(7*24*time.Hour).Before(time.Now()) && user.Activated == false {
			//Delete Profile Image
			if user.ImageURL != "" {
				customErr, _ := imageService.DeleteImage(user.ImageURL)
				if customErr != nil {
					fmt.Println("Error deleting user's picture: ", user.Username, err)
				}
			} // User has not activated account within 7 days
			err := userRepo.DeleteUserByUsername(user.Username)
			if err != nil {
				fmt.Println("Error deleting user: ", user.Username, err)
			} else {
				counter++
			}
		}
	}

	fmt.Println("Deleted ", counter, " unactivated users")
}
