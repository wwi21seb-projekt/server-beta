package utils_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
)

// TestGenerateImageMetadataDTOFromImage tests the GenerateImageMetadataDTOFromImage function
func TestGenerateImageMetadataDTOFromImage(t *testing.T) {
	// Test case where the image is nil
	var nilImage *models.Image
	result := utils.GenerateImageMetadataDTOFromImage(nilImage)
	assert.Nil(t, result, "Expected result to be nil for nil image")

	// Test case where the image is valid
	image := &models.Image{
		Id:     uuid.New(),
		Format: "jpeg",
		Width:  1024,
		Height: 768,
		Tag:    time.Now().UTC(),
	}

	expectedUrl := utils.FormatImageUrl(image.Id.String(), image.Format)
	expectedDto := &models.ImageMetadataDTO{
		Url:    expectedUrl,
		Width:  image.Width,
		Height: image.Height,
		Tag:    image.Tag,
	}

	result = utils.GenerateImageMetadataDTOFromImage(image)
	assert.Equal(t, expectedDto, result, "Expected and actual ImageMetadataDTO do not match")
}

// TestGenerateUserDTOFromUser tests the GenerateUserDTOFromUser function
func TestGenerateUserDTOFromUser(t *testing.T) {
	// Test case where the user is nil
	var nilUser *models.User
	result := utils.GenerateUserDTOFromUser(nilUser)
	assert.Nil(t, result, "Expected result to be nil for nil user")

	// Test case where the user is valid
	image := models.Image{
		Id:     uuid.New(),
		Format: "png",
		Width:  200,
		Height: 200,
		Tag:    time.Now().UTC(),
	}

	user := &models.User{
		Username: "testUser",
		Nickname: "Test User",
		Image:    image,
	}

	pictureDto := utils.GenerateImageMetadataDTOFromImage(&user.Image)
	expectedDto := &models.UserDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Picture:  pictureDto,
	}

	result = utils.GenerateUserDTOFromUser(user)
	assert.Equal(t, expectedDto, result, "Expected and actual UserDTO do not match")

	// Test case where the user has no image
	user = &models.User{
		Username: "testUser",
		Nickname: "Test User",
		Image:    models.Image{},
	}

	expectedDto = &models.UserDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Picture:  nil,
	}

	result = utils.GenerateUserDTOFromUser(user)
	assert.Equal(t, expectedDto, result, "Expected and actual UserDTO do not match")
}

// TestGenerateLocationDTOFromLocation tests the GenerateLocationDTOFromLocation function
func TestGenerateLocationDTOFromLocation(t *testing.T) {
	// Test case where the location is nil
	var nilLocation *models.Location
	result := utils.GenerateLocationDTOFromLocation(nilLocation)
	assert.Nil(t, result, "Expected result to be nil for nil location")

	// Test case where the location is valid
	location := &models.Location{
		Id:        uuid.New(),
		Longitude: 10.0,
		Latitude:  20.0,
		Accuracy:  5.0,
	}

	expectedDto := &models.LocationDTO{
		Longitude: &location.Longitude,
		Latitude:  &location.Latitude,
		Accuracy:  &location.Accuracy,
	}

	result = utils.GenerateLocationDTOFromLocation(location)

	assert.Equal(t, *expectedDto.Longitude, *result.Longitude, "Expected and actual longitude do not match")
	assert.Equal(t, *expectedDto.Latitude, *result.Latitude, "Expected and actual latitude do not match")
	assert.Equal(t, *expectedDto.Accuracy, *result.Accuracy, "Expected and actual accuracy do not match")
}
