package services

import (
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
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

type UserServiceInterface interface {
	sendActivationToken(email string, tokenObject *models.ActivationToken) *customerrors.CustomError
	CreateUser(req models.UserCreateRequestDTO) (*models.UserCreateResponseDTO, *customerrors.CustomError, int)
	LoginUser(req models.UserLoginRequestDTO) (*models.UserLoginResponseDTO, *customerrors.CustomError, int)
	ActivateUser(username string, token string) (*models.UserLoginResponseDTO, *customerrors.CustomError, int)
	ResendActivationToken(username string) (*customerrors.CustomError, int)
	RefreshToken(req *models.UserRefreshTokenRequestDTO) (*models.UserLoginResponseDTO, *customerrors.CustomError, int)
	SearchUser(username string, limit int, offset int, currentUsername string) (*models.UserSearchResponseDTO, *customerrors.CustomError, int)
	UpdateUserInformation(req *models.UserInformationUpdateRequestDTO, currentUsername string) (*models.UserInformationUpdateResponseDTO, *customerrors.CustomError, int)
	ChangeUserPassword(req *models.ChangePasswordDTO, currentUsername string) (*customerrors.CustomError, int)
	GetUserProfile(username string, currentUser string) (*models.UserProfileResponseDTO, *customerrors.CustomError, int)
}

type UserService struct {
	userRepo            repositories.UserRepositoryInterface
	activationTokenRepo repositories.ActivationTokenRepositoryInterface
	mailService         MailServiceInterface
	validator           utils.ValidatorInterface
	postRepo            repositories.PostRepositoryInterface
	imageRepo           repositories.ImageRepositoryInterface
	subscriptionRepo    repositories.SubscriptionRepositoryInterface
	policy              *bluemonday.Policy
}

// NewUserService can be used as a constructor to generate a new UserService "object"
func NewUserService(
	userRepo repositories.UserRepositoryInterface,
	activationTokenRepo repositories.ActivationTokenRepositoryInterface,
	mailService MailServiceInterface,
	validator utils.ValidatorInterface,
	postRepo repositories.PostRepositoryInterface,
	imageRepo repositories.ImageRepositoryInterface,
	subscriptionRepo repositories.SubscriptionRepositoryInterface) *UserService {
	return &UserService{
		userRepo:            userRepo,
		activationTokenRepo: activationTokenRepo,
		mailService:         mailService,
		validator:           validator,
		postRepo:            postRepo,
		imageRepo:           imageRepo,
		subscriptionRepo:    subscriptionRepo,
		policy:              bluemonday.UGCPolicy(),
	}
}

// sendActivationToken deletes old activation tokens, generates a new six-digit code and sends it to user via mail
func (service *UserService) sendActivationToken(email string, tokenObject *models.ActivationToken) *customerrors.CustomError {
	subject := "Verify your account"
	body := utils.GetActivationEmailBody(tokenObject.Token)
	err := service.mailService.SendMail(email, subject, body)
	if err != nil {
		return customerrors.EmailNotSent
	}

	return nil
}

// CreateUser can be called from the controller and saves the user to the db and returns response, error and status code
func (service *UserService) CreateUser(req models.UserCreateRequestDTO) (*models.UserCreateResponseDTO, *customerrors.CustomError, int) {
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

	// Validate and create image object
	var image *models.Image
	var imageId *uuid.UUID
	if req.ProfilePicture != "" {
		imageBytes, err := base64.StdEncoding.DecodeString(req.ProfilePicture)
		if err != nil {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
		valid, format, width, height := service.validator.ValidateImage(imageBytes)
		if !valid {
			return nil, customerrors.BadRequest, http.StatusBadRequest
		}
		image = &models.Image{
			Id:        uuid.New(),
			Format:    format,
			ImageData: imageBytes,
			Width:     width,
			Height:    height,
			Tag:       time.Now().UTC(),
		}
		imageId = &image.Id
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
		Username:     req.Username,
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: passwordHashed,
		CreatedAt:    time.Now(),
		Activated:    false,
		Status:       "", //status is empty in the beginning
	}

	// Add image to user if image was given
	if imageId != nil {
		user.ImageId = imageId
		user.Image = *image
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

	// Create response
	var imageResponseDto *models.ImageMetadataDTO
	if user.ImageId != nil {
		imageResponseDto = &models.ImageMetadataDTO{
			Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
			Width:  user.Image.Width,
			Height: user.Image.Height,
			Tag:    user.Image.Tag,
		}
	}

	response := models.UserCreateResponseDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Picture:  imageResponseDto,
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
func (service *UserService) ActivateUser(username string, token string) (*models.UserLoginResponseDTO, *customerrors.CustomError, int) {

	// Get user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// If user is already activated --> send already reported
	if user.Activated == true {
		return nil, customerrors.UserAlreadyActivated, http.StatusAlreadyReported
	}

	// Get token
	activationToken, err := service.activationTokenRepo.FindActivationToken(username, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.InvalidToken, http.StatusUnauthorized
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if activation token is expired
	if activationToken.ExpirationTime.Before(time.Now()) {

		// Resend token
		_, _ = service.ResendActivationToken(user.Username)
		return nil, customerrors.ActivationTokenExpired, http.StatusUnauthorized
	}

	// Activate user
	user.Activated = true
	if err := service.userRepo.UpdateUser(user); err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send welcome email
	subject := "Welcome to Server Beta"
	body := utils.GetWelcomeEmailBody(username)
	if err := service.mailService.SendMail(user.Email, subject, body); err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Delete token
	if err := service.activationTokenRepo.DeleteActivationTokenByUsername(user.Username); err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Generate access and refresh token
	accessTokenString, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}
	refreshTokenString, err := utils.GenerateRefreshToken(user.Username)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}
	loginResponse := models.UserLoginResponseDTO{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}

	return &loginResponse, nil, http.StatusOK

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

// RefreshToken can be called from the controller with refresh token to return a new access token
func (service *UserService) RefreshToken(req *models.UserRefreshTokenRequestDTO) (*models.UserLoginResponseDTO, *customerrors.CustomError, int) {

	// Verify refresh token
	username, isRefreshToken, err := utils.VerifyJWTToken(req.RefreshToken)
	if err != nil || !isRefreshToken {
		return nil, customerrors.InvalidToken, http.StatusUnauthorized
	}

	// Generate new access token
	accessTokenString, err := utils.GenerateAccessToken(username)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}
	refreshTokenString, err := utils.GenerateRefreshToken(username)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Create response
	loginResponse := models.UserLoginResponseDTO{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}

	return &loginResponse, nil, http.StatusOK
}

// SearchUser can be called from the controller to search for users and returns response, error and status code
func (service *UserService) SearchUser(username string, limit int, offset int, currentUsername string) (*models.UserSearchResponseDTO, *customerrors.CustomError, int) {
	// Get users
	users, totalRecordsCount, err := service.userRepo.SearchUser(username, limit, offset, currentUsername)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response
	response := models.UserSearchResponseDTO{
		Records: []models.UserDTO{},
		Pagination: &models.OffsetPaginationDTO{
			Offset:  offset,
			Limit:   limit,
			Records: totalRecordsCount,
		},
	}

	for _, user := range users {
		var imageDto *models.ImageMetadataDTO
		if user.ImageId != nil {
			imageDto = &models.ImageMetadataDTO{
				Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
				Width:  user.Image.Width,
				Height: user.Image.Height,
				Tag:    user.Image.Tag,
			}
		}

		record := models.UserDTO{
			Username: user.Username,
			Nickname: user.Nickname,
			Picture:  imageDto,
		}
		response.Records = append(response.Records, record)
	}

	return &response, nil, http.StatusOK
}

// UpdateUserInformation can be called from the controller to update a user's nickname and status
func (service *UserService) UpdateUserInformation(req *models.UserInformationUpdateRequestDTO, currentUsername string) (*models.UserInformationUpdateResponseDTO, *customerrors.CustomError, int) {
	// Sanitize status because it is a free text field
	// Other fields are checked with regex patterns, that don't allow for malicious input
	req.Status = strings.Trim(req.Status, " ") // Trim leading and trailing whitespaces
	req.Status = service.policy.Sanitize(req.Status)

	// Check if the new nickname and status are valid
	if !service.validator.ValidateNickname(req.Nickname) || !service.validator.ValidateStatus(req.Status) {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Find the user by username
	user, err := service.userRepo.FindUserByUsername(currentUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.Unauthorized, http.StatusUnauthorized
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Set the new nickname and status
	// Nickname and status are always given and always changed
	user.Nickname = req.Nickname
	user.Status = req.Status

	oldImageId := user.ImageId

	shouldDeleteImage := false
	if req.Picture != nil { // if picture is nil, the user does not want to update the profile picture
		// if picture is "", the user wants to delete the profile picture
		if *req.Picture == "" {
			// set image attributes to nil
			shouldDeleteImage = true
			user.ImageId = nil
			user.Image = models.Image{}
		} else { // if picture is not "", the user wants to update the profile picture
			// Decode and validate image
			imageBytes, err := base64.StdEncoding.DecodeString(*req.Picture)
			if err != nil {
				return nil, customerrors.BadRequest, http.StatusBadRequest
			}
			valid, format, width, height := service.validator.ValidateImage(imageBytes)
			if !valid {
				return nil, customerrors.BadRequest, http.StatusBadRequest
			}

			// Create image id if no image was not present, otherwise use the old image id
			if oldImageId == nil {
				user.Image.Id = uuid.New()
			} else {
				user.Image.Id = *oldImageId
			}

			// Update image attributes
			user.ImageId = &user.Image.Id
			user.Image.Format = format
			user.Image.Width = width
			user.Image.Height = height
			user.Image.Tag = time.Now() // tag is updated, because the image is updated
			user.Image.ImageData = imageBytes
		}
	}

	// Save update to database (also updates image)
	err = service.userRepo.UpdateUser(user)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	if shouldDeleteImage && oldImageId != nil { // if image attributes were set to nil, delete image from database
		err = service.imageRepo.DeleteImageById(oldImageId.String())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		} // if image is not found, ignore error
	}

	// Create and return the response DTO
	var imageDto *models.ImageMetadataDTO
	if user.ImageId != nil {
		imageDto = &models.ImageMetadataDTO{
			Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
			Width:  user.Image.Width,
			Height: user.Image.Height,
			Tag:    user.Image.Tag,
		}
	}

	responseDTO := models.UserInformationUpdateResponseDTO{
		Nickname: user.Nickname,
		Status:   user.Status,
		Picture:  imageDto,
	}

	return &responseDTO, nil, http.StatusOK
}

// ChangeUserPassword can be called from the controller to update a user's password
func (service *UserService) ChangeUserPassword(req *models.ChangePasswordDTO, currentUsername string) (*customerrors.CustomError, int) {
	// Validate the new password
	if !service.validator.ValidatePassword(req.NewPassword) {
		return customerrors.BadRequest, http.StatusBadRequest
	}

	// Find the user by username
	user, err := service.userRepo.FindUserByUsername(currentUsername)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Verify the old password
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return customerrors.InvalidCredentials, http.StatusForbidden
	}

	// Hash the new password
	newPasswordHashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Update the user's password
	user.PasswordHash = newPasswordHashed
	if err := service.userRepo.UpdateUser(user); err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	return nil, http.StatusNoContent
}

// GetUserProfile returns information about the user
func (service *UserService) GetUserProfile(username string, currentUser string) (*models.UserProfileResponseDTO, *customerrors.CustomError, int) {
	// find user
	user, err := service.userRepo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Get counts
	postCount, err := service.postRepo.GetPostCountByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	followerCount, followingCount, err := service.subscriptionRepo.GetSubscriptionCountByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Set subscription id if current user is following
	var subscriptionId *string
	sub, err := service.subscriptionRepo.GetSubscriptionByUsernames(currentUser, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			subscriptionId = nil // if user is not following, return null
		} else {
			return nil, customerrors.DatabaseError, http.StatusInternalServerError
		}
	} else {
		id := sub.Id.String()
		subscriptionId = &id
	}

	// Create response
	var imageDto *models.ImageMetadataDTO
	if user.ImageId != nil {
		imageDto = &models.ImageMetadataDTO{
			Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
			Width:  user.Image.Width,
			Height: user.Image.Height,
			Tag:    user.Image.Tag,
		}
	}

	userProfile := &models.UserProfileResponseDTO{
		Username:       user.Username,
		Nickname:       user.Nickname,
		Status:         user.Status,
		Picture:        imageDto,
		Follower:       followerCount,
		Following:      followingCount,
		Posts:          postCount,
		SubscriptionId: subscriptionId,
	}

	return userProfile, nil, http.StatusOK
}
