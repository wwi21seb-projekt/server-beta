package repositories

import "os"

type FileSystemInterface interface {
	WriteFile(filePath string, file []byte, perm os.FileMode) error
	DeleteFile(filePath string) error
	ReadFile(filePath string) ([]byte, error)
	DoesFileExist(filePath string) bool
	CreateDirectory(dirPath string, perm os.FileMode) error
}

type FileSystem struct {
}

// NewFileSystem can be used as a constructor to create a FileSystem "object"
func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

// WriteFile can be used in other services to save a file to the file system
func (service *FileSystem) WriteFile(filePath string, file []byte, perm os.FileMode) error {
	err := os.WriteFile(filePath, file, perm)
	return err
}

// DeleteFile can be used in other services to delete a file from the file system
func (service *FileSystem) DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	return err
}

// ReadFile can be used in other services to read a file from the file system
func (service *FileSystem) ReadFile(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	return file, err
}

// DoesFileExist can be used in other services to check if a file exists on the file system
func (service *FileSystem) DoesFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CreateDirectory can be used in other services to create a directory on the file system
func (service *FileSystem) CreateDirectory(dirPath string, perm os.FileMode) error {
	err := os.MkdirAll(dirPath, perm)
	return err
}
