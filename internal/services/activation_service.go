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

// SendActivationToken deletes old tokens, generates a new six-digit code and sends it to user via mail
func SendActivationToken(email string, tokenObject *models.ActivationToken) *errors.CustomError {

	err := SendMail(email, "Verification Token", "Your verification code is:\n\n\t"+tokenObject.Token+"\n\nVerify your account now!")
	if err != nil {
		return errors.EmailNotSent
	}

	return nil
}

// ResendActivationToken can be sent from controller to resend a six digit code via mail
func ResendActivationToken(username string) (*errors.CustomError, int) {

	// Delete old codes
	db := initializers.DB
	result := db.Where("username = ?", username).Delete(&models.ActivationToken{})
	if result.Error != nil {
		return errors.DatabaseError, http.StatusInternalServerError
	}

	// Create new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return errors.InternalServerError, http.StatusInternalServerError
	}

	codeObject := models.ActivationToken{
		Id:             uuid.New(),
		Username:       username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	if err := db.Create(&codeObject).Error; err != nil {
		return errors.DatabaseError, http.StatusInternalServerError
	}

	// Get user
	var user models.User
	result = db.Where("username = ?", username).Find(&user)
	if result.Error != nil {
		return errors.DatabaseError, http.StatusInternalServerError
	}
	if result.RowsAffected == 0 {
		return errors.UserNotFound, http.StatusNotFound
	}

	// If user is already activated --> send success
	if user.Activated == true {
		return errors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Else: resend code
	customError := SendActivationToken(user.Email, &codeObject)
	if customError != nil {
		return customError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}
