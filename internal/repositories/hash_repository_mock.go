package repositories

import (
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockHashtagRepository struct {
	mock.Mock
}

func (m *MockHashtagRepository) FindOrCreateHashtag(name string) (models.Hashtag, error) {
	args := m.Called(name)
	if item, ok := args.Get(0).(models.Hashtag); ok {
		return item, args.Error(1)
	}
	return models.Hashtag{}, args.Error(1)
}
