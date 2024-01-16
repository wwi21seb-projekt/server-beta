package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"strings"
)

type ImageControllerInterface interface {
	GetImage(c *gin.Context)
}

type ImageController struct {
	imageService services.ImageServiceInterface
}

// NewImageController can be used as a constructor to create a ImageController "object"
func NewImageController(imageService services.ImageServiceInterface) *ImageController {
	return &ImageController{imageService: imageService}
}

// GetImage is a controller function that returns an image from the file system and can be called from the router
func (controller *ImageController) GetImage(c *gin.Context) {

	// Read image name from request
	filename := c.Param("filename")

	// Get image from service
	image, serviceErr, httpStatus := controller.imageService.GetImage(filename)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	// Respond based on file type
	if strings.HasSuffix(filename, ".jpeg") {
		c.Data(httpStatus, "image/jpeg", image)
		return
	}

	if strings.HasSuffix(filename, ".webp") {
		c.Data(httpStatus, "image/webp", image)
		return
	}
}
