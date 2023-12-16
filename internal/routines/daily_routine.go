package routines

import (
	"fmt"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"time"
)

// StartDailyRoutines can be called when starting the server to ensure all the routines run daily at 3 AM
// called with: `go StartDailyRoutines()`
func StartDailyRoutines() {
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

		// Deleting routine
		DailyUserDeletionRoutine()
	}

}

// DailyUserDeletionRoutine will be called daily to delete users that did not verify their email address
func DailyUserDeletionRoutine() {
	fmt.Println("Start daily user deletion routing...")

	db := initializers.DB

	var activationTokens []models.ActivationToken
	db.Preload("User").
		Where("created_at < ? and activated = ?", time.Now().Add(-7*24*time.Hour), false).
		Find(&activationTokens)

	for _, token := range activationTokens {
		db.Delete(&token)
		db.Delete(&token.User)
	}

}
