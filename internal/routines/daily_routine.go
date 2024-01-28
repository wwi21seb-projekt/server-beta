package routines

import (
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"os"
	"time"
)

// StartDailyRoutines can be called when starting the server to ensure all the routines run daily at 3 AM
// called with: `go StartDailyRoutines()`
func StartDailyRoutines(
	mailService services.MailServiceInterface,
	userService services.UserServiceInterface) {

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
		err := userService.DeleteUnactivatedUsers()
		if err != nil { // Send mail to admin if error occurred
			receiver := os.Getenv("EMAIL_ADDRESS")
			subject := "Error in daily routine: DeleteUnactivatedUsers"
			text := fmt.Sprintf("Error in daily routine: DeleteUnactivatedUsers %s", err.Error())
			_ = mailService.SendMail(receiver, subject, text)
		}
	}
}
