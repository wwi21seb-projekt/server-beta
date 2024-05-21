package services

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ImageServiceInterface interface {
	SaveImage(fileHeader multipart.FileHeader) (string, *customerrors.CustomError, int)
	GetImage(filename string) ([]byte, *customerrors.CustomError, int)
	DeleteImage(filename string) (*customerrors.CustomError, int)
}

type ImageService struct {
	uploadPath string
	fileSystem repositories.FileSystemInterface
	validator  utils.ValidatorInterface
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

	// Check file size
	const maxFileSize = 5 << 20 // 5 MB
	if fileHeader.Size > maxFileSize {
		return "", customerrors.FileTooLarge, http.StatusBadRequest
	}

	// Extract file from fileHeader
	file, err := fileHeader.Open()
	if err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// Read file data into byte array
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Check if image is valid
	if !service.validator.ValidateImage(imageData, fileHeader.Header.Get("Content-Type")) {
		return "", customerrors.BadRequest, http.StatusBadRequest
	}

	// Generate new filename that does not exist yet
	var filename string
	var fullPath string
	for {
		filename = uuid.New().String() + extension
		fullPath = fmt.Sprintf("%s/%s", service.uploadPath, filename)
		if service.fileSystem.DoesFileExist(fullPath) == false {
			break
		}
	}

	// Save file
	if err := service.fileSystem.WriteFile(fullPath, imageData, 0666); err != nil {
		return "", customerrors.InternalServerError, http.StatusInternalServerError
	}

	imageUrl := os.Getenv("SERVER_URL") + "/api/images/" + filename
	return imageUrl, nil, http.StatusCreated
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
func (service *ImageService) DeleteImage(filename string) (*customerrors.CustomError, int) {
	filePath := filepath.Join(service.uploadPath, filename)

	if err := service.fileSystem.DeleteFile(filePath); err != nil { // check for file not found error
		if os.IsNotExist(err) {
			return nil, http.StatusOK // file already deleted
		}
		return customerrors.InternalServerError, http.StatusInternalServerError // else internal server error
	}

	return nil, http.StatusOK
}
