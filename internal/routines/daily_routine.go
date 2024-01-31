package routines

import (
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/initializers"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"time"
)

// StartDailyRoutines can be called when starting the server to ensure all the routines run daily at 3 AM
// called with: `go StartDailyRoutines()`
func StartDailyRoutines() {
	// Arrange
	userRepo := repositories.NewUserRepository(initializers.DB)

	// Execution Time 3:00 AM
	executionTime := time.Date(0, 0, 0, 3, 0, 0, 0, time.Local)

	for {
		now := time.Now()
		nextRun := now.Add(time.Until(executionTime))
		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		timer := time.NewTimer(time.Until(nextRun))
		<-timer.C

		// Will be called daily at 3 AM to delete users that did not verify their email address
		DeleteUnactivatedUsers(userRepo)
	}
}

// DeleteUnactivatedUsers deletes all users that have not activated their account within 7 days
func DeleteUnactivatedUsers(userRepo repositories.UserRepositoryInterface) {
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
		if user.CreatedAt.Add(7*24*time.Hour).Before(time.Now()) && user.Activated == false { // User has not activated account within 7 days
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
