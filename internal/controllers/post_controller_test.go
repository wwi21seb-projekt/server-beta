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
	"net/http"
	"net/http/httptest"
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
	}
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

		nextPosts := []models.Post{
			{
				Id:       uuid.New(),
				Username: "someOtherUsername",
				User: models.User{
					Username:          "someOtherUsername",
					Nickname:          "someOtherNickname",
					ProfilePictureUrl: "",
				},
				Content:   "This is the next post",
				ImageUrl:  "",
				CreatedAt: time.Now().Add(time.Hour * -2),
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

		assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
		assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
		assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
		assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))

		assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
		assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
		assert.Equal(t, nextPosts[1].Id, responsePostFeed.Pagination.LastPostId)

		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetPostFeedBadRequest tests if the GetPostFeed function returns a 400 bad request if the limit is invalid or feedType is invalid
func TestGetPostFeedBadRequest(t *testing.T) {
	urls := []string{
		"/feed?postId=" + uuid.New().String() + "&limit=invalidLimit",                // invalid limit
		"/feed?postId=" + uuid.New().String() + "&limit=10&feedType=invalidFeedType", // invalid feedType
	}
	for _, url := range urls {
		// Arrange
		mockUserRepository := new(repositories.MockUserRepository)
		mockPostRepository := new(repositories.MockPostRepository)

		postService := services.NewPostService(
			mockPostRepository,
			mockUserRepository,
			nil,
		)
		postController := controllers.NewPostController(postService)

		// Setup HTTP request
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/feed", postController.GetPostFeed)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 Bad Request
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetGlobalPostFeedPostNotFound tests if the GetPostFeed function returns a 404 not found if the last post does not exist
func TestGetGlobalPostFeedPostNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
	)
	postController := controllers.NewPostController(postService)

	// Mock expectations
	mockPostRepository.On("GetPostById", mock.AnythingOfType("string")).Return(models.Post{}, gorm.ErrRecordNotFound) // Post not found

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/feed?postId="+uuid.New().String()+"&limit=10", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", postController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 not found
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PreliminaryPostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

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

	nextPosts := []models.Post{
		{
			Id:       uuid.New(),
			Username: "someOtherUsername",
			User: models.User{
				Username:          "someOtherUsername",
				Nickname:          "someOtherNickname",
				ProfilePictureUrl: "",
			},
			Content:   "This is the next post",
			ImageUrl:  "",
			CreatedAt: time.Now().Add(time.Hour * -2),
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

	assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
	assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
	assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
	assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))

	assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
	assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
	assert.Equal(t, nextPosts[1].Id, responsePostFeed.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
}

// TestGetPersonalPostFeedPostNotFound tests if the GetPostFeed function returns a 404 not found if the last post does not exist
func TestGetPersonalPostFeedPostNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	postService := services.NewPostService(
		mockPostRepository,
		mockUserRepository,
		nil,
	)
	postController := controllers.NewPostController(postService)

	username := "thisUser"
	token, err := utils.GenerateAccessToken(username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockPostRepository.On("GetPostById", mock.AnythingOfType("string")).Return(models.Post{}, gorm.ErrRecordNotFound) // Post not found

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/feed?postId="+uuid.New().String()+"&limit=10&feedType=personal", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", postController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 not found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PreliminaryPostNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

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

		expectedCustomError := customerrors.PreliminaryUserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
	}
}
