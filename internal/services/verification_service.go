package services

import (
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"strconv"
	"time"
)

func SendVerificationCode(username string, email string) *errors.ServerBetaError {

	// Delete old codes
	db := initializers.DB
	db.Where("username = ?", username).Delete(&models.VerificationCode{})

	// Create new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return errors.New("Creating verification code failed", http.StatusInternalServerError)
	}

	codeObject := models.VerificationCode{
		Username:       username,
		Code:           strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save and code
	if err := db.Create(&codeObject).Error; err != nil {
		return errors.New("Creating verification code failed", http.StatusInternalServerError)
	}

	err = SendMail(email, "Verification Code", "Your verification code is:\n\n\t"+codeObject.Code+"\n\nVerify your account now!")
	if err != nil {
		return errors.New("Error sending verification code", http.StatusInternalServerError)
	}

	return nil
}

func VerifyUser(username string, code string) *errors.ServerBetaError {

	// Get code
	db := initializers.DB
	var codeObject models.VerificationCode
	if err := db.Where("username = ? and code = ?", username, code).First(&codeObject).Error; err != nil {
		return errors.New("Verification code not found", http.StatusNotFound)
	}

	// Check if code is expired
	if codeObject.ExpirationTime.Before(time.Now()) {

		// Resend code
		SendVerificationCode(username, codeObject.Username)

		return errors.New("Verification code expired", http.StatusUnauthorized)
	}

	// Activate user
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("User not found", http.StatusNotFound)
	}
	user.Activated = true
	if err := db.Save(&user).Error; err != nil {
		return errors.New("Failed to activate user", http.StatusInternalServerError)
	}

	// Send welcome email
	if err := SendMail(user.Email, "Welcome to Server Beta", "Welcome to Server Beta!\n\nYour account os successfully verified. Now you can use our network!"); err != nil {
		return errors.New("Failed to send welcome email", http.StatusInternalServerError)
	}

	// Delete code
	db.Delete(&codeObject)

	return nil

}

func ResendVerificationCode(username string) *errors.ServerBetaError {

	// Get user
	db := initializers.DB
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("User not found", http.StatusNotFound)
	}

	// Send verification code
	if err := SendVerificationCode(user.Username, user.Email); err != nil {
		return errors.New("Failed to send verification code", http.StatusInternalServerError)
	}

	return nil
}
