package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type PasswordResetServiceInterface interface {
	PasswordReset(username string) (*customerrors.CustomError, int)
	SetNewPassword(token string, newPassword string) (*customerrors.CustomError, int)
}

type PasswordResetService struct {
	userRepo          repositories.UserRepositoryInterface
	passwordResetRepo repositories.PasswordResetRepositoryInterface
	mailService       MailServiceInterface
	validator         utils.ValidatorInterface
}

// NewPasswordResetService can be used as a constructor to generate a new PasswordResetService "object"
func NewPasswordResetService(
	userRepo repositories.UserRepositoryInterface,
	passwordResetRepo repositories.PasswordResetRepositoryInterface,
	mailService MailServiceInterface,
	validator utils.ValidatorInterface) *PasswordResetService {
	return &PasswordResetService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		mailService:       mailService,
		validator:         validator,
	}
}

// PasswordReset initiates a password reset process by generating a token and sending it via email
func (service *PasswordResetService) PasswordReset(username string) (*customerrors.CustomError, int) {
	// Find user by username
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.UserNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Generate token
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	resetToken := models.PasswordResetToken{
		Id:             uuid.New(),
		Username:       user.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save token to database
	if err := service.passwordResetRepo.CreatePasswordResetToken(&resetToken); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send email with token
	subject := "Password Reset Token"
	body := "Your password reset code is:\n\n\t" + resetToken.Token + "\n\nUse this to reset your password."
	err = service.mailService.SendMail(user.Email, subject, body)
	if err != nil {
		return customerrors.EmailNotSent, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}

// SetNewPassword sets a new password for the user if the provided token is valid
func (service *PasswordResetService) SetNewPassword(token string, newPassword string) (*customerrors.CustomError, int) {
	// Validate new password
	if !service.validator.ValidatePassword(newPassword) {
		return customerrors.BadRequest, http.StatusBadRequest
	}

	// Find token in database
	resetToken, err := service.passwordResetRepo.FindPasswordResetToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.InvalidToken, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if token is expired
	if resetToken.ExpirationTime.Before(time.Now()) {
		return customerrors.ActivationTokenExpired, http.StatusUnauthorized
	}

	// Find user by username
	user, err := service.userRepo.FindUserByUsername(resetToken.Username)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Hash new password
	newPasswordHashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Update user's password
	user.PasswordHash = newPasswordHashed
	if err := service.userRepo.UpdateUser(user); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.passwordResetRepo.DeletePasswordResetToken(token); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
