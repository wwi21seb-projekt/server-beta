package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
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
