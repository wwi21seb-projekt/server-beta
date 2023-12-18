package utils

import "github.com/stretchr/testify/mock"

type MockValidator struct {
	mock.Mock
	Validator
}

// ValidateEmailExistance Only override ValidateEmailExistance function because it is the only function that needs to be mocked
func (m *MockValidator) ValidateEmailExistance(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}
