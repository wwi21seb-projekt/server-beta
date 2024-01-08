package services

import (
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/repositories"
	"net/http"
)

type SubscriptionServiceInterface interface {
	PostSubscription(follower, following string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int)
	DeleteSubscription(subscriptionId uuid.UUID) (*models.SubscriptionDeleteResponseDTO, *customerrors.CustomError, int)
}

type SubscriptionService struct {
	subscriptionRepo repositories.SubscriptionRepositoryInterface
}

func NewSubscriptionService(subscriptionRepo repositories.SubscriptionRepositoryInterface) *SubscriptionService {
	return &SubscriptionService{subscriptionRepo: subscriptionRepo}
}

func (service *SubscriptionService) PostSubscription(follower, following string) (*models.SubscriptionPostResponseDTO, *customerrors.CustomError, int) {
	// Erstellen einer neuen Subscription-Instanz
	newSubscription := models.Subscription{
		SubscriptionId: uuid.New(),
		Follower:       models.User{Username: follower},
		Following:      models.User{Username: following},
	}

	// Speichern der Subscription in der Datenbank
	err := service.subscriptionRepo.CreateSubscription(&newSubscription)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Erstellen des Antwortobjekts

	return &models.SubscriptionPostResponseDTO{
		SubscriptionId: uuid.New(),
		Follower:       follower,
		Following:      following,
	}, nil, http.StatusCreated
}

func (service *SubscriptionService) DeleteSubscription(subscriptionId uuid.UUID) (*models.SubscriptionDeleteResponseDTO, *customerrors.CustomError, int) {
	// Löschen der Subscription aus der Datenbank
	err := service.subscriptionRepo.DeleteSubscription(subscriptionId)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	return &models.SubscriptionDeleteResponseDTO{
		SubscriptionId: subscriptionId,
		Follower:       "exampleFollower",
		Following:      "exampleFollowing",
	}, nil, http.StatusNoContent
}
