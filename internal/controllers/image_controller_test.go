package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetImageSuccess tests the GetImage function
func TestGetImageSuccess(t *testing.T) {
	filenames := []string{
		"test.jpeg",
		"test.webp",
	}

	for _, filename := range filenames {
		// Arrange
		mockFileSystem := new(repositories.MockFileSystem)

		// Mock expectations
		mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)
		mockFileSystem.On("ReadFile", filename).Return([]byte("test"), nil)

		// Arrange
		imageService := services.NewImageService(mockFileSystem)
		imageController := NewImageController(imageService)

		// Setup HTTP request
		req, _ := http.NewRequest("GET", "/images/"+filename, nil)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/images/:filename", imageController.GetImage)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK

		mockFileSystem.AssertExpectations(t)
	}
}

// TestGetImageNotFound tests if the GetImage function returns a 404 error when the image does not exist
func TestGetImageNotFound(t *testing.T) {
	// Arrange
	mockFileSystem := new(repositories.MockFileSystem)

	// Mock expectations
	mockFileSystem.On("CreateDirectory", mock.AnythingOfType("string"), mock.AnythingOfType("fs.FileMode")).Return(nil)
	mockFileSystem.On("ReadFile", mock.AnythingOfType("string")).Return([]byte{}, fs.ErrNotExist)

	// Arrange
	imageService := services.NewImageService(mockFileSystem)
	imageController := NewImageController(imageService)

	// Setup HTTP request
	req, _ := http.NewRequest("GET", "/images/test.jpeg", nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/images/:filename", imageController.GetImage)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.FileNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockFileSystem.AssertExpectations(t)
}
