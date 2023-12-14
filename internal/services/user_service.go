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

// CreateUser can be called from the controller and saves the user to the db and returns response, error and status code
func CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *errors.CustomError, int) {
	// Validate input
	if !utils.ValidateUsername(req.Username) {
		return nil, errors.BadRequest, http.StatusBadRequest
	}
	if !utils.ValidateNickname(req.Nickname) {
		return nil, errors.BadRequest, http.StatusBadRequest
	}
	if !utils.ValidateEmailSyntax(req.Email) {
		return nil, errors.BadRequest, http.StatusBadRequest
	}
	if !utils.ValidateEmailExistance(req.Email) {
		return nil, errors.EmailUnreachable, http.StatusUnprocessableEntity
	}
	if !utils.ValidatePassword(req.Password) {
		return nil, errors.BadRequest, http.StatusBadRequest
	}

	// Start a transaction
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	// Pessimistic Locking - Check if email or username is taken
	var count int64 = 0
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.EmailTaken, http.StatusConflict
	}

	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.UsernameTaken, http.StatusConflict
	}

	// Hash password
	passwordHashed, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, errors.InternalServerError, http.StatusInternalServerError
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: passwordHashed,
		CreatedAt:    time.Now(),
		Activated:    false,
	}

	// Create new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return nil, errors.InternalServerError, http.StatusInternalServerError
	}

	codeObject := models.ActivationToken{
		Id:             uuid.New(),
		Username:       req.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save user and code to database
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	if err := tx.Create(&codeObject).Error; err != nil {
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	// Send activation code
	if err := SendActivationToken(user.Email, &codeObject); err != nil {
		tx.Rollback()
		return nil, err, http.StatusInternalServerError
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	response := models.UserResponseDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	return &response, nil, http.StatusCreated
}

// LoginUser can be called from the controller and verifies password and returns response, error and status code
func LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *errors.CustomError, int) {

	// Find user by username
	var user models.User
	if err := initializers.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.InvalidCredentials, http.StatusUnauthorized
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.InvalidCredentials, http.StatusUnauthorized
	}

	// Check if user is activated
	if !user.Activated {
		var verificationTokens []models.ActivationToken
		result := initializers.DB.Where("username = ?", req.Username).Find(&verificationTokens)
		if result.Error != nil {
			return nil, errors.DatabaseError, http.StatusInternalServerError
		}

		// Check if there are valid, non-expired tokens
		validTokenFound := false
		for _, token := range verificationTokens {
			if token.ExpirationTime.After(time.Now()) {
				validTokenFound = true
			}
			break
		}

		// If no valid token is found, send a new verification
		if !validTokenFound {
			err, _ := ResendActivationToken(user.Username)
			if err != nil {
				return nil, err, http.StatusInternalServerError
			}
		}

		return nil, errors.UserNotActivated, http.StatusForbidden
	}

	// Create access token
	tokenString, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, errors.InternalServerError, http.StatusInternalServerError
	}

	var loginResponse = models.UserLoginResponseDTO{
		Token:        tokenString,
		RefreshToken: "",
	}

	return &loginResponse, nil, http.StatusOK

}

// ActivateUser can be called from the controller to verify email using token and returns response, error and status code
func ActivateUser(username string, token string) (*errors.CustomError, int) {

	// Get user
	db := initializers.DB
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.UserNotFound, http.StatusNotFound
	}

	// If user is already activated --> send success
	if user.Activated == true {
		return errors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Get token
	var activationToken models.ActivationToken
	if err := db.Where("username = ? and token = ?", username, token).First(&activationToken).Error; err != nil {
		return errors.InvalidToken, http.StatusNotFound
	}

	// Check if activation token is expired
	if activationToken.ExpirationTime.Before(time.Now()) {

		// Resend token
		ResendActivationToken(user.Username)
		return errors.ActivationTokenExpired, http.StatusUnauthorized
	}

	// Verify user
	user.Activated = true
	if err := db.Save(&user).Error; err != nil {
		return errors.InternalServerError, http.StatusInternalServerError
	}

	// Send welcome email
	if err := SendMail(user.Email, "Welcome to Server Beta", "Welcome to Server Beta!\n\nYour account was successfully verified. Now you can use our network!"); err != nil {
		return errors.InternalServerError, http.StatusInternalServerError
	}

	// Delete token
	db.Where("username = ?", user.Username).Delete(&models.ActivationToken{})

	return nil, http.StatusNoContent

}
