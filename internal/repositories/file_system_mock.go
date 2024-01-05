package repositories

import (
	"github.com/stretchr/testify/mock"
	"os"
)

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) WriteFile(filePath string, file []byte, perm os.FileMode) error {
	args := m.Called(filePath, file, perm)
	return args.Error(0)
}

func (m *MockFileSystem) DeleteFile(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *MockFileSystem) ReadFile(filePath string) ([]byte, error) {
	args := m.Called(filePath)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileSystem) DoesFileExist(filePath string) bool {
	args := m.Called(filePath)
	return args.Bool(0)
}

func (m *MockFileSystem) CreateDirectory(dirPath string, perm os.FileMode) error {
	args := m.Called(dirPath, perm)
	return args.Error(0)
}
