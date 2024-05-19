package services

import (
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
	"os"
)

type PushSubscriptionServiceInterface interface {
	GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int)
	CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscription, *customerrors.CustomError, int)
	SendPushMessages(notificationObject interface{}, toUsername string)
}

type PushSubscriptionService struct {
	PushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface

	vapidPrivateKey string
	vapidPublicKey  string
	serverMail      string
}

// NewPushSubscriptionService can be used as a constructor to create a PushSubscriptionService "object"
func NewPushSubscriptionService(pushSubscriptionRepo repositories.PushSubscriptionRepositoryInterface) *PushSubscriptionService {
	vapidPrivateKey := os.Getenv("VAPID_PRIVATE_KEY")
	vapidPublicKey := os.Getenv("VAPID_PUBLIC_KEY")
	serverMail := os.Getenv("EMAIL_ADDRESS")

	return &PushSubscriptionService{PushSubscriptionRepo: pushSubscriptionRepo, vapidPrivateKey: vapidPrivateKey, vapidPublicKey: vapidPublicKey, serverMail: serverMail}
}

// GetVapidKey returns a VAPID key for clients to register for push notifications
func (service *PushSubscriptionService) GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int) {
	// Return response object with public key
	response := models.VapidKeyResponseDTO{
		Key: service.vapidPublicKey,
	}
	return &response, nil, http.StatusOK
}

// CreatePushSubscription saves a new push subscription key to the database to send notifications to the client
func (service *PushSubscriptionService) CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscription, *customerrors.CustomError, int) {
	// type either needs to be "web" or "expo"
	if req.Type != "web" && req.Type != "expo" {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Create a new push subscription
	newPushSubscription := models.PushSubscription{
		Id:       uuid.New(),
		Username: currentUsername,
		Type:     req.Type,
		Endpoint: req.SubscriptionInfo.Endpoint,
		P256dh:   req.SubscriptionInfo.P256dh,
		Auth:     req.SubscriptionInfo.Auth,
	}

	err := service.PushSubscriptionRepo.CreatePushSubscription(&newPushSubscription)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	return &newPushSubscription, nil, http.StatusCreated
}

// SendPushMessages sends push messages to all push subscriptions of a user
func (service *PushSubscriptionService) SendPushMessages(notificationObject interface{}, toUsername string) {
	// Create notification json string from object
	notificationJson, err := json.Marshal(notificationObject)
	if err != nil {
		return
	}
	notificationString := string(notificationJson)

	// Get all push subscriptions by username
	pushSubscriptions, err := service.PushSubscriptionRepo.GetPushSubscriptionsByUsername(toUsername)
	if err != nil {
		return
	}

	// Send push messages
	for _, pushSubscription := range pushSubscriptions {
		sub := &webpush.Subscription{
			Endpoint: pushSubscription.Endpoint,
			Keys: webpush.Keys{
				P256dh: pushSubscription.P256dh,
				Auth:   pushSubscription.Auth,
			},
		}

		resp, err := webpush.SendNotification([]byte(notificationString), sub, &webpush.Options{
			Subscriber:      service.serverMail,
			VAPIDPrivateKey: service.vapidPrivateKey,
			TTL:             30,
		})

		if err != nil {
			print(err, "err sending notification")
		}

		// Ensure the response body is closed after reading
		func() {
			defer resp.Body.Close()
			fmt.Println("Response: ", resp.Status)
		}()
	}
}
