package routines_test

import (
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/routines"
	"testing"
	"time"
)

// TestDeleteUnactivatedUsersSuccess tests the DeleteUnactivatedUsers function to delete users that did not verify their email address
func TestDeleteUnactivatedUsersSuccess(t *testing.T) {

	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)

	unactivatedUsers := []models.User{
		{
			Username:  "test",
			Activated: false,
			CreatedAt: time.Now(), // User created today --> should not be deleted
		},
		{
			Username:  "test2",
			Activated: false,
			CreatedAt: time.Now().Add(time.Hour * -7 * 25), // User created 7 days ago --> should be deleted
		},
	}

	// Mock expectations
	mockUserRepo.On("GetUnactivatedUsers").Return(unactivatedUsers, nil)
	mockUserRepo.On("DeleteUserByUsername", "test2").Return(nil)

	// Act
	routines.DeleteUnactivatedUsers(mockUserRepo)

	// Assert
	mockUserRepo.AssertExpectations(t)
}
