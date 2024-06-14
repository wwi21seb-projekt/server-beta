package services

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockImageService struct {
	mock.Mock
	ImageServiceInterface
}

func (m *MockImageService) SaveImage(imageData string) (*models.Image, *customerrors.CustomError, int) {
	args := m.Called(imageData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Image), nil, args.Int(2)
	}
	return nil, args.Get(1).(*customerrors.CustomError), args.Int(2)
}

func (m *MockImageService) GetImage(filename string) ([]byte, *customerrors.CustomError, int) {
	args := m.Called(filename)
	if args.Get(0) != nil {
		return args.Get(0).([]byte), nil, args.Int(2)
	}
	return nil, args.Get(1).(*customerrors.CustomError), args.Int(2)
}

func (m *MockImageService) DeleteImage(imageUrl string) (*customerrors.CustomError, int) {
	args := m.Called(imageUrl)
	if args.Get(0) != nil {
		return args.Get(0).(*customerrors.CustomError), args.Int(1)
	}
	return nil, args.Int(1)
}

func (m *MockImageService) GetImageMetadata(imageUrl string) (*models.Image, *customerrors.CustomError, int) {
	args := m.Called(imageUrl)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Image), nil, args.Int(2)
	}
	return nil, args.Get(1).(*customerrors.CustomError), args.Int(2)
}
