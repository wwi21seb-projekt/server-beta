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
	"strings"
	"time"
)

type PasswordResetServiceInterface interface {
	PasswordReset(username string) (*models.PasswordResetResponseDTO, *customerrors.CustomError, int)
	SetNewPassword(username string, dto models.SetNewPasswordDTO) (*customerrors.CustomError, int)
}

type PasswordResetService struct {
	userRepo          repositories.UserRepositoryInterface
	passwordResetRepo repositories.PasswordResetRepositoryInterface
	mailService       MailServiceInterface
	validator         utils.ValidatorInterface
}

// NewPasswordResetService can be used as a constructor to generate a new PasswordResetService "object"
func NewPasswordResetService(userRepo repositories.UserRepositoryInterface, passwordResetRepo *repositories.PasswordResetRepository, mailService MailServiceInterface, validator utils.ValidatorInterface) *PasswordResetService {
	return &PasswordResetService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		mailService:       mailService,
		validator:         validator,
	}
}

// PasswordReset initiates a password reset process by generating a token and sending it via email
func (service *PasswordResetService) PasswordReset(username string) (*models.PasswordResetResponseDTO, *customerrors.CustomError, int) {
	// Find user by username
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Generate token
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	resetToken := models.PasswordResetToken{
		Id:             uuid.New(),
		Username:       user.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save token to database
	if err := service.passwordResetRepo.CreatePasswordResetToken(&resetToken); err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send email with token
	subject := "Password Reset Token"
	body := "Your password reset code is:\n\n\t" + resetToken.Token + "\n\nUse this to reset your password."
	err = service.mailService.SendMail(user.Email, subject, body)
	if err != nil {
		return nil, customerrors.EmailNotSent, http.StatusInternalServerError
	}

	// Create response with censored email
	censoredEmail := censorEmail(user.Email)
	response := models.PasswordResetResponseDTO{
		CensoredEmail: censoredEmail,
	}

	return &response, nil, http.StatusOK
}

// censorEmail censors the email address for the response
func censorEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	if len(name) > 3 {
		name = name[:3] + strings.Repeat("*", len(name)-3)
	}
	return name + "@" + parts[1]
}

// SetNewPassword sets a new password for the user if the provided token is valid
func (service *PasswordResetService) SetNewPassword(username string, dto models.SetNewPasswordDTO) (*customerrors.CustomError, int) {
	// Validate new password
	if !service.validator.ValidatePassword(dto.NewPassword) {
		return customerrors.BadRequest, http.StatusBadRequest
	}

	// Find user by username
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.UserNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Find token in database
	resetToken, err := service.passwordResetRepo.FindPasswordResetToken(username, dto.Token)
	if err != nil || resetToken.ExpirationTime.Before(time.Now()) {
		return customerrors.PasswordResetTokenInvalid, http.StatusForbidden
	}

	// Hash new password
	newPasswordHashed, err := utils.HashPassword(dto.NewPassword)
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Update user's password
	user.PasswordHash = newPasswordHashed
	if err := service.userRepo.UpdateUser(user); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.passwordResetRepo.DeletePasswordResetToken(resetToken.Id.String()); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
