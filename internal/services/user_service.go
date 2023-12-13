package services

import (
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"time"
)

// CreateUser can be called from the controller and saves the user to the db and returns response, error and status code
func CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *errors.ServerBetaError, int) {
	// Validate input
	if !utils.ValidateUsername(req.Username) {
		return nil, errors.INVALID_USERNAME, http.StatusBadRequest
	}
	if !utils.ValidateNickname(req.Nickname) {
		return nil, errors.INVALID_NICKNAME, http.StatusBadRequest
	}
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.INVALID_EMAIL, http.StatusBadRequest
	}
	if !utils.ValidatePassword(req.Password) {
		return nil, errors.INVALID_PASSWORD, http.StatusBadRequest
	}

	// Start a transaction
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
	}

	// Pessimistic Locking - Check if email or username is taken
	var count int64 = 0
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.EMAIL_TAKEN, http.StatusConflict
	}

	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.USERNAME_TAKEN, http.StatusConflict
	}

	// Hash password
	passwordHashed, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, errors.SERVER_ERROR, http.StatusInternalServerError
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: passwordHashed,
		CreatedAt:    time.Now(),
		Verified:     false,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
	}

	// Send verification code
	if err := SendVerificationToken(user.Username, user.Email); err != nil {
		tx.Rollback()
		return nil, err, http.StatusInternalServerError
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
	}

	response := models.UserResponseDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	return &response, nil, http.StatusCreated
}

// LoginUser can be called from the controller and verifies password and returns response, error and status code
func LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *errors.ServerBetaError, int) {

	// Find user by username
	var user models.User
	if err := initializers.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.USER_NOT_FOUND, http.StatusNotFound
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.INCORRECT_PASSWORD, http.StatusUnauthorized
	}

	// Check if user is activated
	if !user.Verified {
		var verificationTokens []models.VerificationToken
		result := initializers.DB.Where("username = ?", req.Username).Find(&verificationTokens)
		if result.Error != nil {
			return nil, errors.DATABASE_ERROR, http.StatusInternalServerError
		}

		// Check if there are valid, non-expired tokens
		validTokenFound := false
		for _, token := range verificationTokens {
			if token.ExpirationTime.After(time.Now()) {
				validTokenFound = true
				break
			}
		}

		// If no valid token is found, send a new verification
		if !validTokenFound {
			err := SendVerificationToken(user.Username, user.Email)
			if err != nil {
				return nil, err, http.StatusInternalServerError
			}
			return nil, errors.USER_NOT_VERIFIED, http.StatusForbidden
		}
	}

	// Create token
	tokenString, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, errors.SERVER_ERROR, http.StatusInternalServerError
	}

	var loginResponse = models.UserLoginResponseDTO{
		Token:        tokenString,
		RefreshToken: "",
	}

	return &loginResponse, nil, http.StatusOK

}

// VerifyUser can be called from the controller to verify user using token and returns response, error and status code
func VerifyUser(username string, token string) (*errors.ServerBetaError, int) {

	// Get user
	db := initializers.DB
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.USER_NOT_FOUND, http.StatusNotFound
	}

	// If user is already verified --> send success
	if user.Verified == true {
		return nil, http.StatusNoContent
	}

	// Get token
	var verificationTokenObject models.VerificationToken
	if err := db.Where("username = ? and token = ?", username, token).First(&verificationTokenObject).Error; err != nil {
		return errors.VERIFICATION_TOKEN_NOT_FOUND, http.StatusNotFound
	}

	// Check if token is expired
	if verificationTokenObject.ExpirationTime.Before(time.Now()) {

		// Resend token
		SendVerificationToken(user.Username, user.Email)
		return errors.VERIFICATION_TOKEN_EXPIRED, http.StatusUnauthorized
	}

	// Verify user
	user.Verified = true
	if err := db.Save(&user).Error; err != nil {
		return errors.SERVER_ERROR, http.StatusInternalServerError
	}

	// Send welcome email
	if err := SendMail(user.Email, "Welcome to Server Beta", "Welcome to Server Beta!\n\nYour account was successfully verified. Now you can use our network!"); err != nil {
		return errors.SERVER_ERROR, http.StatusInternalServerError
	}

	// Delete token
	db.Where("username = ?", user.Username).Delete(&models.VerificationToken{})

	return nil, http.StatusNoContent

}
