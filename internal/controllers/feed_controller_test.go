package controllers_test

import (
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

// TestGetPostsByUsernameSuccess tests if the GetPostsByUsername function returns a list of posts and 200 ok if the user exists
func TestGetPostsByUsernameSuccess(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)
	mockLikeRepository := new(repositories.MockLikeRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		mockUserRepository,
		mockLikeRepository,
		mockCommentRepository,
	)
	feedController := controllers.NewFeedController(feedService)

	err := os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	userImage := models.Image{
		Id:        uuid.New(),
		Format:    "png",
		Width:     100,
		Height:    200,
		Tag:       time.Now().UTC(),
		ImageData: []byte("some image data"),
	}
	//expectedUserImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + userImage.Id.String() + "." + userImage.Format

	postImage := models.Image{
		Id:        uuid.New(),
		Format:    "jpeg",
		Width:     300,
		Height:    400,
		Tag:       time.Now().UTC(),
		ImageData: []byte("some image other data"),
	}
	expectedPostImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + postImage.Id.String() + "." + postImage.Format

	originalPostImage := models.Image{
		Id:        uuid.New(),
		Format:    "webp",
		Width:     500,
		Height:    600,
		ImageData: []byte("some image other data 2"),
	}
	expectedOriginalPostImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + originalPostImage.Id.String() + "." + originalPostImage.Format

	user := models.User{
		Username: "testUser",
		Nickname: "testNickname",
		Email:    "test@example.com",
		ImageId:  &userImage.Id,
		Image:    userImage,
	}

	originalPostLocation := models.Location{
		Id:        uuid.New(),
		Longitude: 4,
		Latitude:  5,
	}
	originalPost := models.Post{
		Id:         uuid.New(),
		Username:   user.Username,
		User:       user,
		Content:    "This is a original post",
		CreatedAt:  time.Now().UTC().Add(time.Hour * -1),
		ImageId:    &originalPostImage.Id,
		Image:      originalPostImage,
		LocationId: &originalPostLocation.Id,
		Location:   originalPostLocation,
	}

	locationId := uuid.New()
	posts := []models.Post{
		{
			Id:         uuid.New(), // first post has image, location and is a repost
			Username:   user.Username,
			Content:    "Test Post 1",
			CreatedAt:  time.Now().UTC(),
			LocationId: &locationId,
			Location: models.Location{
				Longitude: 11.1,
				Latitude:  22.2,
				Accuracy:  50,
			},
			ImageId:  &postImage.Id,
			Image:    postImage,
			RepostId: &originalPost.Id,
		},
		{
			Id:         uuid.New(),
			Username:   user.Username,
			Content:    "Test Post 2",
			CreatedAt:  time.Now().UTC().Add(-1 * time.Hour),
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

	mockPostRepository.On("GetPostById", originalPost.Id.String()).Return(originalPost, nil)

	firstPostLikes := int64(0)
	secondPostLikes := int64(10)
	originalPostLikes := int64(20)
	mockLikeRepository.On("CountLikes", posts[0].Id.String()).Return(firstPostLikes, nil)
	mockLikeRepository.On("CountLikes", posts[1].Id.String()).Return(secondPostLikes, nil)
	mockLikeRepository.On("CountLikes", originalPost.Id.String()).Return(originalPostLikes, nil)

	mockLikeRepository.On("FindLike", posts[0].Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound)     // First post not liked by current user
	mockLikeRepository.On("FindLike", posts[1].Id.String(), currentUsername).Return(&models.Like{}, nil)                        // Second post liked by current user
	mockLikeRepository.On("FindLike", originalPost.Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound) // Original post not liked by current user

	firstPostComments := int64(25)
	secondPostComments := int64(5)
	originalPostComments := int64(15)
	mockCommentRepository.On("CountComments", posts[0].Id.String()).Return(firstPostComments, nil)
	mockCommentRepository.On("CountComments", posts[1].Id.String()).Return(secondPostComments, nil)
	mockCommentRepository.On("CountComments", originalPost.Id.String()).Return(originalPostComments, nil)

	// Setup HTTP request
	url := "/users/" + user.Username + "/feed?offset=" + fmt.Sprint(offset) + "&limit=" + fmt.Sprint(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/users/:username/feed", middleware.AuthorizeUser, feedController.GetPostsByUserUsername)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert first post in response body
	assert.Equal(t, posts[0].Id.String(), response.Records[0].PostId)
	assert.Equal(t, posts[0].Content, response.Records[0].Content)
	assert.True(t, posts[0].CreatedAt.Equal(response.Records[0].CreationDate))
	assert.Equal(t, firstPostComments, response.Records[0].Comments)
	assert.Equal(t, firstPostLikes, response.Records[0].Likes)
	assert.Equal(t, false, response.Records[0].Liked)
	assert.NotNil(t, response.Records[0].Location)
	assert.Equal(t, posts[0].Location.Latitude, *response.Records[0].Location.Latitude)
	assert.Equal(t, posts[0].Location.Longitude, *response.Records[0].Location.Longitude)
	assert.Equal(t, posts[0].Location.Accuracy, *response.Records[0].Location.Accuracy)
	assert.NotNil(t, response.Records[0].Picture)
	assert.Equal(t, expectedPostImageUrl, response.Records[0].Picture.Url)
	assert.Equal(t, posts[0].Image.Width, response.Records[0].Picture.Width)
	assert.Equal(t, posts[0].Image.Height, response.Records[0].Picture.Height)
	assert.Equal(t, posts[0].Image.Tag, response.Records[0].Picture.Tag)

	// Assert original post of first post in response body
	assert.NotNil(t, response.Records[0].Repost)
	repost := response.Records[0].Repost
	assert.Equal(t, originalPost.Id.String(), repost.PostId.String())
	assert.Equal(t, originalPost.Content, repost.Content)
	assert.True(t, originalPost.CreatedAt.Equal(repost.CreationDate))
	assert.Equal(t, originalPostComments, repost.Comments)
	assert.Equal(t, originalPostLikes, repost.Likes)
	assert.Equal(t, false, repost.Liked)
	assert.NotNil(t, repost.Location)
	assert.Equal(t, originalPostLocation.Latitude, *repost.Location.Latitude)
	assert.Equal(t, originalPostLocation.Longitude, *repost.Location.Longitude)
	assert.Equal(t, originalPostLocation.Accuracy, *repost.Location.Accuracy)
	assert.NotNil(t, repost.Picture)
	assert.Equal(t, originalPostImage.Width, repost.Picture.Width)
	assert.Equal(t, originalPostImage.Height, repost.Picture.Height)
	assert.Equal(t, originalPostImage.Tag, repost.Picture.Tag)
	assert.Equal(t, expectedOriginalPostImageUrl, repost.Picture.Url)

	// Assert second post in response body
	assert.Equal(t, posts[1].Id.String(), response.Records[1].PostId)
	assert.Equal(t, posts[1].Content, response.Records[1].Content)
	assert.True(t, posts[1].CreatedAt.Equal(response.Records[1].CreationDate))
	assert.Equal(t, secondPostComments, response.Records[1].Comments)
	assert.Equal(t, secondPostLikes, response.Records[1].Likes)
	assert.Equal(t, true, response.Records[1].Liked)
	assert.Nil(t, response.Records[1].Location)
	assert.Nil(t, response.Records[1].Picture)

	assert.Equal(t, offset, response.Pagination.Offset)
	assert.Equal(t, limit, response.Pagination.Limit)
	assert.Equal(t, int64(len(posts)), response.Pagination.Records)

	// Validate Mock Expectations
	mockUserRepository.AssertExpectations(t)
	mockPostRepository.AssertExpectations(t)
	mockLikeRepository.AssertExpectations(t)
	mockCommentRepository.AssertExpectations(t)
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

		feedService := services.NewFeedService(
			mockPostRepository,
			mockUserRepository,
			nil,
			nil,
		)
		feedController := controllers.NewFeedController(feedService)

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
		router.GET("/users/:username/feed", middleware.AuthorizeUser, feedController.GetPostsByUserUsername)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.Unauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestFindPostsByUserUsernameNotFound tests if the GetPostsByUserUsername function returns a 404 not found if the user does not exist
func TestGetPostsByUsernameUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
	)
	feedController := controllers.NewFeedController(feedService)

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
	router.GET("/users/:username/feed", middleware.AuthorizeUser, feedController.GetPostsByUserUsername)
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
		mockLikeRepository := new(repositories.MockLikeRepository)
		mockCommentRepository := new(repositories.MockCommentRepository)

		feedService := services.NewFeedService(
			mockPostRepository,
			mockUserRepository,
			mockLikeRepository,
			mockCommentRepository,
		)
		feedController := controllers.NewFeedController(feedService)

		lastPost := models.Post{
			Id:        uuid.New(),
			Username:  "someUserTest",
			User:      models.User{},
			Content:   "This is the last post",
			CreatedAt: time.Now().UTC().Add(time.Hour * -1),
		}

		locationId := uuid.New()
		nextPosts := []models.Post{
			{
				Id:       uuid.New(),
				Username: "someOtherUsername",
				User: models.User{
					Username: "someOtherUsername",
					Nickname: "someOtherNickname",
				},
				Content:    "This is the next post",
				CreatedAt:  time.Now().UTC().Add(time.Hour * -2),
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
					Username: "anotherTestUsername",
					Nickname: "anotherTestNickname",
				},
				Content:   "This is another next post",
				CreatedAt: time.Now().UTC().Add(time.Hour * -3),
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

		firstPostLikes := int64(0)
		secondPostLikes := int64(10)
		mockLikeRepository.On("CountLikes", nextPosts[0].Id.String()).Return(firstPostLikes, nil)
		mockLikeRepository.On("CountLikes", nextPosts[1].Id.String()).Return(secondPostLikes, nil)

		firstCommentCount := int64(0)
		secondCommentCount := int64(5)
		mockCommentRepository.On("CountComments", nextPosts[0].Id.String()).Return(firstCommentCount, nil)
		mockCommentRepository.On("CountComments", nextPosts[1].Id.String()).Return(secondCommentCount, nil)

		mockLikeRepository.On("FindLike", nextPosts[0].Id.String(), mock.AnythingOfType("string")).Return(&models.Like{}, gorm.ErrRecordNotFound) // First post not liked by current user
		mockLikeRepository.On("FindLike", nextPosts[1].Id.String(), mock.AnythingOfType("string")).Return(&models.Like{}, gorm.ErrRecordNotFound) // Second post not liked by current user

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
		router.GET("/feed", feedController.GetPostFeed)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok
		var responsePostFeed models.GeneralFeedDTO
		err := json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
		assert.NoError(t, err)

		assert.Equal(t, lastPost.Id, capturedLastPost.Id)
		assert.Equal(t, lastPost.Username, capturedLastPost.Username)
		assert.Equal(t, lastPost.Content, capturedLastPost.Content)
		assert.Equal(t, lastPost.ImageId, capturedLastPost.ImageId)
		assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
		assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

		assert.Equal(t, nextPosts[0].Id, responsePostFeed.Records[0].PostId)
		assert.Equal(t, nextPosts[0].Username, responsePostFeed.Records[0].Author.Username)
		assert.Equal(t, nextPosts[0].Content, responsePostFeed.Records[0].Content)
		assert.True(t, nextPosts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
		assert.NotNil(t, responsePostFeed.Records[0].Location)
		assert.Equal(t, firstCommentCount, responsePostFeed.Records[0].Comments)
		assert.Equal(t, firstPostLikes, responsePostFeed.Records[0].Likes)
		assert.Equal(t, false, responsePostFeed.Records[0].Liked)
		assert.Equal(t, nextPosts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
		assert.Equal(t, nextPosts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
		assert.Equal(t, nextPosts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

		assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
		assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
		assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
		assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
		assert.Equal(t, secondCommentCount, responsePostFeed.Records[1].Comments)
		assert.Equal(t, secondPostLikes, responsePostFeed.Records[1].Likes)
		assert.Equal(t, false, responsePostFeed.Records[1].Liked)
		assert.Nil(t, responsePostFeed.Records[1].Location)

		assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
		assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
		assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

		mockPostRepository.AssertExpectations(t)
		mockUserRepository.AssertExpectations(t)
		mockLikeRepository.AssertExpectations(t)
		mockCommentRepository.AssertExpectations(t)
	}
}

// TestGetGlobalPostFeedDefaultParameters tests if the GetPostFeed function returns an empty list when last post is not found and default parameters are used
func TestGetGlobalPostFeedDefaultParameters(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
	)
	feedController := controllers.NewFeedController(feedService)

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
	router.GET("/feed", feedController.GetPostFeed)
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
	mockLikeRepository := new(repositories.MockLikeRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		mockUserRepository,
		mockLikeRepository,
		mockCommentRepository,
	)
	feedController := controllers.NewFeedController(feedService)

	lastPost := models.Post{
		Id:        uuid.New(),
		Username:  "someUserTest",
		User:      models.User{},
		Content:   "This is the last post",
		CreatedAt: time.Now().UTC().Add(time.Hour * -1),
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
				Username: "someOtherUsername",
				Nickname: "someOtherNickname",
			},
			Content:    "This is the next post",
			CreatedAt:  time.Now().UTC().Add(time.Hour * -2),
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
				Username: "anotherTestUsername",
				Nickname: "anotherTestNickname",
			},
			Content:   "This is another next post",
			CreatedAt: time.Now().UTC().Add(time.Hour * -3),
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

	firstPostLikes := int64(0)
	secondPostLikes := int64(10)
	mockLikeRepository.On("CountLikes", nextPosts[0].Id.String()).Return(firstPostLikes, nil)
	mockLikeRepository.On("CountLikes", nextPosts[1].Id.String()).Return(secondPostLikes, nil)

	firstPostComments := int64(0)
	secondPostComments := int64(5)
	mockCommentRepository.On("CountComments", nextPosts[0].Id.String()).Return(firstPostComments, nil)
	mockCommentRepository.On("CountComments", nextPosts[1].Id.String()).Return(secondPostComments, nil)

	mockLikeRepository.On("FindLike", nextPosts[0].Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound) // First post not liked by current user
	mockLikeRepository.On("FindLike", nextPosts[1].Id.String(), currentUsername).Return(&models.Like{}, nil)                    // Second post liked by current user

	// Setup HTTP request
	url := "/feed?postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit) + "&feedType=personal"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/feed", feedController.GetPostFeed)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok
	var responsePostFeed models.GeneralFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
	assert.NoError(t, err)

	assert.Equal(t, lastPost.Id, capturedLastPost.Id)
	assert.Equal(t, lastPost.Username, capturedLastPost.Username)
	assert.Equal(t, lastPost.Content, capturedLastPost.Content)
	assert.Equal(t, lastPost.ImageId, capturedLastPost.ImageId)
	assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
	assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

	assert.Equal(t, nextPosts[0].Id, responsePostFeed.Records[0].PostId)
	assert.Equal(t, nextPosts[0].Username, responsePostFeed.Records[0].Author.Username)
	assert.Equal(t, nextPosts[0].Content, responsePostFeed.Records[0].Content)
	assert.True(t, nextPosts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
	assert.Equal(t, firstPostComments, responsePostFeed.Records[0].Comments)
	assert.Equal(t, firstPostLikes, responsePostFeed.Records[0].Likes)
	assert.Equal(t, false, responsePostFeed.Records[0].Liked)
	assert.NotNil(t, responsePostFeed.Records[0].Location)
	assert.Equal(t, nextPosts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
	assert.Equal(t, nextPosts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
	assert.Equal(t, nextPosts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

	assert.Equal(t, nextPosts[1].Id, responsePostFeed.Records[1].PostId)
	assert.Equal(t, nextPosts[1].Username, responsePostFeed.Records[1].Author.Username)
	assert.Equal(t, nextPosts[1].Content, responsePostFeed.Records[1].Content)
	assert.True(t, nextPosts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
	assert.Equal(t, secondPostComments, responsePostFeed.Records[1].Comments)
	assert.Equal(t, secondPostLikes, responsePostFeed.Records[1].Likes)
	assert.Equal(t, true, responsePostFeed.Records[1].Liked)
	assert.Nil(t, responsePostFeed.Records[1].Location)

	assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
	assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
	assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
	mockUserRepository.AssertExpectations(t)
	mockLikeRepository.AssertExpectations(t)
	mockCommentRepository.AssertExpectations(t)
}

// TestGetPersonalPostFeedDefaultParameters tests if the GetPostFeed function returns a 200 OK and an empty list when last post is not found and default parameters are used
func TestGetPersonalPostFeeDefaultParameters(t *testing.T) {
	// Arrange
	mockUserRepository := new(repositories.MockUserRepository)
	mockPostRepository := new(repositories.MockPostRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		mockUserRepository,
		nil,
		nil,
	)
	feedController := controllers.NewFeedController(feedService)

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
	router.GET("/feed", feedController.GetPostFeed)
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

		feedService := services.NewFeedService(
			mockPostRepository,
			mockUserRepository,
			nil,
			nil,
		)
		feedController := controllers.NewFeedController(feedService)

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
		router.GET("/feed", feedController.GetPostFeed)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.Unauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockPostRepository.AssertExpectations(t)
	}
}

// TestGetPostsByHashtagSuccess tests if the GetPostsByHashtag function returns a list of posts and 200 ok if the hashtag exists
func TestGetPostsByHashtagSuccess(t *testing.T) {
	// Arrange
	mockPostRepository := new(repositories.MockPostRepository)
	mockLikeRepository := new(repositories.MockLikeRepository)
	mockCommentRepository := new(repositories.MockCommentRepository)

	feedService := services.NewFeedService(
		mockPostRepository,
		nil,
		mockLikeRepository,
		mockCommentRepository,
	)
	feedController := controllers.NewFeedController(feedService)

	hashtag := "Post"
	locationId := uuid.New()
	posts := []models.Post{
		{
			Id:       uuid.New(),
			Username: "testUser",
			User: models.User{
				Username: "testUser",
				Nickname: "testNickname",
			},
			Content:    "Test #Post 2",
			CreatedAt:  time.Now().UTC(),
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
				Username: "testUser",
				Nickname: "testNickname",
			},
			Content:   "Test #Post 3",
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
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
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}

	// Mock expectations
	var capturedLastPost *models.Post
	mockPostRepository.On("GetPostById", lastPost.Id.String()).Return(lastPost, nil)
	mockPostRepository.On("GetPostsByHashtag", hashtag, &lastPost, limit).
		Run(func(args mock.Arguments) {
			capturedLastPost = args.Get(1).(*models.Post) // Save argument to captor
		}).Return(posts, totalCount, nil)

	firstPostLikes := int64(0)
	secondPostLikes := int64(10)
	mockLikeRepository.On("CountLikes", posts[0].Id.String()).Return(firstPostLikes, nil)
	mockLikeRepository.On("CountLikes", posts[1].Id.String()).Return(secondPostLikes, nil)

	firstPostComments := int64(0)
	secondPostComments := int64(5)
	mockCommentRepository.On("CountComments", posts[0].Id.String()).Return(firstPostComments, nil)
	mockCommentRepository.On("CountComments", posts[1].Id.String()).Return(secondPostComments, nil)

	mockLikeRepository.On("FindLike", posts[0].Id.String(), currentUsername).Return(&models.Like{}, gorm.ErrRecordNotFound) // First post not liked by current user
	mockLikeRepository.On("FindLike", posts[1].Id.String(), currentUsername).Return(&models.Like{}, nil)                    // Second post liked by current user

	// Setup HTTP request
	url := "/posts?q=" + hashtag + "&postId=" + lastPost.Id.String() + "&limit=" + fmt.Sprint(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/posts", middleware.AuthorizeUser, feedController.GetPostsByHashtag)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 ok

	var responsePostFeed models.GeneralFeedDTO
	err = json.Unmarshal(w.Body.Bytes(), &responsePostFeed)
	assert.NoError(t, err)

	assert.Equal(t, lastPost.Id, capturedLastPost.Id)
	assert.Equal(t, lastPost.Username, capturedLastPost.Username)
	assert.Equal(t, lastPost.Content, capturedLastPost.Content)
	assert.Equal(t, lastPost.ImageId, capturedLastPost.ImageId)
	assert.True(t, lastPost.CreatedAt.Equal(capturedLastPost.CreatedAt))
	assert.Equal(t, lastPost.Hashtags, capturedLastPost.Hashtags)

	assert.Equal(t, posts[0].Id, responsePostFeed.Records[0].PostId)
	assert.Equal(t, posts[0].Username, responsePostFeed.Records[0].Author.Username)
	assert.Equal(t, posts[0].User.Nickname, responsePostFeed.Records[0].Author.Nickname)
	assert.Equal(t, posts[0].Content, responsePostFeed.Records[0].Content)
	assert.True(t, posts[0].CreatedAt.Equal(responsePostFeed.Records[0].CreationDate))
	assert.Equal(t, firstPostComments, responsePostFeed.Records[0].Comments)
	assert.Equal(t, firstPostLikes, responsePostFeed.Records[0].Likes)
	assert.Equal(t, false, responsePostFeed.Records[0].Liked)
	assert.NotNil(t, responsePostFeed.Records[0].Location)
	assert.Equal(t, posts[0].Location.Latitude, *responsePostFeed.Records[0].Location.Latitude)
	assert.Equal(t, posts[0].Location.Longitude, *responsePostFeed.Records[0].Location.Longitude)
	assert.Equal(t, posts[0].Location.Accuracy, *responsePostFeed.Records[0].Location.Accuracy)

	assert.Equal(t, posts[1].Id, responsePostFeed.Records[1].PostId)
	assert.Equal(t, posts[1].Username, responsePostFeed.Records[1].Author.Username)
	assert.Equal(t, posts[1].User.Nickname, responsePostFeed.Records[1].Author.Nickname)
	assert.Equal(t, posts[1].Content, responsePostFeed.Records[1].Content)
	assert.True(t, posts[1].CreatedAt.Equal(responsePostFeed.Records[1].CreationDate))
	assert.Equal(t, secondPostComments, responsePostFeed.Records[1].Comments)
	assert.Equal(t, secondPostLikes, responsePostFeed.Records[1].Likes)
	assert.Equal(t, true, responsePostFeed.Records[1].Liked)
	assert.Nil(t, responsePostFeed.Records[1].Location)

	assert.Equal(t, limit, responsePostFeed.Pagination.Limit)
	assert.Equal(t, totalCount, responsePostFeed.Pagination.Records)
	assert.Equal(t, responsePostFeed.Records[1].PostId.String(), responsePostFeed.Pagination.LastPostId)

	mockPostRepository.AssertExpectations(t)
	mockLikeRepository.AssertExpectations(t)
	mockCommentRepository.AssertExpectations(t)
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

		feedService := services.NewFeedService(
			mockPostRepository,
			nil,
			nil,
			nil,
		)
		feedController := controllers.NewFeedController(feedService)

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
		router.GET("/posts", middleware.AuthorizeUser, feedController.GetPostsByHashtag)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized

		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.Unauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}
