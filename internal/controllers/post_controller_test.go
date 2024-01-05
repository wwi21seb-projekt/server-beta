package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/controllers"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/middleware"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/marcbudd/server-beta/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

// TestCreatePostSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully
func TestCreatePostSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		new(services.ImageService),
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
	var responsePost models.PostCreateResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePost)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, user.ProfilePictureUrl, responsePost.Author.ProfilePictureUrl)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Equal(t, content, capturedPost.Content)
	assert.Equal(t, user.Username, capturedPost.Username)
	assert.NotNil(t, capturedPost.CreatedAt)
	assert.NotNil(t, capturedPost.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
}

// TestCreatePostBadRequest tests if the CreatePost function returns a 400 bad request if the content is empty
func TestCreatePostBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,                       // invalid body
		`{"content": ""}`,                                 // empty content
		`{"content": "` + strings.Repeat("A", 300) + `"}`, // content too long
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

		expectedCustomError := customerrors.PreliminaryUserUnauthorized
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

	content := "This is a test post."
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content: content,
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

	expectedCustomError := customerrors.PreliminaryUserUnauthorized
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
		{"This is a test post.", "test.jpeg", "This is an image", "image/jpeg"},           // test jpeg
		{"This is a test post text.", "test.webp", "This is also an image", "image/webp"}, // test webp
		{"", "test.jpeg", "This is another image", "image/jpeg"},                          // test only image
	}

	for _, testCase := range testCases {
		// Create multipart request body
		content := testCase[0]
		testImageName := testCase[1]
		testImageContent := testCase[2]
		testImageContentType := testCase[3]

		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)
		mockHashtagRepository := new(repositories.MockHashtagRepository)
		mockFileSystem := new(repositories.MockFileSystem)

		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			mockHashtagRepository,
			services.NewImageService(mockFileSystem),
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
		var responsePost models.PostCreateResponseDTO
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
		assert.Equal(t, "api/images"+capturedFilename, capturedPost.ImageUrl)
		assert.Equal(t, capturedFile, []uint8(testImageContent))

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

	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		services.NewImageService(mockFileSystem),
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

	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		services.NewImageService(mockFileSystem),
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
	var responsePost models.PostCreateResponseDTO
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
