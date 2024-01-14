package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type UserServiceInterface interface {
	sendActivationToken(email string, tokenObject *models.ActivationToken) *customerrors.CustomError
	CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *customerrors.CustomError, int)
	LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *customerrors.CustomError, int)
	ActivateUser(username string, token string) (*customerrors.CustomError, int)
	ResendActivationToken(username string) (*customerrors.CustomError, int)
	SearchUser(username string, limit int, offset int) (*models.UserSearchResponseDTO, *customerrors.CustomError, int)
}

type UserService struct {
	userRepo            repositories.UserRepositoryInterface
	activationTokenRepo repositories.ActivationTokenRepositoryInterface
	mailService         MailServiceInterface
	validator           utils.ValidatorInterface
}

// NewUserService can be used as a constructor to generate a new UserService "object"
func NewUserService(
	userRepo repositories.UserRepositoryInterface,
	activationTokenRepo repositories.ActivationTokenRepositoryInterface,
	maliService MailServiceInterface,
	validator utils.ValidatorInterface) *UserService {
	return &UserService{
		userRepo:            userRepo,
		activationTokenRepo: activationTokenRepo,
		mailService:         maliService,
		validator:           validator}
}

// SendActivationToken deletes old activation tokens, generates a new six-digit code and sends it to user via mail
func (service *UserService) sendActivationToken(email string, tokenObject *models.ActivationToken) *customerrors.CustomError {
	subject := "Verification Token"
	body := "Your verification code is:\n\n\t" + tokenObject.Token + "\n\nVerify your account now!"
	err := service.mailService.SendMail(email, subject, body)
	if err != nil {
		return customerrors.EmailNotSent
	}

	return nil
}

// CreateUser can be called from the controller and saves the user to the db and returns response, error and status code
func (service *UserService) CreateUser(req models.UserCreateRequestDTO) (*models.UserResponseDTO, *customerrors.CustomError, int) {
	// Validate input
	if !service.validator.ValidateUsername(req.Username) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !service.validator.ValidateNickname(req.Nickname) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !service.validator.ValidateEmailSyntax(req.Email) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !service.validator.ValidateEmailExistance(req.Email) {
		return nil, customerrors.EmailUnreachable, http.StatusUnprocessableEntity
	}
	if !service.validator.ValidatePassword(req.Password) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Start a transaction
	tx := service.userRepo.BeginTx()
	if tx.Error != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Pessimistic Locking - Check if email or username is taken
	emailExists, err := service.userRepo.CheckEmailExistsForUpdate(req.Email, tx)
	if err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	if emailExists {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.EmailTaken, http.StatusConflict
	}

	usernameExists, err := service.userRepo.CheckUsernameExistsForUpdate(req.Username, tx)
	if err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	if usernameExists {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.UsernameTaken, http.StatusConflict
	}

	// Hash Password
	passwordHashed, err := utils.HashPassword(req.Password)
	if err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Create user
	user := models.User{
		Username:          req.Username,
		Nickname:          req.Nickname,
		Email:             req.Email,
		ProfilePictureUrl: "", // Profile picture is empty in the beginning
		PasswordHash:      passwordHashed,
		CreatedAt:         time.Now(),
		Activated:         false,
	}

	// Create new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	codeObject := models.ActivationToken{
		Id:             uuid.New(),
		Username:       req.Username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	// Save user and code to database
	if err := service.userRepo.CreateUserTx(&user, tx); err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	if err := service.activationTokenRepo.CreateActivationTokenTx(&codeObject, tx); err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send activation code
	if err := service.sendActivationToken(user.Email, &codeObject); err != nil {
		service.userRepo.RollbackTx(tx)
		return nil, err, http.StatusInternalServerError
	}

	// Commit the transaction
	if err := service.userRepo.CommitTx(tx); err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	response := models.UserResponseDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	return &response, nil, http.StatusCreated
}

// LoginUser can be called from the controller and verifies password and returns response, error and status code
func (service *UserService) LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *customerrors.CustomError, int) {

	// Find user by username
	user, err := service.userRepo.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.InvalidCredentials, http.StatusUnauthorized
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, customerrors.InvalidCredentials, http.StatusUnauthorized
	}

	// Check if user is activated
	if !user.Activated {

		// Check if there are valid, non-expired tokens
		verificationTokens, err := service.activationTokenRepo.FindTokenByUsername(user.Username)
		if err != nil {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
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

		return nil, customerrors.UserNotActivated, http.StatusForbidden
	}

	// Create access token
	accessTokenString, err := utils.GenerateAccessToken(user.Username)
	refreshTokenString, err := utils.GenerateRefreshToken(user.Username)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	var loginResponse = models.UserLoginResponseDTO{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}

	return &loginResponse, nil, http.StatusOK

}

// ActivateUser can be called from the controller to verify email using token and returns response, error and status code
func (service *UserService) ActivateUser(username string, token string) (*customerrors.CustomError, int) {

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.UserNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// If user is already activated --> send already reported
	if user.Activated == true {
		return customerrors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Get token
	activationToken, err := service.activationTokenRepo.FindActivationToken(username, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.InvalidToken, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if activation token is expired
	if activationToken.ExpirationTime.Before(time.Now()) {

		// Resend token
		_, _ = service.ResendActivationToken(user.Username)
		return customerrors.ActivationTokenExpired, http.StatusUnauthorized
	}

	// Activate user
	user.Activated = true
	if err := service.userRepo.UpdateUser(user); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send welcome email
	subject := "Welcome to Server Beta"
	body := "Welcome to Server Beta!\n\nYour account was successfully verified. Now you can use our network!"
	if err := service.mailService.SendMail(user.Email, subject, body); err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.activationTokenRepo.DeleteActivationTokenByUsername(user.Username); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}

// ResendActivationToken can be sent from controller to resend a six digit code via mail
func (service *UserService) ResendActivationToken(username string) (*customerrors.CustomError, int) {

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.UserNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// If user is already activated --> send Already Reported
	if user.Activated == true {
		return customerrors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Else: Delete old codes
	if err := service.activationTokenRepo.DeleteActivationTokenByUsername(username); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create and save new code
	digits, err := utils.GenerateSixDigitCode()
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	codeObject := models.ActivationToken{
		Id:             uuid.New(),
		Username:       username,
		Token:          strconv.FormatInt(digits, 10),
		ExpirationTime: time.Now().Add(2 * time.Hour),
	}

	if err := service.activationTokenRepo.CreateActivationToken(&codeObject); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Resend code
	customError := service.sendActivationToken(user.Email, &codeObject)
	if customError != nil {
		return customError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent

}

// SearchUser can be called from the controller to search for users and returns response, error and status code
func (service *UserService) SearchUser(username string, limit int, offset int) (*models.UserSearchResponseDTO, *customerrors.CustomError, int) {
	// Get users
	users, totalRecordsCount, err := service.userRepo.SearchUser(username, limit, offset)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response
	var records []models.UserSearchRecordDTO
	for _, user := range users {
		record := models.UserSearchRecordDTO{
			Username:          user.Username,
			Nickname:          user.Nickname,
			ProfilePictureUrl: user.ProfilePictureUrl,
		}
		records = append(records, record)
	}
	paginationDto := models.UserSearchPaginationDTO{
		Offset:  offset,
		Limit:   limit,
		Records: totalRecordsCount,
	}
	response := models.UserSearchResponseDTO{
		Records:    records,
		Pagination: &paginationDto,
	}

	return &response, nil, http.StatusOK
}
