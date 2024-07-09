package controllers_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetImageSuccess tests the GetImageById function
func TestGetImageSuccess(t *testing.T) {
	fileTypes := []string{
		"jpeg",
		"png",
		"webp",
	}
	for _, fileType := range fileTypes {
		// Arrange
		mockImageRepo := new(repositories.MockImageRepository)
		imageService := services.NewImageService(mockImageRepo)
		imageController := controllers.NewImageController(imageService)

		imageId := uuid.New()
		image := models.Image{
			Id:        imageId,
			Format:    fileType,
			ImageData: []byte("test"),
			Width:     100,
			Height:    200,
			Tag:       time.Now().UTC(),
		}

		mockImageRepo.On("GetImageById", imageId.String()).Return(&image, nil) // Expect image to be found

		// Setup HTTP request
		url := "/images/" + imageId.String() + "." + fileType
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.GET("/images/:imageId", imageController.GetImageById)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK

		contentType := w.Header().Get("Content-Type")
		assert.Equal(t, "image/"+fileType, contentType)
		assert.Equal(t, image.ImageData, w.Body.Bytes())

		mockImageRepo.AssertExpectations(t)
	}

}

// TestGetImageNotFound tests if the GetImageById function returns a 404 error when the image does not exist
func TestGetImageNotFound(t *testing.T) {
	// Arrange
	mockImageRepo := new(repositories.MockImageRepository)
	imageService := services.NewImageService(mockImageRepo)
	imageController := controllers.NewImageController(imageService)

	imageId := uuid.New()

	mockImageRepo.On("GetImageById", imageId.String()).Return(&models.Image{}, gorm.ErrRecordNotFound) // Expect image not to be found

	// Setup HTTP request
	req := httptest.NewRequest("GET", "/images/"+imageId.String(), nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/images/:imageId", imageController.GetImageById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ImageNotFound
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)

	mockImageRepo.AssertExpectations(t)
}

// TestGetImageNotFoundNoFormat tests if the GetImageById function returns a 404 error when no format is provided in the image id
func TestGetImageNotFoundNoFormat(t *testing.T) {
	// Arrange
	mockImageRepo := new(repositories.MockImageRepository)
	imageService := services.NewImageService(mockImageRepo)
	imageController := controllers.NewImageController(imageService)

	imageId := uuid.New()

	mockImageRepo.On("GetImageById", imageId.String()).Return(&models.Image{}, nil) // Expect image to be found

	// Setup HTTP request
	req := httptest.NewRequest("GET", "/images/"+imageId.String(), nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/images/:imageId", imageController.GetImageById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ImageNotFound
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)

	mockImageRepo.AssertExpectations(t)
}

// TestGetImageNotFoundWrongFormat tests if the GetImageById function returns a 404 error when the format of the image is not the same as the requested format
func TestGetImageNotFoundWrongFormat(t *testing.T) {
	// Arrange
	mockImageRepo := new(repositories.MockImageRepository)
	imageService := services.NewImageService(mockImageRepo)
	imageController := controllers.NewImageController(imageService)

	imageId := uuid.New()
	image := models.Image{
		Id:        imageId,
		Format:    "jpeg",
		ImageData: []byte("test"),
		Width:     100,
		Height:    200,
		Tag:       time.Now().UTC(),
	}

	mockImageRepo.On("GetImageById", imageId.String()).Return(&image, nil) // Expect image to be found

	// Setup HTTP request
	req := httptest.NewRequest("GET", "/images/"+imageId.String()+".png", nil)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/images/:imageId", imageController.GetImageById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ImageNotFound
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)

	mockImageRepo.AssertExpectations(t)
}
