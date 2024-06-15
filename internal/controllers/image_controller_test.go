package controllers_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetImageSuccess tests the GetImageById function
func TestGetImageSuccess(t *testing.T) {
	filenames := []string{
		"test.jpeg",
		"test.webp",
	}

	for _, filename := range filenames {
		// Arrange
		mockFileSystem := new(repositories.MockFileSystem)
		mockValidator := new(utils.MockValidator)

		// Mock expectations
		var pathCaptor string
		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)
		mockFileSystem.On("ReadFile", mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				pathCaptor = args.Get(0).(string) // Save argument to captor
			}).Return([]byte("test"), nil)

		// Arrange
		imageService := services.NewImageService(mockFileSystem, mockValidator)
		imageController := controllers.NewImageController(imageService)

		// Setup HTTP request
		req, _ := http.NewRequest("GET", "/images/"+filename, nil)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/images/:filename", imageController.GetImageById)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK

		assert.Contains(t, pathCaptor, filename)

		mockFileSystem.AssertExpectations(t)
	}
}

// TestGetImagePathTraversal tests if the GetImageById function prevents path traversal and removes relative paths
func TestGetImagePathTraversal(t *testing.T) {
	filenames := []string{
		"../test.jpeg",
		"../test.webp",
		"../../test.jpeg",
		"../../../test.webp",
	}

	for _, filename := range filenames {
		// Arrange
		mockFileSystem := new(repositories.MockFileSystem)
		mockValidator := new(utils.MockValidator)

		// Mock expectations
		var pathCaptor string
		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)
		mockFileSystem.On("ReadFile", mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				pathCaptor = args.Get(0).(string) // Save argument to captor
			}).Return([]byte("test"), nil)

		// Arrange
		imageService := services.NewImageService(mockFileSystem, mockValidator)

		// Act
		imageData, err, statusCode := imageService.GetImageById(filename)

		// Assert
		assert.Equal(t, http.StatusNotFound, statusCode) // Expect HTTP 404 Not Found
		assert.Equal(t, customerrors.ImageNotFound, err)
		assert.Nil(t, imageData)

		assert.NotContains(t, pathCaptor, "..", "Path contains directory traversal characters")
	}
}

// TestGetImageNotFound tests if the GetImageById function returns a 404 error when the image does not exist
func TestGetImageNotFound(t *testing.T) {
	// Arrange
	mockFileSystem := new(repositories.MockFileSystem)
	mockValidator := new(utils.MockValidator)

	// Mock expectations
	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)
	mockFileSystem.On("ReadFile", mock.AnythingOfType("string")).Return([]byte{}, fs.ErrNotExist)

	// Arrange
	imageService := services.NewImageService(mockFileSystem, mockValidator)
	imageController := controllers.NewImageController(imageService)

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/images/test.jpeg", nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/images/:filename", imageController.GetImageById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ImageNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockFileSystem.AssertExpectations(t)
}
