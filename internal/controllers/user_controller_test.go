package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/controllers"
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/marcbudd/server-beta/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCreateUserSuccess tests if CreateUser returns 201-Created when user is created successfully
func TestCreateUserSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)
	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "test@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	// Mock expectations
	mockTx := new(gorm.DB)
	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", email, mockTx).Return(false, nil)       // Email does not exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(false, nil) // Username does not exist
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)            // Send mail successfully
	mockValidator.On("ValidateEmailExistance", email).Return(true)                             // Email exists

	type ArgumentCaptor struct {
		User  *models.User
		Token *models.ActivationToken
	}
	var captor ArgumentCaptor
	mockUserRepository.On("CreateUserTx", mock.AnythingOfType("*models.User"), mockTx).
		Run(func(args mock.Arguments) {
			captor.User = args.Get(0).(*models.User) // Save argument to captor
		}).Return(nil) // Create user successfully
	mockActivationTokenRepository.On("CreateActivationTokenTx", mock.AnythingOfType("*models.ActivationToken"), mockTx).
		Run(func(args mock.Arguments) {
			captor.Token = args.Get(0).(*models.ActivationToken) // Save argument to captor
		}).Return(nil) // Create activation token successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusCreated, w.Code) // Expect HTTP 201 Created status
	var responseUser models.UserResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseUser)
	assert.NoError(t, err)
	assert.Equal(t, username, responseUser.Username)
	assert.Equal(t, nickname, responseUser.Nickname)
	assert.Equal(t, email, responseUser.Email)

	// Assert user saved to database
	assert.Equal(t, username, captor.User.Username) // Expect username to be saved to database
	assert.Equal(t, nickname, captor.User.Nickname) // Expect nickname to be saved to database
	assert.Equal(t, email, captor.User.Email)       // Expect email to be saved to database
	passwordCheck := utils.CheckPassword(password, captor.User.PasswordHash)
	assert.True(t, passwordCheck)          // Expect password to be hashed and saved to database
	assert.False(t, captor.User.Activated) // Expect user to be not activated

	// Assert token saved to database
	assert.Equal(t, username, captor.Token.Username) // Expect username to be saved to database
	assert.Equal(t, 6, len(captor.Token.Token))      // Expect token to be 6 digits long

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

// TestCreateUserInvalidBody tests if CreateUser returns 400-Bad Request when request body/username/... is invalid
func TestCreateUserBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`, // invalid body
		`{"username": "", "nickname": "", "password": "Password123!", "email": "email_test@testdomain.de"}`,   // no username
		`{"username": "testUser", "nickname": "", "password": "passwd", "email": "email_test@testdomain.de"}`, // password does not meet specifications
		`{"username": "testUser", "nickname": "", "password": "passwd123!", "email": "testDomain.de"}`,        // invalid email syntax
	}

	for _, body := range invalidBodies {
		// Setup mocks
		mockUserRepository := new(repositories.MockUserRepository)
		mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
		mockMailService := new(services.MockMailService)
		mockValidator := new(utils.MockValidator)
		mockValidator.On("ValidateEmailExistance", mock.AnythingOfType("string")).Return(true)

		userService := services.NewUserService(
			mockUserRepository,
			mockActivationTokenRepository,
			mockMailService,
			mockValidator,
		)

		userController := controllers.NewUserController(userService)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/users", userController.CreateUser)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse errors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := errors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestCreateUserUsernameExists tests if CreateUser returns 409-Conflict when username already exists
func TestCreateUserUsernameExists(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	// Mock expectations
	mockTx := new(gorm.DB)
	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("RollbackTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", email, mockTx).Return(false, nil)      // Email does not exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(true, nil) // Username does exist
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)           // Send mail successfully
	mockValidator.On("ValidateEmailExistance", email).Return(true)                            // Email exists

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusConflict, w.Code) // Expect HTTP 409 Conflict status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.UsernameTaken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestCreateEmailExists tests if CreateUser returns 409-Conflict when email already exists
func TestCreateEmailExists(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	// Mock expectations
	mockTx := new(gorm.DB)
	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("RollbackTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", email, mockTx).Return(true, nil)        // Email does exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(false, nil) // Username does not exist
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)            // Send mail successfully
	mockValidator.On("ValidateEmailExistance", email).Return(true)                             // Email exists

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusConflict, w.Code) // Expect HTTP 409 Conflict status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.EmailTaken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestCreateUserEmailUnreachable tests if CreateUser returns 422-Unprocessable Entity when email is unreachable
func TestCreateUserEmailUnreachable(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "unreachablemail@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	mockValidator.On("ValidateEmailExistance", email).Return(false) // Email does not exist

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code) // Expect HTTP 422 Unprocessable Entity status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.EmailUnreachable
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

}

// TestCreateUserInternalServerErrorDatabase tests if CreateUser returns 500-Internal Server Error when database error occurs
func TestCreateUserInternalServerErrorDatabase(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	// Mock expectations
	mockTx := new(gorm.DB)
	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("RollbackTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", email, mockTx).Return(false, nil)                                // Email does not exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(false, nil)                          // Username does not exist
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)                                     // Send mail successfully
	mockValidator.On("ValidateEmailExistance", email).Return(true)                                                      // Email exists
	mockUserRepository.On("CreateUserTx", mock.AnythingOfType("*models.User"), mockTx).Return(fmt.Errorf("test error")) // Create user fails

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Expect HTTP 500 Internal Server Error status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.DatabaseError
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestCreateUserInternalServerErrorServer tests if CreateUser returns 500-Internal Server Error when email service fails
func TestCreateUserInternalServerErrorServer(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"

	userRequest := models.UserCreateRequestDTO{
		Username: username,
		Password: password,
		Nickname: nickname,
		Email:    email,
	}

	// Mock expectations
	mockTx := new(gorm.DB)
	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("RollbackTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", email, mockTx).Return(false, nil)                                            // Email does not exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(false, nil)                                      // Username does not exist
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(fmt.Errorf("test error"))                            // Send mail fails
	mockValidator.On("ValidateEmailExistance", email).Return(true)                                                                  // Email exists
	mockUserRepository.On("CreateUserTx", mock.AnythingOfType("*models.User"), mockTx).Return(nil)                                  // Create user successfully
	mockActivationTokenRepository.On("CreateActivationTokenTx", mock.AnythingOfType("*models.ActivationToken"), mockTx).Return(nil) // Create activation token successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users", userController.CreateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Expect HTTP 500 Internal Server Error status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.EmailNotSent
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestLoginSuccess tests if Login returns 200-OK when user is logged in successfully
func TestLoginSuccess(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		Email:        email,
		Activated:    true,
		CreatedAt:    time.Now(),
	}

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: password,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(user, nil) // Find user successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/login", userController.Login)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status
	var responseDto models.UserLoginResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(t, err)
	assert.NotEmpty(t, responseDto.Token)
	assert.NotEmpty(t, responseDto.RefreshToken)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
}

// TestLoginInvalidCredentialsUserNotFound tests if Login returns 401-Unauthorized when user cannot be found
func TestLoginInvalidCredentialsUserNotFound(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	password := "Password123!"

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: password,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(models.User{}, nil) // Do not find user

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/login", userController.Login)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.InvalidCredentials
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
}

// TestLoginInvalidCredentialsPasswordIncorrect tests if Login returns 401-Unauthorized when password is incorrect
func TestLoginInvalidCredentialsPasswordIncorrect(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	incorrectPassword := "wrongPassword"

	user := models.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		Email:        email,
		Activated:    true,
		CreatedAt:    time.Now(),
	}

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: incorrectPassword,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(user, nil) // Find user successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/login", userController.Login)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.InvalidCredentials
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
}

// TestLoginUserNotActivated tests if Login returns 403-Forbidden when user is not activated
func TestLoginUserNotActivated(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		Email:        email,
		Activated:    false,
		CreatedAt:    time.Now(),
	}

	sixDigitToken := "123456"
	activationToken := models.ActivationToken{
		Id:             uuid.New(),
		Username:       username,
		User:           user,
		Token:          sixDigitToken,
		ExpirationTime: time.Now().Add(time.Minute * 15),
	}
	activationTokenList := []models.ActivationToken{activationToken}

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: password,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(user, nil) // Find user successfully
	mockActivationTokenRepository.On("FindTokenByUsername", username).Return(activationTokenList, nil)

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/login", userController.Login)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect HTTP 403 Forbidden status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.UserNotActivated
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
}

// TestLoginUserNotActivated tests if Login returns 403-Forbidden when user is not activated and send new token if all are expired
func TestLoginUserNotActivatedExpiredToken(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		mockValidator,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	nickname := "Test User"
	email := "somemail@domain.com"
	password := "Password123!"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		Email:        email,
		Activated:    false,
		CreatedAt:    time.Now(),
	}

	sixDigitToken := "123456"
	activationToken := models.ActivationToken{
		Id:             uuid.New(),
		Username:       username,
		User:           user,
		Token:          sixDigitToken,
		ExpirationTime: time.Now().Add(time.Minute * -15), // Expired token
	}
	activationTokenList := []models.ActivationToken{activationToken}

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: password,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(user, nil)                                               // Find user successfully
	mockActivationTokenRepository.On("FindTokenByUsername", username).Return(activationTokenList, nil)                    // Find expired token
	mockActivationTokenRepository.On("DeleteActivationTokenByUsername", username).Return(nil)                             // Delete expired token successfully
	mockActivationTokenRepository.On("CreateActivationToken", mock.AnythingOfType("*models.ActivationToken")).Return(nil) // Create new token successfully
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)                                       // Send mail successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/login", userController.Login)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect HTTP 403 Forbidden status
	var errorResponse errors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)

	expectedCustomError := errors.UserNotActivated
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}
