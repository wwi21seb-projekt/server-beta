package services

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
	"strconv"
	"time"
)

type UserServiceInterface interface {
	sendActivationToken(email string, tokenObject *models.ActivationToken) *errors.CustomError
	CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *errors.CustomError, int)
	LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *errors.CustomError, int)
	ActivateUser(username string, token string) (*errors.CustomError, int)
	ResendActivationToken(username string) (*errors.CustomError, int)
}

type UserService struct {
	userRepo            repositories.UserRepositoryInterface
	activationTokenRepo repositories.ActivationTokenRepositoryInterface
	mailService         MailServiceInterface
}

// NewUserService can be used as a constructor to generate a new UserService "object"
func NewUserService(
	userRepo repositories.UserRepositoryInterface,
	activationTokenRepo repositories.ActivationTokenRepositoryInterface,
	maliService MailServiceInterface) *UserService {
	return &UserService{userRepo: userRepo, activationTokenRepo: activationTokenRepo, mailService: maliService}
}

// SendActivationToken deletes old activation tokens, generates a new six-digit code and sends it to user via mail
func (service *UserService) sendActivationToken(email string, tokenObject *models.ActivationToken) *errors.CustomError {
	subject := "Verification Token"
	body := "Your verification code is:\n\n\t" + tokenObject.Token + "\n\nVerify your account now!"
	err := service.mailService.SendMail(email, subject, body)
	if err != nil {
		return errors.EmailNotSent
	}

	return nil
}

// CreateUser can be called from the controller and saves the user to the db and returns response, error and status code
func (service *UserService) CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *errors.CustomError, int) {
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
	tx := service.userRepo.BeginTx()
	if tx.Error != nil {
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	// Pessimistic Locking - Check if email or username is taken
	emailExists, err := service.userRepo.CheckEmailExistsForUpdate(req.Email, tx)
	if err != nil || emailExists {
		tx.Rollback()
		return nil, errors.EmailTaken, http.StatusConflict
	}

	usernameExists, err := service.userRepo.CheckUsernameExistsForUpdate(req.Username, tx)
	if err != nil || usernameExists {
		tx.Rollback()
		return nil, errors.UsernameTaken, http.StatusConflict
	}

	// Hash Password
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
		tx.Rollback()
		return nil, errors.InternalServerError, http.StatusInternalServerError
	}

	codeObject := models.ActivationToken{
		Id:             uuid.New(),
		Username:       req.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save user and code to database
	if err := tx.Create(&user); err != nil {
		tx.Rollback()
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	if err := tx.Create(&codeObject).Error; err != nil {
		tx.Rollback()
		return nil, errors.DatabaseError, http.StatusInternalServerError
	}

	// Send activation code
	if err := service.sendActivationToken(user.Email, &codeObject); err != nil {
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
func (service *UserService) LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *errors.CustomError, int) {

	// Find user by username
	user, err := service.userRepo.FindUserByUsername(req.Username)
	if err != nil {
		return nil, errors.InvalidCredentials, http.StatusUnauthorized
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.InvalidCredentials, http.StatusUnauthorized
	}

	// Check if user is activated
	if !user.Activated {

		// Check if there are valid, non-expired tokens
		verificationTokens, err := service.activationTokenRepo.FindTokenByUsername(user.Username)
		if err != nil {
			return nil, errors.DatabaseError, http.StatusInternalServerError
		}

		validTokenFound := false
		for _, token := range verificationTokens {
			if token.ExpirationTime.After(time.Now()) {
				validTokenFound = true
			}
			break
		}

		// If no valid token is found, send a new activation token
		if !validTokenFound {
			err, _ := service.ResendActivationToken(user.Username)
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
func (service *UserService) ActivateUser(username string, token string) (*errors.CustomError, int) {

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		return errors.UserNotFound, http.StatusNotFound
	}

	// If user is already activated --> send success
	if user.Activated == true {
		return errors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Get token
	activationToken, err := service.activationTokenRepo.FindActivationToken(username, token)
	if err != nil {
		return errors.InvalidToken, http.StatusNotFound
	}

	// Check if activation token is expired
	if activationToken.ExpirationTime.Before(time.Now()) {

		// Resend token
		service.ResendActivationToken(user.Username)
		return errors.ActivationTokenExpired, http.StatusUnauthorized
	}

	// Activate user
	user.Activated = true
	if err := service.userRepo.UpdateUser(&user); err != nil {
		return errors.DatabaseError, http.StatusInternalServerError
	}

	// Send welcome email
	subject := "Welcome to Server Beta"
	body := "Welcome to Server Beta!\n\nYour account was successfully verified. Now you can use our network!"
	if err := service.mailService.SendMail(user.Email, subject, body); err != nil {
		return errors.InternalServerError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.activationTokenRepo.DeleteActivationTokenByUsername(user.Username); err != nil {
		return errors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}

// ResendActivationToken can be sent from controller to resend a six digit code via mail
func (service *UserService) ResendActivationToken(username string) (*errors.CustomError, int) {

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
	customError := service.sendActivationToken(user.Email, &codeObject)
	if customError != nil {
		return customError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}
