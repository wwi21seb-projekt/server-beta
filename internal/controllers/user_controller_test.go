package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
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
		nil,
		nil,
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
			nil,
			nil,
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
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UsernameTaken
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.EmailTaken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestCreateEmailExistsRollback test if CreateUser does not block username when email already existed in previous request
func TestCreateEmailExistsRollback(t *testing.T) {
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.EmailTaken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Setup second request
	userRequest.Email = "anothermail@mail.com"
	requestBody, err = json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}

	mockUserRepository.On("BeginTx").Return(mockTx)
	mockUserRepository.On("CommitTx", mockTx).Return(nil)
	mockUserRepository.On("RollbackTx", mockTx).Return(nil)
	mockUserRepository.On("CheckEmailExistsForUpdate", userRequest.Email, mockTx).Return(false, nil) // New email does not exist
	mockUserRepository.On("CheckUsernameExistsForUpdate", username, mockTx).Return(false, nil)       // Username does not exist
	mockMailService.On("SendMail", userRequest.Email, mock.Anything, mock.Anything).Return(nil)      // Send mail successfully
	mockValidator.On("ValidateEmailExistance", userRequest.Email).Return(true)
	mockActivationTokenRepository.On("CreateActivationTokenTx", mock.AnythingOfType("*models.ActivationToken"), mockTx).Return(nil)
	mockUserRepository.On("CreateUserTx", mock.AnythingOfType("*models.User"), mockTx).Return(nil)

	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusCreated, w.Code) // Expect HTTP 201 Created status
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.EmailUnreachable
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.DatabaseError
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
		nil,
		nil,
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.EmailNotSent
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
		nil,
		nil,
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
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil) // Find user successfully

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

	extractedUsername, isRefresh, err := utils.VerifyJWTToken(responseDto.Token)
	assert.NoError(t, err)
	assert.False(t, isRefresh)
	assert.Equal(t, username, extractedUsername)

	extractedUsername, isRefresh, err = utils.VerifyJWTToken(responseDto.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, isRefresh)
	assert.Equal(t, username, extractedUsername)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
}

// TestLoginBadRequest tests if Login returns 400-Bad Request when request body is invalid
func TestLoginBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,    // invalid body
		`{"password": "Password123!"}`, // no username
		`{"username": "testUser"}`,     // no password
	}

	for _, body := range invalidBodies {
		controller := controllers.NewUserController(nil)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/users/login", controller.Login)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
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
		nil,
		nil,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	password := "Password123!"

	userRequest := models.UserLoginRequestDTO{
		Username: username,
		Password: password,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&models.User{}, nil) // Do not find user

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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.InvalidCredentials
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
		nil,
		nil,
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
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil) // Find user successfully

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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.InvalidCredentials
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
		nil,
		nil,
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
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil) // Find user successfully
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotActivated
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
		nil,
		nil,
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
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil)                                              // Find user successfully
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
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotActivated
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestActivateUserSuccess tests if ActivateUser returns 200-OK and tokens when user is activated successfully
func TestActivateUserSuccess(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
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

	updatedUser := user
	updatedUser.Activated = true

	sixDigitToken := "123456"
	activationToken := models.ActivationToken{
		Id:             uuid.New(),
		Username:       username,
		User:           user,
		Token:          sixDigitToken,
		ExpirationTime: time.Now().Add(time.Minute * 15),
	}

	activationRequest := models.UserActivationRequestDTO{
		Token: sixDigitToken,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil)                                  // Find user successfully
	mockActivationTokenRepository.On("FindActivationToken", username, sixDigitToken).Return(&activationToken, nil) // Find token successfully
	mockActivationTokenRepository.On("DeleteActivationTokenByUsername", username).Return(nil)                      // Delete token successfully
	mockUserRepository.On("UpdateUser", &updatedUser).Return(nil)                                                  // Activate user successfully
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)                                // Send mail successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(activationRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/activate", username), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/activate", userController.ActivateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status

	var loginResponse models.UserLoginResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)

	assert.NotEmpty(t, loginResponse.Token)
	assert.NotEmpty(t, loginResponse.RefreshToken)

	extractedUsername, isRefresh, err := utils.VerifyJWTToken(loginResponse.Token)
	assert.NoError(t, err)
	assert.False(t, isRefresh)
	assert.Equal(t, username, extractedUsername)

	extractedUsername, isRefresh, err = utils.VerifyJWTToken(loginResponse.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, isRefresh)
	assert.Equal(t, username, extractedUsername)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)

}

// TestActivateUserSuccess tests if ActivateUser returns 400-Bad Request when request body is invalid
func TestActivateUserBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`, // invalid field
		`{}`,                        // empty body
	}

	for _, body := range invalidBodies {
		controller := controllers.NewUserController(nil)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/users/:username/activate", controller.Login)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/users/testUser/activate", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestActivateUserAlreadyReported tests if ActivateUser returns 208-Already Reported when user is already activated
func TestActivateUserAlreadyReported(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		nil,
		nil,
		nil,
		nil,
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

	sixDigitToken := "123456"

	activationRequest := models.UserActivationRequestDTO{
		Token: sixDigitToken,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // Find user successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(activationRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/activate", username), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/activate", userController.ActivateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusAlreadyReported, w.Code) // Expect HTTP 208 Already Reported status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserAlreadyActivated
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
}

// TestActivateUserNotFound tests if ActivateUser returns 404-Not Found when user cannot be found
func TestActivateUserNotFound(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"
	sixDigitToken := "123456"
	activationRequest := models.UserActivationRequestDTO{
		Token: sixDigitToken,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&models.User{}, gorm.ErrRecordNotFound) // User not found

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(activationRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/activate", username), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/activate", userController.ActivateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestActivateUserTokenNotFound tests if ActivateUser returns 404-Not Found when token cannot be found
func TestActivateUserTokenNotFound(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
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
	activationRequest := models.UserActivationRequestDTO{
		Token: sixDigitToken,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil)                                                                   // User found successfully
	mockActivationTokenRepository.On("FindActivationToken", username, sixDigitToken).Return(&models.ActivationToken{}, gorm.ErrRecordNotFound) // Token not found

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(activationRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/activate", username), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/activate", userController.ActivateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.InvalidToken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestActivateUserTokenExpired tests if ActivateUser returns 401-Unauthorized when token is expired
func TestActivateUserTokenExpired(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
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
	activationRequest := models.UserActivationRequestDTO{
		Token: sixDigitToken,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil)                                              // User found successfully
	mockActivationTokenRepository.On("FindActivationToken", username, sixDigitToken).Return(&activationToken, nil)        // Expired token found successfully
	mockActivationTokenRepository.On("DeleteActivationTokenByUsername", username).Return(nil)                             // Delete token successfully
	mockActivationTokenRepository.On("CreateActivationToken", mock.AnythingOfType("*models.ActivationToken")).Return(nil) // Create new token successfully
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)                                       // Send mail successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(activationRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/activate", username), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/activate", userController.ActivateUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ActivationTokenExpired
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestResendActivationTokenSuccess tests if ResendToken returns 204-No Content when token is resent successfully
func TestResendActivationTokenSuccess(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
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

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil)                                              // Find user successfully
	mockActivationTokenRepository.On("DeleteActivationTokenByUsername", username).Return(nil)                             // Delete token successfully
	mockActivationTokenRepository.On("CreateActivationToken", mock.AnythingOfType("*models.ActivationToken")).Return(nil) // Create new token successfully
	mockMailService.On("SendMail", email, mock.Anything, mock.Anything).Return(nil)                                       // Send mail successfully

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/resend", username), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/users/:username/resend", userController.ResendActivationToken)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect HTTP 204 No Content status

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestResendActivationTokenAlreadyReported tests if ResendActivationToken returns 208-Already Reported when user is already activated
func TestResendActivationTokenAlreadyReported(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
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

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&user, nil) // Find user successfully

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/resend", username), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/users/:username/resend", userController.ResendActivationToken)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusAlreadyReported, w.Code) // Expect HTTP 208 Already Reported status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserAlreadyActivated
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestResendActivationTokenUserNotfound tests if ResendActivationToken returns 404-Not Found when user cannot be found
func TestResendActivationTokenUserNotfound(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)
	mockActivationTokenRepository := new(repositories.MockActivationTokenRepository)
	mockMailService := new(services.MockMailService)

	userService := services.NewUserService(
		mockUserRepository,
		mockActivationTokenRepository,
		mockMailService,
		nil,
		nil,
		nil,
	)

	userController := controllers.NewUserController(userService)

	username := "testUser"

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&models.User{}, gorm.ErrRecordNotFound) // User not found

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/resend", username), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/users/:username/resend", userController.ResendActivationToken)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found status
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatal(err)
	}

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	// Verify that all expectations are met
	mockUserRepository.AssertExpectations(t)
	mockActivationTokenRepository.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestRefreshTokenSuccess tests if RefreshToken returns 200-OK and new tokens when refresh token is valid
func TestRefreshTokenSuccess(t *testing.T) {
	// Setup
	userService := services.NewUserService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	userController := controllers.NewUserController(userService)

	currentUsername := "testUser"
	refreshToken, err := utils.GenerateRefreshToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	request := models.UserRefreshTokenRequestDTO{
		RefreshToken: refreshToken,
	}

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/users/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/refresh", userController.RefreshToken)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status

	var loginResponse models.UserLoginResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, loginResponse.Token)
	assert.NotEmpty(t, loginResponse.RefreshToken)

	extractedUsername, isRefresh, err := utils.VerifyJWTToken(loginResponse.Token)
	assert.NoError(t, err)
	assert.False(t, isRefresh)
	assert.Equal(t, currentUsername, extractedUsername)

	extractedUsername, isRefresh, err = utils.VerifyJWTToken(loginResponse.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, isRefresh)
	assert.Equal(t, currentUsername, extractedUsername)
}

// TestRefreshTokenBadRequest tests if RefreshToken returns 400-Bad Request when request body is invalid
func TestRefreshTokenBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`, // invalid field
		`{}`,                        // empty body
	}

	for _, body := range invalidBodies {
		controller := controllers.NewUserController(nil)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/users/refresh", controller.RefreshToken)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/users/refresh", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestRefreshTokenInvalidToken tests if RefreshToken returns 401-Unauthorized when refresh token is invalid
func TestRefreshTokenInvalidToken(t *testing.T) {
	accessToken, err := utils.GenerateAccessToken("testUser")
	if err != nil {
		t.Fatal(err)
	}
	invalidTokens := []string{
		"invalidToken",
		"Bearer invalidToken",
		accessToken,
	}

	for _, token := range invalidTokens {
		service := services.NewUserService(nil, nil, nil, nil, nil, nil)
		controller := controllers.NewUserController(service)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/users/refresh", controller.RefreshToken)

		// Create request
		refreshReq := models.UserRefreshTokenRequestDTO{
			RefreshToken: token,
		}
		requestBody, err := json.Marshal(refreshReq)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/users/refresh", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.InvalidToken
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestSearchUserSuccess tests if SearchUser returns 200-OK and list of users
func TestSearchUserSuccess(t *testing.T) {
	// Setup mocks
	mockUserRepository := new(repositories.MockUserRepository)

	userService := services.NewUserService(
		mockUserRepository,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	userController := controllers.NewUserController(userService)

	foundUsers := []models.User{
		{
			Username:          "testUser1",
			Nickname:          "Test User 1",
			ProfilePictureUrl: "",
		},
		{
			Username:          "testUser2",
			Nickname:          "Test User 2",
			ProfilePictureUrl: "",
		},
	}

	searchQuery := "testUser"
	limit := 10
	offset := 0

	currentUsername := "currentUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepository.On("SearchUser", searchQuery, limit, offset, currentUsername).Return(foundUsers, int64(len(foundUsers)), nil) // Find foundUsers successfully

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/foundUsers?username=%s&limit=%d&offset=%d", searchQuery, limit, offset), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/foundUsers", middleware.AuthorizeUser, userController.SearchUser)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status
	var responseDto models.UserSearchResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(t, err)

	assert.Equal(t, len(foundUsers), len(responseDto.Records))
	assert.Equal(t, int64(len(foundUsers)), responseDto.Pagination.Records)
	assert.Equal(t, limit, responseDto.Pagination.Limit)
	assert.Equal(t, offset, responseDto.Pagination.Offset)

	for i, user := range foundUsers {
		assert.Equal(t, user.Username, responseDto.Records[i].Username)
		assert.Equal(t, user.Nickname, responseDto.Records[i].Nickname)
		assert.Equal(t, user.ProfilePictureUrl, responseDto.Records[i].ProfilePictureUrl)
	}

	mockUserRepository.AssertExpectations(t)
}

// TestSearchUserBadRequest tests if SearchUser returns 400-Bad Request when request query is invalid
func TestSearchUserBadRequest(t *testing.T) {
	urls := []string{
		"/users?username=&limit=10&offset=0",               // empty username
		"/users?limit=q0&offset=0",                         // no username
		"/users?username=testUser&limit=10&offset=invalid", // invalid offset
		"/users?username=testUser&limit=invalid&offset=0",  // invalid limit
	}

	for _, url := range urls {
		controller := controllers.NewUserController(nil)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/users", controller.SearchUser)

		// Create request
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestSearchUserUnauthorized tests if SearchUser returns 401-Unauthorized when user is not authenticated
func TestSearchUserUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"invalidToken",
		"Bearer invalidToken",
		"",
	}

	for _, token := range invalidTokens {

		controller := controllers.NewUserController(nil)

		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/users", middleware.AuthorizeUser, controller.SearchUser)

		// Create request
		req, err := http.NewRequest(http.MethodGet, "/users?username=testUser&limit=10&offset=0", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestUpdateUserInformationSuccess tests if UpdateUserInformation returns 200-OK when user information is updated successfully
func TestUpdateUserInformationSuccess(t *testing.T) {
	requests := []models.UserInformationUpdateDTO{
		{
			Nickname: "New Nickname",
			Status:   "New status",
		},

		// Regression test
		// No failures when either nickname or status is empty
		{
			Nickname: "New Nickname",
			Status:   "", // New status is empty/delete nickname
		},
		{
			Nickname: "", // New nickname is empty/delete nickname
			Status:   "New status",
		},
		{
			Nickname: "", // New nickname is empty/delete nickname
			Status:   "", // New status is empty/delete nickname
		},
	}

	for _, request := range requests {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		validator := utils.NewValidator()
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			validator,
			nil,
			nil,
		)
		userController := controllers.NewUserController(userService)

		user := models.User{
			Username: "testUser",
			Nickname: "Old Nickname",
			Status:   "Old status",
		}

		authenticationToken, err := utils.GenerateAccessToken(user.Username)
		if err != nil {
			t.Fatal(err)
		}

		// Mock expectations
		var capturedUpdatedUser *models.User
		mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // Find user successfully
		mockUserRepository.On("UpdateUser", mock.AnythingOfType("*models.User")).
			Run(func(args mock.Arguments) {
				capturedUpdatedUser = args.Get(0).(*models.User)
			}).Return(nil) // Update user successfully

		// Setup HTTP request and recorder
		requestBody, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PUT("/users", middleware.AuthorizeUser, userController.UpdateUserInformation)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status

		var responseDto models.UserInformationUpdateDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDto)
		assert.NoError(t, err)

		assert.Equal(t, request.Nickname, responseDto.Nickname)
		assert.Equal(t, request.Status, responseDto.Status)
		assert.Equal(t, request.Nickname, capturedUpdatedUser.Nickname)
		assert.Equal(t, request.Status, capturedUpdatedUser.Status)

		mockUserRepository.AssertExpectations(t)
	}
}

// TestUpdateUserInformationBadRequest tests if UpdateUserInformation returns 400-Bad Request when request body is invalid
func TestUpdateUserInformationBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,  // invalid field
		`{}`,                         // empty body
		`{status: "New status"}`,     // no nickname
		`{nickname: "New nickname"}`, // no status
		`{status: "` + strings.Repeat("a", 256) + `", nickname: "New nickname"}`, // status too long
		`{status: "New status", nickname: "` + strings.Repeat("a", 256) + `"}`,   // nickname too long
	}

	for _, body := range invalidBodies {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		validator := utils.NewValidator()
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			validator,
			nil,
			nil,
		)

		userController := controllers.NewUserController(userService)

		user := models.User{
			Username: "testUser",
			Nickname: "Old Nickname",
			Status:   "Old status",
		}

		authenticationToken, err := utils.GenerateAccessToken(user.Username)
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PUT("/users", middleware.AuthorizeUser, userController.UpdateUserInformation)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
	}
}

// TestUpdateUserInformationUnauthorized tests if UpdateUserInformation returns 401-Unauthorized when user is not authorized
func TestUpdateUserInformationUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"invalidToken",
		"Bearer invalidToken",
		"",
	}
	for _, token := range invalidTokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		validator := utils.NewValidator()
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			validator,
			nil,
			nil,
		)

		userController := controllers.NewUserController(userService)

		userRequest := models.UserInformationUpdateDTO{
			Nickname: "New Nickname",
			Status:   "New status",
		}

		// Setup HTTP request and recorder
		requestBody, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PUT("/users", middleware.AuthorizeUser, userController.UpdateUserInformation)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatal(err)
		}

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
	}
}

// TestChangePasswordSuccess tests if ChangePassword returns 204-No Content when password is changed successfully
func TestChangePasswordSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	validator := utils.NewValidator()
	userService := services.NewUserService(
		mockUserRepository,
		nil,
		nil,
		validator,
		nil,
		nil,
	)
	userController := controllers.NewUserController(userService)

	userRequest := models.ChangePasswordDTO{
		OldPassword: "Old password123!",
		NewPassword: "New password123!",
	}

	hashedOldPassword, err := utils.HashPassword(userRequest.OldPassword)
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		Username:     "testUser",
		PasswordHash: hashedOldPassword,
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	var capturedUpdatedUser *models.User
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // Find user successfully
	mockUserRepository.On("UpdateUser", mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			capturedUpdatedUser = args.Get(0).(*models.User)
		}).Return(nil) // Update user successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPatch, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect HTTP 204 No Content status

	assert.Nil(t, w.Body.Bytes())
	success := utils.CheckPassword(userRequest.NewPassword, capturedUpdatedUser.PasswordHash)
	assert.True(t, success)

	mockUserRepository.AssertExpectations(t)
}

// TestChangePasswordBadRequest tests if ChangePassword returns 400-Bad Request when request body is invalid
func TestChangePasswordBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,         // invalid field
		`{}`,                                // empty body
		`{oldPassword: "Old password123!"}`, // no new password
		`{newPassword: "New password123!"}`, // no old password
		`{oldPassword: "Old password123!", newPassword: "` + strings.Repeat("a", 256) + `"}`, // new password does not meet specifications
	}

	for _, body := range invalidBodies {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		validator := utils.NewValidator()
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			validator,
			nil,
			nil,
		)
		userController := controllers.NewUserController(userService)

		username := "testUser"
		authenticationToken, err := utils.GenerateAccessToken(username)
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodPatch, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
	}
}

// TestChangePasswordUnauthorized tests if ChangePassword returns 401-Unauthorized when user is not authorized
func TestChangePasswordUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"invalidToken",
		"Bearer invalidToken",
		"",
	}

	for _, token := range invalidTokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		validator := utils.NewValidator()
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			validator,
			nil,
			nil,
		)

		userController := controllers.NewUserController(userService)

		userRequest := models.ChangePasswordDTO{
			OldPassword: "Old password123!",
			NewPassword: "New password123!",
		}

		// Setup HTTP request and recorder
		requestBody, err := json.Marshal(userRequest)
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequest(http.MethodPatch, "/users", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatal(err)
		}

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
	}
}

// TestChangePasswordOldPasswordIncorrect tests if ChangePassword returns 403-Forbidden when old password is incorrect
func TestChangePasswordOldPasswordIncorrect(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	validator := utils.NewValidator()

	userService := services.NewUserService(
		mockUserRepository,
		nil,
		nil,
		validator,
		nil,
		nil,
	)
	userController := controllers.NewUserController(userService)

	userRequest := models.ChangePasswordDTO{
		OldPassword: "Wrong old password123!",
		NewPassword: "New password123!",
	}

	hashedOldPassword, err := utils.HashPassword("Old password123!")
	if err != nil {
		t.Fatal(err)
	}

	user := models.User{
		Username:     "testUser",
		PasswordHash: hashedOldPassword,
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // Find user successfully

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPatch, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect HTTP 403 Forbidden status
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.OldPasswordIncorrect
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepository.AssertExpectations(t)
}

// TestGetUserProfileSuccess tests if GetUserProfile returns 200-OK when user profile is retrieved successfully
func TestGetUserProfileSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockSubscriptionRepository := new(repositories.MockSubscriptionRepository)
	userService := services.NewUserService(
		mockUserRepository,
		nil,
		nil,
		nil,
		mockPostRepository,
		mockSubscriptionRepository,
	)
	userController := controllers.NewUserController(userService)

	user := models.User{
		Username:          "testUser",
		Nickname:          "Test User",
		Status:            "Test status",
		ProfilePictureUrl: "",
	}
	postCount := int64(64)
	followerCount := int64(1)
	followingCount := int64(1)

	currentUsername := "currentUsername"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	subscription := models.Subscription{
		Id:                uuid.New(),
		SubscriptionDate:  time.Now(),
		FollowerUsername:  currentUsername,
		FollowingUsername: user.Username,
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // Find user successfully
	mockPostRepository.On("GetPostCountByUsername", user.Username).Return(postCount, nil)
	mockSubscriptionRepository.On("GetSubscriptionByUsernames", currentUsername, user.Username).Return(&subscription, nil)
	mockSubscriptionRepository.On("GetSubscriptionCountByUsername", user.Username).Return(followerCount, followingCount, nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/users/testUser", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/users/:username", middleware.AuthorizeUser, userController.GetUserProfile)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK status

	var responseDto models.UserProfileResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, responseDto.Username)
	assert.Equal(t, user.Nickname, responseDto.Nickname)
	assert.Equal(t, user.Status, responseDto.Status)
	assert.Equal(t, user.ProfilePictureUrl, responseDto.ProfilePictureUrl)
	assert.Equal(t, postCount, responseDto.Posts)
	assert.Equal(t, followerCount, responseDto.Follower)
	assert.Equal(t, followingCount, responseDto.Following)
	assert.Equal(t, subscription.Id.String(), responseDto.SubscriptionId)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
}

// TestGetUserProfileUnauthorized tests if GetUserProfile returns 401-Unauthorized when user is not logged-in
func TestGetUserProfileUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"invalidToken",
		"Bearer Invalid Token",
		"",
	}
	for _, token := range invalidTokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		userService := services.NewUserService(
			mockUserRepository,
			nil,
			nil,
			nil,
			mockPostRepository,
			nil,
		)
		userController := controllers.NewUserController(userService)

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodGet, "/users/testUser", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/users/:username", middleware.AuthorizeUser, userController.GetUserProfile)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status

		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetUserProfileUserNotFound tests if GetUserProfile returns 404-Not Found when there is no user to the given username
func TestGetUserProfileUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	userService := services.NewUserService(
		mockUserRepository,
		nil,
		nil,
		nil,
		mockPostRepository,
		nil,
	)
	userController := controllers.NewUserController(userService)

	currentUsername := "currentUsername"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	queryUsername := "queryUsername"

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", queryUsername).Return(&models.User{}, gorm.ErrRecordNotFound)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/users/"+queryUsername, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/users/:username", middleware.AuthorizeUser, userController.GetUserProfile)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found status

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
}
