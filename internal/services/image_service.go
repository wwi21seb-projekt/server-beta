package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ImageServiceInterface interface {
	SaveImage(imageData string) (*models.Image, *customerrors.CustomError, int)
	GetImage(filename string) ([]byte, *customerrors.CustomError, int)
	DeleteImage(imageUrl string) (*customerrors.CustomError, int)
	GetImageMetadata(imageUrl string) (*models.Image, *customerrors.CustomError, int)
}

type ImageService struct {
	uploadPath string
	fileSystem repositories.FileSystemInterface
	validator  utils.ValidatorInterface
	imageRepo  repositories.ImageRepositoryInterface
}

// NewImageService can be used as a constructor to create a ImageService "object"
func NewImageService(fileSystem repositories.FileSystemInterface, validator utils.ValidatorInterface) *ImageService {
	uploadPath := os.Getenv("IMAGES_PATH")
	if err := fileSystem.CreateDirectory(uploadPath, os.ModePerm); err != nil {
		panic(err)
	}

	return &ImageService{
		uploadPath: uploadPath,
		fileSystem: fileSystem,
		validator:  validator,
	}
}

// SaveImage can be used in other services to save an image to the file system and return the image url
func getExtensionFromContentType(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpeg", nil
	case "image/png":
		return ".png", nil
	case "image/gif":
		return ".gif", nil
	case "image/webp":
		return ".webp", nil
	case "image/svg+xml":
		return ".svg", nil
	default:
		return "", errors.New("unsupported content type")
	}
}

func (service *ImageService) SaveImage(imageData string) (*models.Image, *customerrors.CustomError, int) {
	// Check file size
	const maxFileSize = 5 << 20 // 5 MB
	if len(imageData) > maxFileSize {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Validate the image
	valid, contentType, err := service.validator.ValidateImage(imageData)
	if err != nil {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}
	if !valid {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Get file extension from content type
	extension, err := getExtensionFromContentType(contentType)
	if err != nil {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Generate new filename that does not exist yet
	var filename string
	var fullPath string
	for {
		id := uuid.New()
		filename = id.String() + extension
		fullPath = filepath.Join(service.uploadPath, filename)
		if !service.fileSystem.DoesFileExist(fullPath) {
			break
		}
	}

	// Decode the Base64 string to get image dimensions and save the file
	decodedImageData, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		log.Println("Error decoding Base64 string:", err)
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Decode the image to get its dimensions
	var img image.Image
	var width, height int
	if contentType == "image/webp" {
		img, err = webp.Decode(bytes.NewReader(decodedImageData))
	} else {
		img, _, err = image.Decode(bytes.NewReader(decodedImageData))
	}
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}
	width = img.Bounds().Dx()
	height = img.Bounds().Dy()

	// Save file
	if err := service.fileSystem.WriteFile(fullPath, decodedImageData, 0666); err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	imageUrl := os.Getenv("SERVER_URL") + "/api/images/" + filename

	imageObj := models.Image{
		ImageUrl: imageUrl,
		Width:    width,
		Height:   height,
	}
	err = service.imageRepo.CreateImage(&imageObj)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	return &imageObj, nil, http.StatusCreated
}

// GetImage can be used in image controller to return an image from the file system
func (service *ImageService) GetImage(filename string) ([]byte, *customerrors.CustomError, int) {
	// Get absolute path of upload directory
	uploadDir, err := filepath.Abs(service.uploadPath)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Get absolute path of requested file
	filePath, err := filepath.Abs(filepath.Join(service.uploadPath, filename)) // abs calls clean internally to remove relative path elements (e.g. ../)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Check if requested file is in upload directory
	if !strings.HasPrefix(filePath, uploadDir+string(os.PathSeparator)) {
		return nil, customerrors.FileNotFound, http.StatusNotFound
	}

	imageData, err := service.fileSystem.ReadFile(filePath)
	if err != nil { // check for file not found error
		if os.IsNotExist(err) {
			return nil, customerrors.FileNotFound, http.StatusNotFound
		}
		return nil, customerrors.InternalServerError, http.StatusInternalServerError // else internal server error
	}

	return imageData, nil, http.StatusOK
}

// DeleteImage can be used in other services to delete an image from the file system
func (service *ImageService) DeleteImage(imageUrl string) (*customerrors.CustomError, int) {
	// Parse the URL
	parsedURL, err := url.Parse(imageUrl)
	if err != nil {
		return customerrors.BadRequest, http.StatusBadRequest
	}

	// Get the file path from the URL
	filename := filepath.Base(parsedURL.Path)
	if filename == "" {
		return customerrors.BadRequest, http.StatusBadRequest
	}
	filePath := filepath.Join(service.uploadPath, filename)

	if err := service.fileSystem.DeleteFile(filePath); err != nil { // check for file not found error
		if os.IsNotExist(err) {
			return nil, http.StatusOK // file already deleted
		}
		return customerrors.InternalServerError, http.StatusInternalServerError // else internal server error
	}

	return nil, http.StatusOK
}
func (service *ImageService) GetImageMetadata(imageUrl string) (*models.Image, *customerrors.CustomError, int) {
	imageData, err := service.imageRepo.GetImageMetadata(imageUrl)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	return imageData, nil, http.StatusOK
}
