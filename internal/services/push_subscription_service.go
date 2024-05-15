package services

import (
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
	"os"
)

type PushSubscriptionServiceInterface interface {
	GetVapidKey() (*models.VapidKeyResponseDTO, *customerrors.CustomError, int)
	CreatePushSubscription(req *models.PushSubscriptionRequestDTO, currentUsername string) (*models.PushSubscription, *customerrors.CustomError, int)
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
	// type either needs to be web or expo
	if req.Type != "web" && req.Type != "expo" {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Create a new push subscription
	// TODO: Implement

	return nil, nil, http.StatusNotImplemented
}

// SendPushMessages sends push messages to all push subscriptions of a user
func (service *PushSubscriptionService) SendPushMessages(body string, toUsername string) {
	// TODO: Finalize

	// Get all push subscriptions by username
	pushSubscriptions, err := service.PushSubscriptionRepo.GetPushSubscriptionsByUsername(toUsername)
	if err != nil {
		return
	}

	// Send push messages
	for _, pushSubscription := range pushSubscriptions {
		fmt.Println(pushSubscription)
		sub := &webpush.Subscription{
			Endpoint: "",
			Keys: webpush.Keys{
				P256dh: "",
				Auth:   "",
			},
		}

		resp, err := webpush.SendNotification([]byte(body), sub, &webpush.Options{
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
