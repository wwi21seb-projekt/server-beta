package controllers_test

import (
	"bytes"
	"encoding/base64"
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
	"os"
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
	validator := new(utils.Validator)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
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
	assert.Nil(t, capturedPost.RepostId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Empty(t, capturedPost.ImageId)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Nil(t, responsePost.Author.Picture)

	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Location)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))

	assert.Nil(t, responsePost.Location)
	assert.Nil(t, responsePost.Repost)
	assert.Nil(t, responsePost.Picture)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
}

// TestCreatePostWithLocationSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully with location
func TestCreatePostWithLocationSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
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
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, capturedPost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, capturedPost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, capturedPost.Location.Accuracy)
	assert.Equal(t, capturedPost.Location.Id, *capturedPost.LocationId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Nil(t, capturedPost.ImageId)
	assert.Nil(t, capturedPost.RepostId)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Nil(t, responsePost.Author.Picture)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.NotNil(t, responsePost.Location)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, *responsePost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, *responsePost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, *responsePost.Location.Accuracy)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Repost)
	assert.Nil(t, responsePost.Picture)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
}

// Regression Test
// TestCreatePostWithLocationZeroValues tests if a post is created successfully with location where accuracy, latitude and longitude are all zero
func TestCreatePostWithLocationZeroValues(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username:     "testUser",
		Nickname:     "testNickname",
		Email:        "test@domain.com",
		PasswordHash: "passwordHash",
		CreatedAt:    time.Now().UTC().Add(time.Hour * -24),
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
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, capturedPost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, capturedPost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, capturedPost.Location.Accuracy)
	assert.NotNil(t, *capturedPost.LocationId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.Nil(t, capturedPost.ImageId)
	assert.Nil(t, capturedPost.RepostId)
	assert.Equal(t, capturedPost.Hashtags[0].Id, expectedHashtagOne.Id)
	assert.Equal(t, capturedPost.Hashtags[0].Name, expectedHashtagOne.Name)
	assert.Equal(t, capturedPost.Hashtags[1].Id, expectedHashtagTwo.Id)
	assert.Equal(t, capturedPost.Hashtags[1].Name, expectedHashtagTwo.Name)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Nil(t, responsePost.Author.Picture)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.NotNil(t, responsePost.Location)
	assert.Equal(t, *postCreateRequestDTO.Location.Longitude, *responsePost.Location.Longitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Latitude, *responsePost.Location.Latitude)
	assert.Equal(t, *postCreateRequestDTO.Location.Accuracy, *responsePost.Location.Accuracy)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Repost)
	assert.Nil(t, responsePost.Picture)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
}

// TestCreatePostWithImage tests if the CreatePost function returns a postDto and 201 created if post is created successfully with image
func TestCreatePostWithImage(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	validator := new(utils.Validator)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
		validator,
		nil,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	err := os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	profileImage := models.Image{
		Id:     uuid.New(),
		Format: ".jpeg",
		Height: 100,
		Width:  101,
		Tag:    time.Now().UTC(),
	}
	expectedProfileImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + profileImage.Id.String() + "." + profileImage.Format

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
		ImageId:  &profileImage.Id,
		Image:    profileImage,
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Create request dto
	imageData, err := os.ReadFile("../../tests/resources/valid.jpeg")
	if err != nil {
		t.Fatal(err)
	}
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Picture: imageBase64,
	}

	// Mock expectations
	var capturedPost *models.Post
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully

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
	assert.Equal(t, "", capturedPost.Content)
	assert.Nil(t, capturedPost.LocationId)
	assert.Nil(t, capturedPost.RepostId)
	assert.NotEmpty(t, capturedPost.CreatedAt)
	assert.NotEmpty(t, capturedPost.ImageId.String())
	assert.Equal(t, 444, capturedPost.Image.Height)
	assert.Equal(t, 670, capturedPost.Image.Width)
	assert.Equal(t, imageData, capturedPost.Image.ImageData)
	assert.Equal(t, "jpeg", capturedPost.Image.Format)
	assert.NotEmpty(t, capturedPost.Image.Tag)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.NotNil(t, responsePost.Author.Picture)
	assert.Equal(t, expectedProfileImageUrl, responsePost.Author.Picture.Url)
	assert.Equal(t, profileImage.Width, responsePost.Author.Picture.Width)
	assert.Equal(t, profileImage.Height, responsePost.Author.Picture.Height)
	assert.Equal(t, profileImage.Tag, responsePost.Author.Picture.Tag)

	assert.Empty(t, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Location)
	assert.Nil(t, responsePost.Repost)
	assert.NotNil(t, responsePost.Picture)

	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
}

// TestCreatePostWithRepostSuccess tests if the CreatePost function returns a postDto and 201 created if post is created successfully with repost
func TestCreatePostWithRepostSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLikeRepository := new(repositories.MockLikeRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)

	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		mockLikeRepository,
		mockCommentRepository,
		notificationService,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
	}
	originalUser := models.User{
		Username: "originalUser",
	}
	originalPost := models.Post{
		Id:         uuid.New(),
		Username:   originalUser.Username,
		User:       originalUser,
		Content:    "This is the original post.",
		CreatedAt:  time.Now().UTC().Add(time.Hour * -24),
		LocationId: nil,
		RepostId:   nil,
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	content := "This is a test."

	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content:        content,
		RepostedPostId: originalPost.Id.String(),
	}

	totalCommentsCount := int64(5)
	totalLikesCount := int64(10)

	// Mock expectations
	var capturedPost *models.Post
	var capturedNotification *models.Notification
	mockUserRepository.On("FindUserByUsername", user.Username).Return(&user, nil) // User found successfully
	mockPostRepository.On("CreatePost", mock.AnythingOfType("*models.Post")).
		Run(func(args mock.Arguments) {
			capturedPost = args.Get(0).(*models.Post) // Save argument to captor
		}).Return(nil) // Post created successfully
	mockPostRepository.On("GetPostById", originalPost.Id.String()).Return(originalPost, nil)
	mockLikeRepository.On("CountLikes", originalPost.Id.String()).Return(totalLikesCount, nil)
	mockLikeRepository.On("FindLike", originalPost.Id.String(), user.Username).Return(&models.Like{}, gorm.ErrRecordNotFound)
	mockCommentRepository.On("CountComments", originalPost.Id.String()).Return(totalCommentsCount, nil)
	mockNotificationRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).
		Run(func(args mock.Arguments) {
			capturedNotification = args.Get(0).(*models.Notification) // Save argument to captor
		}).Return(nil)
	mockNotificationRepo.On("GetNotificationById", mock.AnythingOfType("string")).Return(models.Notification{}, nil)
	mockPushSubscriptionRepo.On("GetPushSubscriptionsByUsername", originalPost.Username).Return([]models.PushSubscription{}, nil)

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
	assert.Empty(t, capturedPost.ImageId)
	assert.Equal(t, capturedPost.RepostId, &originalPost.Id)

	assert.Equal(t, user.Username, responsePost.Author.Username)
	assert.Equal(t, user.Nickname, responsePost.Author.Nickname)
	assert.Equal(t, content, responsePost.Content)
	assert.Equal(t, capturedPost.Id, responsePost.PostId)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))
	assert.Nil(t, responsePost.Location)
	assert.True(t, capturedPost.CreatedAt.Equal(responsePost.CreationDate))

	assert.Equal(t, originalPost.Id, responsePost.Repost.PostId)
	assert.Equal(t, originalPost.Content, responsePost.Repost.Content)
	assert.True(t, originalPost.CreatedAt.Equal(responsePost.Repost.CreationDate))
	assert.Equal(t, originalPost.Username, responsePost.Repost.Author.Username)
	assert.Equal(t, originalUser.Nickname, responsePost.Repost.Author.Nickname)
	assert.Equal(t, totalCommentsCount, responsePost.Repost.Comments)
	assert.Equal(t, totalLikesCount, responsePost.Repost.Likes)
	assert.False(t, responsePost.Repost.Liked)
	assert.Nil(t, responsePost.Repost.Location)
	assert.Nil(t, responsePost.Repost.Repost)

	assert.Equal(t, user.Username, capturedNotification.FromUsername)
	assert.Equal(t, originalPost.Username, capturedNotification.ForUsername)
	assert.Equal(t, "repost", capturedNotification.NotificationType)
	assert.NotNil(t, capturedNotification.Timestamp)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockLikeRepository.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockCommentRepository.AssertExpectations(t)
}

// TestCreatePostRepostNotFound tests if the CreatePost function returns a 404 not found if the original post is not found
func TestCreatePostRepostNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLikeRepository := new(repositories.MockLikeRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		mockLikeRepository,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	content := "This is a test."

	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content:        content,
		RepostedPostId: uuid.New().String(),
	}

	// Mock expectations
	mockPostRepository.On("GetPostById", postCreateRequestDTO.RepostedPostId).Return(models.Post{}, gorm.ErrRecordNotFound) // Post not found

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
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 not found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockHashtagRepository.AssertExpectations(t)
	mockLikeRepository.AssertExpectations(t)
}

// TestCreatePostRepostOfRepost tests if the CreatePost function returns a 400 bad request if the post is a repost of a repost
func TestCreatePostRepostOfRepost(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockHashtagRepository := new(repositories.MockHashtagRepository)
	validator := new(utils.Validator)
	mockLikeRepository := new(repositories.MockLikeRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		mockHashtagRepository,
		validator,
		mockLikeRepository,
		nil,
		nil,
	)
	postController := controllers.NewPostController(postService)

	user := models.User{
		Username: "testUser",
	}

	tempId := uuid.New()
	originalPost := models.Post{
		Id:         uuid.New(),
		Username:   "originalUser",
		Content:    "This is the original post.",
		CreatedAt:  time.Now().UTC().Add(time.Hour * -24),
		LocationId: nil,
		RepostId:   &tempId,
	}

	authenticationToken, err := utils.GenerateAccessToken(user.Username)
	if err != nil {
		t.Fatal(err)
	}

	content := "This is a test."

	originalPostIdString := originalPost.Id.String()
	postCreateRequestDTO := models.PostCreateRequestDTO{
		Content:        content,
		RepostedPostId: originalPostIdString,
	}

	// Mock expectations
	mockPostRepository.On("GetPostById", originalPostIdString).Return(originalPost, nil) // Return original post

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
	mockLikeRepository.AssertExpectations(t)

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
			nil,
			nil,
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
			nil,
			nil,
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

		expectedCustomError := customerrors.Unauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
		mockHashtagRepository.AssertExpectations(t)
	}
}

// TestDeletePostSuccess tests if the DeletePostSuccess function returns a post and 204 no content if the request is valid
func TestDeletePostSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil, nil)
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
		nil,
	)
	postController := controllers.NewPostController(postService)

	postId := uuid.New()

	// Setup HTTP request without Authorization Header
	req, _ := http.NewRequest("DELETE", "/posts/"+postId.String(), nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.Unauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPostRepository.AssertExpectations(t)
}

// TestDeletePostForbidden checks for forbidden access when a user tries to delete others' posts.
func TestDeletePostForbidden(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil, nil)
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

	postService := services.NewPostService(mockPostRepository, nil, nil, nil, nil, nil, nil)
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

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPostRepository.AssertExpectations(t)
}
