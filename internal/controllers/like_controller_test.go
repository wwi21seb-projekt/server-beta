package controllers_test

import (
	"encoding/json"
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
	"testing"
)

// TestPostLikeSuccess tests if PostLike creates a like for a given post id and the current logged-in user and returns 204 No Content
func TestPostLikeSuccess(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	// Mock expectations
	var capturedLike *models.Like
	mockPostRepo.On("GetPostById", post.Id.String()).Return(post, nil)                                            // post exists
	mockLikeRepo.On("FindLike", post.Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound) // user did not like yet
	mockLikeRepo.On("CreateLike", mock.AnythingOfType("*models.Like")).
		Run(func(args mock.Arguments) {
			capturedLike = args.Get(0).(*models.Like)
		}).Return(nil) // post created successfully

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/likes"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/likes", middleware.AuthorizeUser, likeController.PostLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect 204 No Content

	assert.Equal(t, "", w.Body.String())

	assert.NotNil(t, capturedLike)
	assert.NotEmpty(t, capturedLike.Id)
	assert.Equal(t, post.Id, capturedLike.PostId)
	assert.Equal(t, currentUsername, capturedLike.Username)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestPostLikeUnauthorized tests if PostLike returns 401 Unauthorized if the user is not logged in
func TestPostLikeUnauthorized(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	// Setup HTTP request
	url := "/posts/1/likes"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/likes", middleware.AuthorizeUser, likeController.PostLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestPostLikePostNotFound tests if PostLike returns 404 Not Found if the post does not exist
func TestPostLikePostNotFound(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	postId := uuid.New().String()

	// Mock expectations
	mockPostRepo.On("GetPostById", postId).Return(models.Post{}, gorm.ErrRecordNotFound) // post does not exist

	// Setup HTTP request
	url := "/posts/" + postId + "/likes"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/likes", middleware.AuthorizeUser, likeController.PostLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestPostLikeAlreadyLiked tests if PostLike returns 409 Conflict if the user already liked the post
func TestPostLikeAlreadyLiked(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	// Mock expectations
	mockPostRepo.On("GetPostById", post.Id.String()).Return(post, nil)                         // post exists
	mockLikeRepo.On("FindLike", post.Id.String(), currentUsername).Return(&models.Like{}, nil) // user already liked

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/likes"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts/:postId/likes", middleware.AuthorizeUser, likeController.PostLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code) // Expect 409 Conflict
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.AlreadyLiked
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestDeleteLikeSuccess tests if DeleteLike deletes a like for a given post id and the current logged-in user and returns 204 No Content
func TestDeleteLikeSuccess(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	like := models.Like{
		Id: uuid.New(),
	}

	// Mock expectations
	mockPostRepo.On("GetPostById", post.Id.String()).Return(post, nil)                // post exists
	mockLikeRepo.On("FindLike", post.Id.String(), currentUsername).Return(&like, nil) // user liked
	mockLikeRepo.On("DeleteLike", like.Id.String()).Return(nil)                       // delete successful

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/likes"
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId/likes", middleware.AuthorizeUser, likeController.DeleteLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect 204 No Content

	assert.Equal(t, "", w.Body.String())

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestDeleteLikeUnauthorized tests if DeleteLike returns 401 Unauthorized if the user is not logged in
func TestDeleteLikeUnauthorized(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	// Setup HTTP request
	url := "/posts/1/likes"
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId/likes", middleware.AuthorizeUser, likeController.DeleteLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestDeleteLikePostNotFound tests if DeleteLike returns 404 Not Found if the post does not exist
func TestDeleteLikePostNotFound(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	postId := uuid.New().String()

	// Mock expectations
	mockPostRepo.On("GetPostById", postId).Return(models.Post{}, gorm.ErrRecordNotFound) // post does not exist

	// Setup HTTP request
	url := "/posts/" + postId + "/likes"
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId/likes", middleware.AuthorizeUser, likeController.DeleteLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

// TestDeleteLikeNotLiked tests if DeleteLike returns 409 Conflict if the user did not like the post
func TestDeleteLikeNotLiked(t *testing.T) {
	// Arrange
	mockPostRepo := new(repositories.MockPostRepository)
	mockLikeRepo := new(repositories.MockLikeRepository)

	likeService := services.NewLikeService(mockLikeRepo, mockPostRepo)
	likeController := controllers.NewLikeController(likeService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	post := models.Post{
		Id: uuid.New(),
	}

	// Mock expectations
	mockPostRepo.On("GetPostById", post.Id.String()).Return(post, nil)                                            // post exists
	mockLikeRepo.On("FindLike", post.Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound) // user did not like

	// Setup HTTP request
	url := "/posts/" + post.Id.String() + "/likes"
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId/likes", middleware.AuthorizeUser, likeController.DeleteLike)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code) // Expect 409 Conflict
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.NotLiked
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockLikeRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}
