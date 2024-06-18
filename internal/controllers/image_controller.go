package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type ImageControllerInterface interface {
	GetImageById(c *gin.Context)
}

type ImageController struct {
	imageService services.ImageServiceInterface
}

// NewImageController can be used as a constructor to create a ImageController "object"
func NewImageController(imageService services.ImageServiceInterface) *ImageController {
	return &ImageController{imageService: imageService}
}

// GetImageById is a controller function that returns an image from the database and can be called from the router
func (controller *ImageController) GetImageById(c *gin.Context) {

	// Read image name from request
	imageId := c.Param("imageId")

	// Get image from service
	imageDto, serviceErr, httpStatus := controller.imageService.GetImageById(imageId)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	var contentType string
	switch {
	case imageDto.Format == "jpeg" || imageDto.Format == "jpg":
		contentType = "image/jpeg"
	case imageDto.Format == "png":
		contentType = "image/png"
	case imageDto.Format == "webp":
		contentType = "image/webp"
	case imageDto.Format == "svg":
		contentType = "image/svg+xml"
	default: // If the image format is not supported, return not found error
		c.JSON(http.StatusNotFound, gin.H{
			"error": customerrors.ImageNotFound,
		})
		return
	}

	c.Header("Content-Type", contentType)

	// Respond based on file type
	c.Data(httpStatus, contentType, imageDto.Data)
}
