package services

import "github.com/stretchr/testify/mock"

// MockMailService is a mock implementation of the MailServiceInterface
type MockMailService struct {
	mock.Mock
	MailServiceInterface
}

func (m *MockMailService) SendMail(receiver string, subject string, body string) error {
	args := m.Called(receiver, subject, body)
	return args.Error(0)
}
