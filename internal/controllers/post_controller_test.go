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
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestCreatePostWithLocationSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully with location
func TestCreatePostWithLocationSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLocationRepository := new(repositories.MockLocationRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		new(services.ImageService),
		validator,
		mockLocationRepository,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    true,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	coordinate := 11.1
	accuracy := uint(50)

	content := "This is a test #post. #postings_are_fun"
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content: content,
		Location: &models.LocationDTO{
			Longitude: &coordinate,
			Latitude:  &coordinate,
			Accuracy:  &accuracy,
		},
	}

	expectedHashtagOne := models.Hashtag{
		Id:   uuid.New(),
		Name: "post",
	}
	expectedHashtagTwo := models.Hashtag{
		Id:   uuid.New(),
		Name: "postings_are_fun",
	}

	// Mock expectations
	var capturedPost *models.Post
	var capturedLocation *models.Location
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully
	mockLocationRepository.On("CreateLocation", mock.AnythingOfType("*models.Location")).
		Run(func(args mock.Arguments) {
			capturedLocation = args.Get(0).(*models.Location) // Save argument to captor
		}).Return(nil) // Location created successfully
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagOne.Name).Return(expectedHashtagOne, nil)
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagTwo.Name).Return(expectedHashtagTwo, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(postCreateRequestDTO)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 created
	var responsePost models.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePost)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, capturedPost.Username)
	assert.Equal(t, postCreateRequestDTO.Content, capturedPost.Content)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, capturedLocation.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, capturedLocation.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, capturedLocation.Accuracy)
	assert.Equal(t, capturedLocation.Id, *capturedPost.LocationId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Empty(t, capturedPost.ImageUrl)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.NotNil(t, responsePost.Location)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, *responsePost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, *responsePost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, *responsePost.Location.Accuracy)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockLocationRepository.AssertExpectations(t)
}

// Regression Test
// TestCreatePostWithLocationZeroValues tests if a post is created successfully with location where accuracy, latitude and longitude are all zero
func TestCreatePostWithLocationZeroValues(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLocationRepository := new(repositories.MockLocationRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		new(services.ImageService),
		validator,
		mockLocationRepository,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    true,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	coordinate := 0.0
	accuracy := uint(0)

	content := "This is a test #post. #postings_are_fun"
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content: content,
		Location: &models.LocationDTO{
			Longitude: &coordinate,
			Latitude:  &coordinate,
			Accuracy:  &accuracy,
		},
	}

	expectedHashtagOne := models.Hashtag{
		Id:   uuid.New(),
		Name: "post",
	}
	expectedHashtagTwo := models.Hashtag{
		Id:   uuid.New(),
		Name: "postings_are_fun",
	}

	// Mock expectations
	var capturedPost *models.Post
	var capturedLocation *models.Location
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully
	mockLocationRepository.On("CreateLocation", mock.AnythingOfType("*models.Location")).
		Run(func(args mock.Arguments) {
			capturedLocation = args.Get(0).(*models.Location) // Save argument to captor
		}).Return(nil) // Location created successfully
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagOne.Name).Return(expectedHashtagOne, nil)
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagTwo.Name).Return(expectedHashtagTwo, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(postCreateRequestDTO)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 created
	var responsePost models.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePost)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, capturedPost.Username)
	assert.Equal(t, postCreateRequestDTO.Content, capturedPost.Content)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, capturedLocation.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, capturedLocation.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, capturedLocation.Accuracy)
	assert.Equal(t, capturedLocation.Id, *capturedPost.LocationId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Empty(t, capturedPost.ImageUrl)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.NotNil(t, responsePost.Location)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, *responsePost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, *responsePost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, *responsePost.Location.Accuracy)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockLocationRepository.AssertExpectations(t)
}

// TestCreatePostWithoutLocationSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully without location
func TestCreatePostWithoutLocationSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLocationRepository := new(repositories.MockLocationRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		new(services.ImageService),
		validator,
		mockLocationRepository,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    true,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	content := "This is a test #post. #postings_are_fun"
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content: content,
	}

	expectedHashtagOne := models.Hashtag{
		Id:   uuid.New(),
		Name: "post",
	}
	expectedHashtagTwo := models.Hashtag{
		Id:   uuid.New(),
		Name: "postings_are_fun",
	}

	// Mock expectations
	var capturedPost *models.Post
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagOne.Name).Return(expectedHashtagOne, nil)
	mockHashtagRepository.On("FindOrCreateHashtag", expectedHashtagTwo.Name).Return(expectedHashtagTwo, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(postCreateRequestDTO)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 created
	var responsePost models.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePost)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, capturedPost.Username)
	assert.Equal(t, postCreateRequestDTO.Content, capturedPost.Content)
	assert.Nil(t, capturedPost.LocationId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Empty(t, capturedPost.ImageUrl)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Location)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockLocationRepository.AssertExpectations(t)
}

// TestCreatePostBadRequest tests if the CreatePost function returns a 400 bad request if the content is empty
func TestCreatePostBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,                       // invalid body
		`{"content": ""}`,                                 // empty content
		`{"content": "` + strings.Repeat("A", 300) + `"}`, // content too long
		"", // empty body
		`{"content: "test", "location":{"latitude": "abc", "longitude": "abc2", "accuracy": 11}`,  // invalid coordinates
		`{"content: "test", "location":{"latitude": "11.2", "longitude": "13.2", "accuracy": 11}`, // string coordinates
	}

	for _, body := range invalidBodies {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			new(services.ImageService),
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		authenticationToken, err := utils.GenerateAccessToken("testUser")
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 bad request
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockUserRepository.AssertExpectations(t)
		mockPostRepository.AssertExpectations(t)
		mockHashtagRepository.AssertExpectations(t)
	}
}

// TestCreatePostUnauthorized tests if the CreatePost function returns a 401 unauthorized if the user is not authenticated
func TestCreatePostUnauthorized(t *testing.T) {
	nonExistingUserToken, err := utils.GenerateAccessToken("nonExistingUser")
	if err != nil {
		t.Fatal(err)
	}

	accessTokens := []string{
		"",                       // empty token
		nonExistingUserToken,     // some access token whose user does not exist
		strings.Repeat("A", 300), // some text
	}

	for _, token := range accessTokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			new(services.ImageService),
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		mockUserRepository.On("FindUserByUsername", mock.AnythingOfType("string")).Return(&models.User{}, gorm.ErrRecordNotFound) // User not found

		// Setup HTTP request
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(`{"content": "This is the body"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
		mockHashtagRepository.AssertExpectations(t)
	}
}

// TestCreatePostUserNotActivated tests if the CreatePost function returns a 401 unauthorized if the user is not activated
func TestCreatePostUserNotActivated(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		new(services.ImageService),
		new(utils.Validator),
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    false,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	coordinate := 11.1
	accuracy := uint(50)

	content := "This is a test #post. #postings_are_fun"
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content: content,
		Location: &models.LocationDTO{
			Longitude: &coordinate,
			Latitude:  &coordinate,
			Accuracy:  &accuracy,
		},
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully

	// Setup HTTP request
	requestBody, err := json.Marshal(postCreateRequestDTO)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
}

func createFormFile(writer *multipart.Writer, fieldName, fileName string, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	h.Set("Content-Type", contentType)

	return writer.CreatePart(h)
}

// TestCreatePostWithImageSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully with image
func TestCreatePostWithImageSuccess(t *testing.T) {
	testCases := [][]string{
		{"This is a test post.", "../../tests/resources/valid.jpeg", "image/jpeg"},      // test jpeg
		{"This is a test post text.", "../../tests/resources/valid.webp", "image/webp"}, // test webp
		{"", "../../tests/resources/valid.jpeg", "image/jpeg"},                          // test only image
	}

	for _, testCase := range testCases {
		content := testCase[0]
		testImageFilePath := testCase[1]
		testImageContentType := testCase[2]

		// Read the image file
		imageData, err := os.ReadFile(testImageFilePath)
		if err != nil {
			t.Fatalf("Failed to read image file: %s", err)
		}

		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)
		mockFileSystem := new(repositories.MockFileSystem)
		validator := new(utils.Validator)

		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			services.NewImageService(mockFileSystem, validator),
			validator,
			nil,
		)
		postController := controllers.NewPostController(postService)

		user := models.User{
			Username:     "testUser",
			Nickname:     "testNickname",
			Email:        "test@domain.com",
			PasswordHash: "passwordHash",
			CreatedAt:    time.Now().Add(time.Hour * -24),
			Activated:    true,
		}
		authenticationToken, err := utils.GenerateAccessToken(user.Username)
		if err != nil {
			t.Fatal(err)
		}

		// Mock expectations
		var capturedPost *models.Post
		var capturedFile []uint8
		var capturedFilename string
		mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
		mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
			Run(func(args mock.Arguments) {
				capturedPost = args.Get(0).(*models.Post) // Save argument to captor
			}).Return(nil) // Post created successfully
		mockFileSystem.On("DoesFileExist", mock.AnythingOfType("string")).Return(false, nil)
		mockFileSystem.On("WriteFile", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("fs.FileMode")).
			Run(func(args mock.Arguments) {
				capturedFilename = args.Get(0).(string) // Save argument to captor
				capturedFile = args.Get(1).([]uint8)    // Save argument to captor
			}).Return(nil)

		// Create multipart request body
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add text field
		if content != "" {
			err = writer.WriteField("content", content)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Add image file
		part, err := createFormFile(writer, "image", testImageFilePath, testImageContentType)
		if err != nil {
			t.Fatal(err)
		}

		_, err = part.Write(imageData)
		if err != nil {
			t.Fatal(err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request
		req, _ := http.NewRequest("POST", "/posts", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 created
		var responsePost models.PostResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responsePost)
		assert.NoError(t, err)

		assert.Equal(t, capturedPost.Id, responsePost.PostId)
		assert.NotNil(t, capturedPost.Id)
		assert.Equal(t, user.Username, responsePost.Author.Username)
		assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
		assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
		assert.Equal(t, content, responsePost.Content)
		assert.Equal(t, content, capturedPost.Content)
		assert.Equal(t, user.Username, capturedPost.Username)
		assert.NotNil(t, capturedPost.CreatedAt)
		assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
		assert.Equal(t, "/api/images"+capturedFilename, capturedPost.ImageUrl)
		assert.True(t, reflect.DeepEqual(imageData, capturedFile))

		mockPostRepository.AssertExpectations(t)
		mockHashtagRepository.AssertExpectations(t)
		mockFileSystem.AssertExpectations(t)
		mockUserRepository.AssertExpectations(t)
	}
}

// TestCreatePostWithImageBadRequest tests if the CreatePost function returns a 400 bad request if image is not webp or jpeg
func TestCreatePostWithImageBadRequest(t *testing.T) {

	// Create multipart request body
	content := "Some text"
	testImageName := "InvalidImageType.png"
	testImageContent := "Some image content"
	testImageContentType := "image/png"

	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	mockFileSystem := new(repositories.MockFileSystem)
	validator := new(utils.Validator)

	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		services.NewImageService(mockFileSystem, validator),
		validator,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    true,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully

	// Create multipart request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("content", content)
	part, err := createFormFile(writer, "image", testImageName, testImageContentType)

	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(testImageContent))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Setup HTTP request
	req, _ := http.NewRequest("POST", "/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 Bad Request
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.BadRequest
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockFileSystem.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)

}

// TestCreatePostWithEmptyImageSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully with empty image
func TestCreatePostWithEmptyImageSuccess(t *testing.T) {
	// Create multipart request body
	content := "Some text"
	testImageName := ""
	testImageContent := ""
	testImageContentType := ""

	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	mockFileSystem := new(repositories.MockFileSystem)
	validator := new(utils.MockValidator)

	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		services.NewImageService(mockFileSystem, validator),
		validator,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().Add(time.Hour * -24),
		Activated:    true,
	}
	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	var capturedPost *models.Post
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully

	// Create multipart request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("content", content)
	part, err := createFormFile(writer, "image", testImageName, testImageContentType)

	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(testImageContent))
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Setup HTTP request
	req, _ := http.NewRequest("POST", "/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 created
	var responsePost models.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePost)
	assert.NoError(t, err)

	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.NotNil(t, capturedPost.Id)
	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, content, capturedPost.Content)
	assert.Equal(t, user.Username, capturedPost.Username)
	assert.NotNil(t, capturedPost.CreatedAt)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Equal(t, capturedPost.ImageUrl, "")

	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockFileSystem.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// Regression Test
// TestCreatePostWithWrongContentTypeBadRequest tests if the CreatePost function returns a 400 bad request if the content type is not multipart/form-data or application/json
func TestCreatePostWithWrongContentTypeBadRequest(t *testing.T) {
	for _, contentType := range []string{
		"application/xml",
		"text/plain",
		"application/pdf",
	} {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)
		mockFileSystem := new(repositories.MockFileSystem)
		validator := new(utils.Validator)

		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			services.NewImageService(mockFileSystem, validator),
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		username := "testUser"
		authenticationToken, err := utils.GenerateAccessToken(username)
		if err != nil {
			t.Fatal(err)
		}

		// Set up HTTP request
		req, _ := http.NewRequest("POST", "/posts", bytes.NewBufferString(`{"content": "This is the body"}`))
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 Bad Request
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestGetPostsByUsernameSuccess tests if the GetPostsByUsername function returns a list of posts and 200 ok if the user exists
func TestGetPostsByUsernameSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
		Email:    "test@example.com",
	}

	locationId := uuid.New()
	posts := []models.Post{
		{
			Id:         uuid.New(),
			Username:   user.Username,
			Content:    "Test Post 1",
			CreatedAt:  time.Now(),
			LocationId: &locationId,
			Location: models.Location{
				Longitude: 11.1,
				Latitude:  22.2,
				Accuracy:  50,
			},
		},
		{
			Id:         uuid.New(),
			Username:   user.Username,
			Content:    "Test Post 2",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			LocationId: nil,
			Location:   models.Location{},
		},
	}

	currentUsername := "someOtherUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	limit := 10
	offset := 0

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("GetPostsByUsername", user.Username, 0, 10).Return(posts, int64(len(posts)), nil)

	// Setup HTTP request
	url := "/users/" + user.Username + "/feed?offset=" + fmt.Sprint(offset) + "&limit=" + fmt.Sprint(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/users/:username/feed", middleware.AuthorizeUser, postController.GetPostsByUserUsername)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, posts[0].Id.String(), response.Records[0].PostId)
	assert.Equal(t, posts[0].Content, response.Records[0].Content)
	assert.True(t, posts[0].CreatedAt.Equal(response.Records[0].CreationDate))
	assert.NotNil(t, response.Records[0].Location)
	assert.Equal(t, posts[0].Location.Latitude, *response.Records[0].Location.Latitude)
	assert.Equal(t, posts[0].Location.Longitude, *response.Records[0].Location.Longitude)
	assert.Equal(t, posts[0].Location.Accuracy, *response.Records[0].Location.Accuracy)

	assert.Equal(t, posts[1].Id.String(), response.Records[1].PostId)
	assert.Equal(t, posts[1].Content, response.Records[1].Content)
	assert.True(t, posts[1].CreatedAt.Equal(response.Records[1].CreationDate))
	assert.Nil(t, response.Records[1].Location)

	assert.Equal(t, offset, response.Pagination.Offset)
	assert.Equal(t, limit, response.Pagination.Limit)
	assert.Equal(t, int64(len(posts)), response.Pagination.Records)

	// Validate Mock Expectations
	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
}

// TestGetPostsByUsernameUnauthorized tests if the GetPostsByUserUsername function returns a 401 unauthorized if the user is not authenticated
func TestGetPostsByUsernameUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"",                 // empty token
		"someInvalidToken", // some invalid token
	}

	for _, token := range invalidTokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			nil,
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		// Setup HTTP request
		req, _ := http.NewRequest("GET", "/users/testUser/feed", nil)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/users/:username/feed", middleware.AuthorizeUser, postController.GetPostsByUserUsername)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestFindPostsByUserUsernameNotFound tests if the GetPostsByUserUsername function returns a 404 not found if the user does not exist
func TestGetPostsByUsernameUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	username := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepository.On("FindUserByUsername", username).Return(&models.User{}, gorm.ErrRecordNotFound) // User not found

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/users/testUser/feed", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/users/:username/feed", middleware.AuthorizeUser, postController.GetPostsByUserUsername)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 not found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestGetGlobalPostFeedSuccess tests if the GetPostFeed function returns a post feed and 200 ok if the request is valid
func TestGetGlobalPostFeedSuccess(t *testing.T) {
	validToken, err := utils.GenerateAccessToken("someUser")
	if err != nil {
		t.Fatal(err)
	}
	tokens := []string{
		"",                 // empty token
		"someInvalidToken", // some invalid token
		validToken,         // some valid token
	}

	for _, token := range tokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			nil,
			nil,
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		lastPost := models.Post{
			Id:        uuid.New(),
			Username:  "someUserTest",
			User:      models.User{},
			Content:   "This is the last post",
			ImageUrl:  "",
			CreatedAt: time.Now().Add(time.Hour * -1),
		}

		locationId := uuid.New()
		nextPosts := []models.Post{
			{
				Id:       uuid.New(),
				Username: "someOtherUsername",
				User: models.User{
					Username:          "someOtherUsername",
					Nickname:          "someOtherNickname",
					ProfilePictureUrl: "",
				},
				Content:    "This is the next post",
				ImageUrl:   "",
				CreatedAt:  time.Now().Add(time.Hour * -2),
				LocationId: &locationId,
				Location: models.Location{
					Longitude: 11.1,
					Latitude:  22.2,
					Accuracy:  50,
				},
			},
			{
				Id:       uuid.New(),
				Username: "anotherTestUsername",
				User: models.User{
					Username:          "anotherTestUsername",
					Nickname:          "anotherTestNickname",
					ProfilePictureUrl: "",
				},
				Content:   "This is another next post",
				ImageUrl:  "",
				CreatedAt: time.Now().Add(time.Hour * -3),
			},
		}
		limit := 2
		totalCount := int64(2)

		// Mock expectations
		var capturedLastPost *models.Post
		mockPostRepository.On("GetPostById", lastPost.Id.String()).Return(lastPost, nil) // Post found successfully
		mockPostRepository.On("GetPostsGlobalFeed", mock.AnythingOfType("*models.Post"), limit).
			Run(func(args mock.Arguments) {
				capturedLastPost = args.Get(0).(*models.Post) // Save argument to captor
			}).Return(nextPosts, totalCount, nil) // Posts returned successfully

		// Setup HTTP request
		url := "/feed?postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit)
		if token != "" {
			url += "&feedType=global"
		}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/feed", postController.GetPostFeed)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok
		var responsePostFeed models.GeneralFeedDTO
		err := json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
		assert.NoError(t, err)

		assert.Equal(t, lastPost.Id, capturedLastPost.Id)
		assert.Equal(t, lastPost.Username, capturedLastPost.Username)
		assert.Equal(t, lastPost.Content, capturedLastPost.Content)
		assert.Equal(t, lastPost.ImageUrl, capturedLastPost.ImageUrl)
		assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
		assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

		assert.Equal(t, nextPosts[0].Id, responsePostFeed.Records[0].PostId)
		assert.Equal(t, nextPosts[0].Username, responsePostFeed.Records[0].Author.Username)
		assert.Equal(t, nextPosts[0].Content, responsePostFeed.Records[0].Content)
		assert.True(t, nextPosts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
		assert.NotNil(t, responsePostFeed.Records[0].Location)
		assert.Equal(t, nextPosts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
		assert.Equal(t, nextPosts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
		assert.Equal(t, nextPosts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

		assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
		assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
		assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
		assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
		assert.Nil(t, responsePostFeed.Records[1].Location)

		assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
		assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
		assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetGlobalPostFeedDefaultParameters tests if the GetPostFeed function returns an empty list when last post is not found and default parameters are used
func TestGetGlobalPostFeedDefaultParameters(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	totalRecords := int64(10)

	// Mock expectations
	mockPostRepository.On("GetPostById", mock.AnythingOfType("string")).Return(models.Post{}, gorm.ErrRecordNotFound) // Post not found
	mockPostRepository.On("GetPostsGlobalFeed", mock.AnythingOfType("*models.Post"), 10).Return([]models.Post{}, totalRecords, nil)

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/feed?postId=invalid&limit=invalid", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", postController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response models.GeneralFeedDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, len(response.Records) == 0)
	assert.Equal(t, totalRecords, response.Pagination.Records)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, "", response.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPersonalPostFeedSuccess tests if the GetPostFeed function returns a post feed and 200 ok if the request is valid
func TestGetPersonalPostFeedSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	lastPost := models.Post{
		Id:        uuid.New(),
		Username:  "someUserTest",
		User:      models.User{},
		Content:   "This is the last post",
		ImageUrl:  "",
		CreatedAt: time.Now().Add(time.Hour * -1),
	}

	currentUsername := "thisUser"
	token, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	locationId := uuid.New()
	nextPosts := []models.Post{
		{
			Id:       uuid.New(),
			Username: "someOtherUsername",
			User: models.User{
				Username:          "someOtherUsername",
				Nickname:          "someOtherNickname",
				ProfilePictureUrl: "",
			},
			Content:    "This is the next post",
			ImageUrl:   "",
			CreatedAt:  time.Now().Add(time.Hour * -2),
			LocationId: &locationId,
			Location: models.Location{
				Longitude: 11.1,
				Latitude:  22.2,
				Accuracy:  50,
			},
		},
		{
			Id:       uuid.New(),
			Username: "anotherTestUsername",
			User: models.User{
				Username:          "anotherTestUsername",
				Nickname:          "anotherTestNickname",
				ProfilePictureUrl: "",
			},
			Content:   "This is another next post",
			ImageUrl:  "",
			CreatedAt: time.Now().Add(time.Hour * -3),
		},
	}
	limit := 2
	totalCount := int64(2)

	// Mock expectations
	var capturedLastPost *models.Post
	mockPostRepository.On("GetPostById", lastPost.Id.String()).Return(lastPost, nil) // Post found successfully
	mockPostRepository.On("GetPostsPersonalFeed", currentUsername, mock.AnythingOfType("*models.Post"), limit).
		Run(func(args mock.Arguments) {
			capturedLastPost = args.Get(1).(*models.Post) // Save argument to captor
		}).Return(nextPosts, totalCount, nil) // Posts returned successfully

	// Setup HTTP request
	url := "/feed?postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit) + "&feedType=personal"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", postController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok
	var responsePostFeed models.GeneralFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
	assert.NoError(t, err)

	assert.Equal(t, lastPost.Id, capturedLastPost.Id)
	assert.Equal(t, lastPost.Username, capturedLastPost.Username)
	assert.Equal(t, lastPost.Content, capturedLastPost.Content)
	assert.Equal(t, lastPost.ImageUrl, capturedLastPost.ImageUrl)
	assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
	assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

	assert.Equal(t, nextPosts[0].Id, responsePostFeed.Records[0].PostId)
	assert.Equal(t, nextPosts[0].Username, responsePostFeed.Records[0].Author.Username)
	assert.Equal(t, nextPosts[0].Content, responsePostFeed.Records[0].Content)
	assert.True(t, nextPosts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
	assert.NotNil(t, responsePostFeed.Records[0].Location)
	assert.Equal(t, nextPosts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
	assert.Equal(t, nextPosts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
	assert.Equal(t, nextPosts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

	assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
	assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
	assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
	assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
	assert.Nil(t, responsePostFeed.Records[1].Location)

	assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
	assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
	assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPersonalPostFeedDefaultParameters tests if the GetPostFeed function returns a 200 OK and an empty list when last post is not found and default parameters are used
func TestGetPersonalPostFeeDefaultParameters(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	username := "thisUser"
	token, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Fatal(err)
	}

	postCount := int64(10)
	defaultLimit := 10

	// Mock expectations
	mockPostRepository.On("GetPostById", mock.AnythingOfType("string")).Return(models.Post{}, gorm.ErrRecordNotFound) // Post not found
	mockPostRepository.On("GetPostsPersonalFeed", mock.AnythingOfType("string"), mock.AnythingOfType("*models.Post"), defaultLimit).Return([]models.Post{}, postCount, nil)

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/feed?postId=invalid&feedType=personal", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", postController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response models.GeneralFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, len(response.Records) == 0)
	assert.Equal(t, postCount, response.Pagination.Records)
	assert.Equal(t, defaultLimit, response.Pagination.Limit)
	assert.Equal(t, "", response.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPersonalPostFeedUnauthorized tests if the GetPostFeed function returns a 401 unauthorized if the user is not authenticated and requests a personal feed
func TestGetPersonalPostFeedUnauthorized(t *testing.T) {
	tokens := []string{
		"",                 // empty token
		"someInvalidToken", // some invalid token
	}
	for _, token := range tokens {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			nil,
			nil,
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		// Setup HTTP request
		req, _ := http.NewRequest("GET", "/feed?postId="+uuid.New().String()+"&limit=10&feedType=personal", nil)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/feed", postController.GetPostFeed)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetPostByIdSuccess tests if the GetPostById function returns a post and 204 no content if the request is valid
func TestDeletePostSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil)
	postController := controllers.NewPostController(postService)

	postId := uuid.New().String()
	username := "testUser"
	authenticationToken, _ := utils.GenerateAccessToken(username)

	mockPostRepository.On("GetPostById", postId).Return(models.Post{Username: username}, nil)
	mockPostRepository.On("DeletePostById", postId).Return(nil)

	// Setup HTTP request
	req, _ := http.NewRequest("DELETE", "/posts/"+postId, nil)
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	mockPostRepository.AssertExpectations(t)
}

// TestDeletePostUnauthorized verifies response for deletion requests without valid authentication.
func TestDeletePostUnauthorized(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	postId := uuid.New()

	// Setup HTTP request ohne Authorization Header
	req, _ := http.NewRequest("DELETE", "/posts/"+postId.String(), nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockPostRepository.AssertExpectations(t)
}

// TestDeletePostForbidden checks for forbidden access when a user tries to delete others' posts.
func TestDeletePostForbidden(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil)
	postController := controllers.NewPostController(postService)

	postId := uuid.New().String()
	username := "testUser"
	authenticationToken, _ := utils.GenerateAccessToken(username)

	mockPostRepository.On("GetPostById", postId).Return(models.Post{Username: "anotherUser"}, nil)

	// Setup HTTP request
	req, _ := http.NewRequest("DELETE", "/posts/"+postId, nil)
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)

	mockPostRepository.AssertExpectations(t)
}

// TestDeletePostNotFound verifies response when a post to delete is not found.
func TestDeletePostNotFound(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil)
	postController := controllers.NewPostController(postService)

	postId := uuid.New().String()
	username := "testUser"
	authenticationToken, _ := utils.GenerateAccessToken(username)

	mockPostRepository.On("GetPostById", postId).Return(models.Post{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	req, _ := http.NewRequest("DELETE", "/posts/"+postId, nil)
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPostsByHashtagSuccess tests if the GetPostsByHashtag function returns a list of posts and 200 ok if the hashtag exists
func TestGetPostsByHashtagSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	hashtag := "Post"
	locationId := uuid.New()
	posts := []models.Post{
		{
			Id:       uuid.New(),
			Username: "testUser",
			User: models.User{
				Username:          "testUser",
				Nickname:          "testNickname",
				ProfilePictureUrl: "",
			},
			Content:    "Test #Post 2",
			CreatedAt:  time.Now(),
			LocationId: &locationId,
			Location: models.Location{
				Longitude: 11.1,
				Latitude:  22.2,
				Accuracy:  50,
			},
		},
		{
			Id:       uuid.New(),
			Username: "testUser",
			User: models.User{
				Username:          "testUser",
				Nickname:          "testNickname",
				ProfilePictureUrl: "",
			},
			Content:   "Test #Post 3",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	currentUsername := "someOtherUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	totalCount := int64(4)
	limit := 2
	lastPost := models.Post{
		Id:        uuid.New(),
		Username:  "testUser",
		Content:   "Test #Post 1",
		ImageUrl:  "",
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Mock expectations
	var capturedLastPost *models.Post
	mockPostRepository.On("GetPostById", lastPost.Id.String()).Return(lastPost, nil)
	mockPostRepository.On("GetPostsByHashtag", hashtag, &lastPost, limit).
		Run(func(args mock.Arguments) {
			capturedLastPost = args.Get(1).(*models.Post) // Save argument to captor
		}).Return(posts, totalCount, nil)

	// Setup HTTP request
	url := "/posts?q=" + hashtag + "&postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts", middleware.AuthorizeUser, postController.GetPostsByHashtag)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok

	var responsePostFeed models.GeneralFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
	assert.NoError(t, err)

	assert.Equal(t, lastPost.Id, capturedLastPost.Id)
	assert.Equal(t, lastPost.Username, capturedLastPost.Username)
	assert.Equal(t, lastPost.Content, capturedLastPost.Content)
	assert.Equal(t, lastPost.ImageUrl, capturedLastPost.ImageUrl)
	assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
	assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

	assert.Equal(t, posts[0].Id, responsePostFeed.Records[0].PostId)
	assert.Equal(t, posts[0].Username, responsePostFeed.Records[0].Author.Username)
	assert.Equal(t, posts[0].User.Nickname, responsePostFeed.Records[0].Author.Nickname)
	assert.Equal(t, posts[0].User.ProfilePictureUrl, responsePostFeed.Records[0].Author.ProfilePictureUrl)
	assert.Equal(t, posts[0].Content, responsePostFeed.Records[0].Content)
	assert.True(t, posts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
	assert.NotNil(t, responsePostFeed.Records[0].Location)
	assert.Equal(t, posts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
	assert.Equal(t, posts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
	assert.Equal(t, posts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

	assert.Equal(t, posts[1].Id, responsePostFeed.Records[1].PostId)
	assert.Equal(t, posts[1].Username, responsePostFeed.Records[1].Author.Username)
	assert.Equal(t, posts[1].User.Nickname, responsePostFeed.Records[1].Author.Nickname)
	assert.Equal(t, posts[1].User.ProfilePictureUrl, responsePostFeed.Records[1].Author.ProfilePictureUrl)
	assert.Equal(t, posts[1].Content, responsePostFeed.Records[1].Content)
	assert.True(t, posts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
	assert.Nil(t, responsePostFeed.Records[1].Location)

	assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
	assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
	assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPostsByHashtagUnauthorized tests if the GetPostsByHashtag function returns a 401 unauthorized if the user is not authenticated
func TestGetPostsByHashtagUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"",               // empty token
		"invalidToken",   // invalid token
		"Bearer invalid", // invalid token
	}
	for _, token := range invalidTokens {
		// Arrange
		mockPostRepository := new(repositories.MockPostRepository)

		postService := services.NewPostService(
			mockPostRepository,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		postController := controllers.NewPostController(postService)

		hashtag := "Post"
		limit := 2
		lastPost := models.Post{
			Id: uuid.New(),
		}

		// Setup HTTP request
		url := "/posts?q=" + hashtag + "&postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/posts", middleware.AuthorizeUser, postController.GetPostsByHashtag)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized

		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}
