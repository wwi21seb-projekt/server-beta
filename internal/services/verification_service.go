package services

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"strconv"
	"time"
)

func SendVerificationToken(username string, email string) *errors.ServerBetaError {

	// Delete old codes
	db := initializers.DB
	result := db.Where("username = ?", username).Delete(&models.VerificationToken{})
	if result.Error != nil {
		return errors.DATABASE_ERROR
	}

	// Create new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return errors.SERVER_ERROR
	}

	codeObject := models.VerificationToken{
		Id:             uuid.New(),
		Username:       username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save and code
	if err := db.Create(&codeObject).Error; err != nil {
		return errors.SERVER_ERROR
	}

	err = SendMail(email, "Verification Token", "Your verification code is:\n\n\t"+codeObject.Token+"\n\nVerify your account now!")
	if err != nil {
		return errors.EMAIL_NOT_SENT
	}

	return nil
}

func ResendVerificationToken(username string) (*errors.ServerBetaError, int) {

	// Get user
	db := initializers.DB
	var user models.User
	result := db.Where("username = ?", username).Find(&user)
	if result.Error != nil {
		return errors.DATABASE_ERROR, http.StatusInternalServerError
	}
	if result.RowsAffected == 0 {
		return errors.USER_NOT_FOUND, http.StatusNotFound
	}

	// Else: resend code
	err := SendVerificationToken(user.Username, user.Email)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}
