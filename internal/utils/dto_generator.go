package utils

import (
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

// GenerateImageMetadataDTOFromImage generates an ImageMetadataDTO from an Image and returns nil if the Image is nil
func GenerateImageMetadataDTOFromImage(image *models.Image) *models.ImageMetadataDTO {
	if image == nil || image.Id == uuid.Nil { // additional uuid.Nil check because user.Image is not a pointer and thus can't be nil
		return nil
	}

	return &models.ImageMetadataDTO{
		Url:    FormatImageUrl(image.Id.String(), image.Format),
		Width:  image.Width,
		Height: image.Height,
		Tag:    image.Tag,
	}
}

// GenerateUserDTOFromUser generates a UserDTO from a User and returns nil if the User is nil
// This function is used multiple times in the services to easily convert a User to a DTO with correct image wrapping
func GenerateUserDTOFromUser(user *models.User) *models.UserDTO {
	if user == nil || user.Username == "" {
		return nil
	}

	pictureDto := GenerateImageMetadataDTOFromImage(&user.Image)

	return &models.UserDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Picture:  pictureDto,
	}
}

// GenerateLocationDTOFromLocation generates a LocationDTO from a Location and returns nil if the Location is nil
func GenerateLocationDTOFromLocation(location *models.Location) *models.LocationDTO {
	if location == nil || location.Id == uuid.Nil {
		return nil
	}

	// Copy values to avoid pointer issues when this function is used in a loop
	tempLongitude := location.Longitude
	tempLatitude := location.Latitude
	tempAccuracy := location.Accuracy

	return &models.LocationDTO{
		Longitude: &tempLongitude,
		Latitude:  &tempLatitude,
		Accuracy:  &tempAccuracy,
	}
}
