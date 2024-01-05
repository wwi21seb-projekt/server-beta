package services

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ImageServiceInterface interface {
	SaveImage(fileHeader multipart.FileHeader) (string, *customerrors.CustomError, int)
	GetImage(filename string) ([]byte, *customerrors.CustomError, int)
	//DeleteImage(filename string) (*customerrors.CustomError, int)
}

type ImageService struct {
	uploadPath string
}

// NewImageService can be used as a constructor to create a ImageService "object"
func NewImageService() *ImageService {
	uploadPath := os.Getenv("IMAGES_PATH")
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		panic(err)
	}

	return &ImageService{
		uploadPath: uploadPath,
	}
}

// SaveImage can be used in other services to save an image to the file system and return the image url
func (service *ImageService) SaveImage(fileHeader multipart.FileHeader) (string, *customerrors.CustomError, int) {
	// Get file type
	var extension string
	switch fileHeader.Header.Get("Content-Type") {
	case "image/jpeg":
		extension = ".jpeg"
	case "image/webp":
		extension = ".webp"
	default:
		return "", customerrors.BadRequest, http.StatusBadRequest
	}

	// Extract file from fileHeader
	file, err := fileHeader.Open()
	if err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}
	defer file.Close()

	// Read file data into byte array
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Generate new filename that does not exist yet
	var filename string
	var fullPath string
	for {
		filename = uuid.New().String() + extension
		fullPath = fmt.Sprintf("%s/%s", service.uploadPath, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			break
		}
	}

	// TODO: Check for file size and compress if necessary

	// Save file
	if err := os.WriteFile(fullPath, imageData, 0666); err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}

	imageUrl := os.Getenv("IMAGES_URL") + filename
	return imageUrl, nil, http.StatusOK
}

// GetImage can be used in image controller to return an image from the file system
func (service *ImageService) GetImage(filename string) ([]byte, *customerrors.CustomError, int) {
	filePath := filepath.Join(service.uploadPath, filename)

	imageData, err := os.ReadFile(filePath)
	if err != nil { // check for file not found error
		if os.IsNotExist(err) {
			return nil, customerrors.PreliminaryFileNotFound, http.StatusNotFound
		}
		return nil, customerrors.InternalServerError, http.StatusInternalServerError // else internal server error
	}

	return imageData, nil, http.StatusOK
}
