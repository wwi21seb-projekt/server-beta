package services

import (
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
)

func CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *errors.ServerBetaError) {
	// Validate input
	if !utils.ValidateUsername(req.Username) {
		return nil, errors.New("username is not valid", http.StatusBadRequest)
	}
	if !utils.ValidateNickname(req.Nickname) {
		return nil, errors.New("nickname is not valid", http.StatusBadRequest)
	}
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.New("email is not valid", http.StatusBadRequest)
	}
	if !utils.ValidatePassword(req.Password) {
		return nil, errors.New("password is not valid", http.StatusBadRequest)
	}

	// Start a transaction
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return nil, errors.New("database error", http.StatusInternalServerError)
	}

	// Pessimistic Locking - Check if email or username is taken
	var count int64 = 0
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("database error", http.StatusInternalServerError)
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.New("email is already taken", http.StatusConflict)
	}

	if err := tx.Set("gorm:query_option", "FOR UPDATE").Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("database error", http.StatusInternalServerError)
	}
	if count > 0 {
		tx.Rollback()
		return nil, errors.New("username is already taken", http.StatusConflict)
	}

	// Hash password
	passwordHashed, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to hash password", http.StatusInternalServerError)
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: passwordHashed,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, errors.New(err.Error(), http.StatusInternalServerError)
	}

	// Send verification code
	if err := SendVerificationCode(user.Username, user.Email); err != nil {
		tx.Rollback()
		return nil, errors.New("failed to send verification code", http.StatusInternalServerError)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit transaction", http.StatusInternalServerError)
	}

	response := models.UserResponseDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	return &response, nil
}

func LoginUser(req models.UserLoginRequestDTO) (*models.User, *errors.ServerBetaError) {

	// Find user by username
	var user models.User
	if err := initializers.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("username not found", http.StatusNotFound)
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("password is incorrect", http.StatusUnauthorized)
	}

	// Check if user is activated
	if !user.Activated {
		return nil, errors.New("user is not activated", http.StatusUnauthorized)
	}

	return &user, nil

}
