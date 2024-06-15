package services

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ImageServiceInterface interface {
	GetImageById(imageId string) (*models.ImageDTO, *customerrors.CustomError, int)
}

type ImageService struct {
	imageRepo repositories.ImageRepositoryInterface
}

// NewImageService can be used as a constructor to create a ImageService "object"
func NewImageService(imageRepo repositories.ImageRepositoryInterface) *ImageService {
	return &ImageService{imageRepo: imageRepo}
}

// GetImageById can be used in image controller to return an image from the database
func (service *ImageService) GetImageById(imageId string) (*models.ImageDTO, *customerrors.CustomError, int) {
	// Image id consists of the image name and the file format
	// The image name is the primary key in the database
	// The image format is the file format of the image
	imageIdSeperated := strings.SplitN(imageId, ".", 2)

	image, err := service.imageRepo.GetImageById(imageIdSeperated[0])
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ImageNotFound, http.StatusNotFound
		}
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Check if the format of the image is the same as the requested format
	if image.Format != imageIdSeperated[1] {
		return nil, customerrors.ImageNotFound, http.StatusNotFound
	}

	// Create response
	response := models.ImageDTO{
		Format: image.Format,
		Data:   image.ImageData,
	}

	return &response, nil, http.StatusOK
}
