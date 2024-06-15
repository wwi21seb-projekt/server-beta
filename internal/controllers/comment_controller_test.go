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
	"os"
	"testing"
	"time"
)

// TestCreateCommentSuccess tests the CreateComment function if it returns 200 OK after successfully creating a comment
func TestCreateCommentSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	testUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(testUsername)
	if err != nil {
		t.Fatal(err)
	}

	imageId := uuid.New()
	imageFormat := "png"
	err = os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	expectedImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + imageId.String() + "." + imageFormat

	user := models.User{
		Username: testUsername,
		Nickname: "test user",
		ImageId:  &imageId,
		Image: models.Image{
			Id:     imageId,
			Format: imageFormat,
			Width:  100,
			Height: 100,
			Tag:    time.Now().UTC(),
		},
	}

	post := models.Post{
		Id: uuid.New(),
	}
	commentCreateRequest := models.CommentCreateRequestDTO{
		Content: "Test comment",
	}

	// Mock expectations
	var capturedComment *models.Comment
	mockPostRepository.On("GetPostById", post.Id.String()).Return(post, nil)
	mockCommentRepository.On("CreateComment", mock.AnythingOfType("*models.Comment")).
		Run(func(args mock.Arguments) {
			capturedComment = args.Get(0).(*models.Comment)
		}).Return(nil)
	mockUserRepository.On("FindUserByUsername", testUsername).Return(&user, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(commentCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	url := "/posts/" + post.Id.String() + "/comments"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts/:postId/comments", middleware.AuthorizeUser, commentController.CreateComment)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 Created

	var responseComment models.CommentResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseComment)
	assert.NoError(t, err)

	assert.NotEmpty(t, capturedComment.Id)
	assert.Equal(t, post.Id, capturedComment.PostID)
	assert.Equal(t, testUsername, capturedComment.Username)
	assert.Equal(t, commentCreateRequest.Content, capturedComment.Content)
	assert.NotNil(t, capturedComment.CreatedAt)

	assert.Equal(t, capturedComment.Id, responseComment.CommentId)
	assert.Equal(t, commentCreateRequest.Content, responseComment.Content)
	assert.True(t, capturedComment.CreatedAt.Equal(responseComment.CreationDate))
	assert.Equal(t, user.Username, responseComment.Author.Username)
	assert.Equal(t, user.Nickname, responseComment.Author.Nickname)

	assert.NotNil(t, responseComment.Author.Picture)
	assert.Equal(t, expectedImageUrl, responseComment.Author.Picture.Url)
	assert.Equal(t, user.Image.Width, responseComment.Author.Picture.Width)
	assert.Equal(t, user.Image.Height, responseComment.Author.Picture.Height)
	assert.Equal(t, user.Image.Tag, responseComment.Author.Picture.Tag)

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestCreateCommentBadRequest tests the CreateComment function if it returns 400 Bad Request when the request body is invalid
func TestCreateCommentBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{}`,                          // Empty body
		`{"invalid": "Test comment"}`, // invalid field
		`{"content": ""}`,             // Empty content
		`{"content": "` + string(make([]rune, 129)) + `"}`, // Content exceeds 128 characters
	}

	for _, body := range invalidBodies {
		// Arrange
		mockPostRepository := new(repositories.MockPostRepository)
		mockCommentRepository := new(repositories.MockCommentRepository)
		mockUserRepository := new(repositories.MockUserRepository)

		commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
		commentController := controllers.NewCommentController(commentService)

		testUsername := "testUser"
		authenticationToken, err := utils.GenerateAccessToken(testUsername)
		if err != nil {
			t.Fatal(err)
		}

		post := models.Post{
			Id: uuid.New(),
		}

		// Setup HTTP request
		url := "/posts/" + post.Id.String() + "/comments"
		req, _ := http.NewRequest("POST", url, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/posts/:postId/comments", middleware.AuthorizeUser, commentController.CreateComment)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 Bad Request
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockCommentRepository.AssertExpectations(t)
		mockPostRepository.AssertExpectations(t)
		mockUserRepository.AssertExpectations(t)
	}
}

// TestCreateCommentUnauthorized tests the CreateComment function if it returns 401 Unauthorized when the user is not authenticated
func TestCreateCommentUnauthorized(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	post := models.Post{
		Id: uuid.New(),
	}

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/comments"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts/:postId/comments", middleware.AuthorizeUser, commentController.CreateComment)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestCreateCommentPostNotFound tests the CreateComment function if it returns 404 Not Found when the post does not exist
func TestCreateCommentPostNotFound(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	testUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(testUsername)
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	// Mock expectations
	mockPostRepository.On("GetPostById", post.Id.String()).Return(models.Post{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	commentCreateRequest := models.CommentCreateRequestDTO{
		Content: "Test comment",
	}
	requestBody, err := json.Marshal(commentCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	url := "/posts/" + post.Id.String() + "/comments"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/posts/:postId/comments", middleware.AuthorizeUser, commentController.CreateComment)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestGetCommentsByPostIdSuccess tests the GetCommentsByPostId function if it returns 200 OK after successfully retrieving comments
func TestGetCommentsByPostIdSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	authenticationToken, err := utils.GenerateAccessToken("myUser")
	if err != nil {
		t.Fatal(err)
	}

	imageId := uuid.New()
	imageFormat := "png"
	err = os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	expectedImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + imageId.String() + "." + imageFormat

	post := models.Post{
		Id: uuid.New(),
	}
	comments := []models.Comment{
		{
			Id:       uuid.New(),
			PostID:   post.Id,
			Username: "testUser",
			User: models.User{
				Username: "testUser",
				Nickname: "test user",
				ImageId:  &imageId,
				Image: models.Image{
					Id:     imageId,
					Format: imageFormat,
					Width:  100,
					Height: 100,
					Tag:    time.Now().UTC(),
				},
			},
			Content:   "Test comment 1",
			CreatedAt: time.Now().UTC(),
		},
		{
			Id:       uuid.New(),
			PostID:   post.Id,
			Username: "testUser2",
			User: models.User{
				Username: "testUser2",
				Nickname: "test user 2",
			},
			Content:   "Test comment 2",
			CreatedAt: time.Now().UTC(),
		},
	}

	totalNumberOfComments := 7
	offset := 4
	limit := 2

	// Mock expectations
	mockPostRepository.On("GetPostById", post.Id.String()).Return(models.Post{}, nil)
	mockCommentRepository.On("GetCommentsByPostId", post.Id.String(), offset, limit).Return(comments, int64(totalNumberOfComments), nil)

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/comments?offset=" + fmt.Sprint(offset) + "&limit=" + fmt.Sprint(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/comments", middleware.AuthorizeUser, commentController.GetCommentsByPostId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var responseList models.CommentFeedResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseList)
	assert.NoError(t, err)

	assert.Len(t, responseList.Records, 2)
	assert.Equal(t, int64(totalNumberOfComments), responseList.Pagination.Records)
	assert.Equal(t, offset, responseList.Pagination.Offset)
	assert.Equal(t, limit, responseList.Pagination.Limit)

	for i, comment := range comments {
		fmt.Printf("Expected Comment User: %+v\n", comment.User)
		fmt.Printf("Actual Comment User: %+v\n", responseList.Records[i].Author.Picture)
		fmt.Printf("Actual Comment User: %+v\n", i)
		assert.Equal(t, comment.Id, responseList.Records[i].CommentId)
		assert.Equal(t, comment.Content, responseList.Records[i].Content)
		assert.True(t, comment.CreatedAt.Equal(responseList.Records[i].CreationDate))
		assert.Equal(t, comment.User.Username, responseList.Records[i].Author.Username)
		assert.Equal(t, comment.User.Nickname, responseList.Records[i].Author.Nickname)

		if comment.User.ImageId != nil {
			assert.NotNil(t, responseList.Records[i].Author.Picture)
			assert.Equal(t, expectedImageUrl, responseList.Records[i].Author.Picture.Url)
			assert.Equal(t, comment.User.Image.Width, responseList.Records[i].Author.Picture.Width)
			assert.Equal(t, comment.User.Image.Height, responseList.Records[i].Author.Picture.Height)
			assert.Equal(t, comment.User.Image.Tag, responseList.Records[i].Author.Picture.Tag)
		} else {
			assert.Nil(t, responseList.Records[i].Author.Picture)
		}
	}

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestGetCommentsByPostIdUnauthorized tests the GetCommentsByPostId function if it returns 401 Unauthorized when the user is not authenticated
func TestGetCommentsByPostIdUnauthorized(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	post := models.Post{
		Id: uuid.New(),
	}

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/comments"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/comments", middleware.AuthorizeUser, commentController.GetCommentsByPostId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestGetCommentsByPostIdPostNotFound tests the GetCommentsByPostId function if it returns 404 Not Found when the post does not exist
func TestGetCommentsByPostIdPostNotFound(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)
	mockUserRepository := new(repositories.MockUserRepository)

	commentService := services.NewCommentService(mockCommentRepository, mockPostRepository, mockUserRepository)
	commentController := controllers.NewCommentController(commentService)

	authenticationToken, err := utils.GenerateAccessToken("myUser")
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	// Mock expectations
	mockPostRepository.On("GetPostById", post.Id.String()).Return(models.Post{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/comments"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/comments", middleware.AuthorizeUser, commentController.GetCommentsByPostId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockCommentRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}
