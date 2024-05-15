package services

import (
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
)

type PushSubscriptionServiceInterface interface {
	GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int)
	CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscription, *customerrors.CustomError, int)
}

type PushSubscriptionService struct {
	PushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface
}

// NewPushSubscriptionService can be used as a constructor to create a PushSubscriptionService "object"
func NewPushSubscriptionService(pushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface) *PushSubscriptionService {
	return &PushSubscriptionService{PushSubscriptionRepo: pushSubscriptionRepo}
}

// GetVapidKey returns a VAPID key for clients to register for push notifications
func (service *PushSubscriptionService) GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int) {
	// TODO: Implement
	return nil, nil, http.StatusNotImplemented
}

// CreatePushSubscription saves a new push subscription key to the database to send notifications to the client
func (service *PushSubscriptionService) CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscription, *customerrors.CustomError, int) {
	// type either needs to be web or expo
	if req.Type != "web" && req.Type != "expo" {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Create a new push subscription
	// TODO: Implement

	return nil, nil, http.StatusNotImplemented
}
