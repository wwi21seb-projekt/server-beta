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
	InitiatePasswordReset(username string) (*models.InitiatePasswordResetResponseDTO, *customerrors.CustomError, int)
	ResetPassword(username string, req *models.ResetPasswordRequestDTO) (*customerrors.CustomError, int)
}

type PasswordResetService struct {
	userRepo          repositories.UserRepositoryInterface
	passwordResetRepo repositories.PasswordResetRepositoryInterface
	mailService       MailServiceInterface
	validator         utils.ValidatorInterface
}

// NewPasswordResetService can be used as a constructor to generate a new PasswordResetService "object"
func NewPasswordResetService(userRepo repositories.UserRepositoryInterface, passwordResetRepo repositories.PasswordResetRepositoryInterface, mailService MailServiceInterface, validator utils.ValidatorInterface) *PasswordResetService {
	return &PasswordResetService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		mailService:       mailService,
		validator:         validator,
	}
}

// InitiatePasswordReset initiates a password reset process by generating a token and sending it via email
func (service *PasswordResetService) InitiatePasswordReset(username string) (*models.InitiatePasswordResetResponseDTO, *customerrors.CustomError, int) {
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

	// Reset old tokens of user from database
	err = service.passwordResetRepo.DeletePasswordResetTokensByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create new token and save it to database
	resetToken := models.PasswordResetToken{
		Id:             uuid.New(),
		Username:       user.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

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
	censoredEmail := utils.CensorEmail(user.Email)
	response := models.InitiatePasswordResetResponseDTO{
		Email: censoredEmail,
	}

	return &response, nil, http.StatusOK
}

// ResetPassword sets a new password for the user if the provided token is valid and the password meets policy requirements
func (service *PasswordResetService) ResetPassword(username string, req *models.ResetPasswordRequestDTO) (*customerrors.CustomError, int) {
	// Validate new password
	if !service.validator.ValidatePassword(req.NewPassword) {
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
	resetToken, err := service.passwordResetRepo.FindPasswordResetToken(username, req.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.PasswordResetTokenInvalid, http.StatusForbidden // send 403 if token cannot be found
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}
	if resetToken.ExpirationTime.Before(time.Now()) {
		return customerrors.PasswordResetTokenInvalid, http.StatusForbidden // send 403 if token is expired
	}

	// Hash new password
	newPasswordHashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Update user's password
	user.PasswordHash = newPasswordHashed
	if err := service.userRepo.UpdateUser(user); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.passwordResetRepo.DeletePasswordResetTokenById(resetToken.Id.String()); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}
