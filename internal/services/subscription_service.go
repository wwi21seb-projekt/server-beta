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
	// MARC: hier übernimmst du dann den createRequest und den username des aktuellen nutzers aus dem controller

	// Erstellen einer neuen Subscription-Instanz
	newSubscription := models.Subscription{
		SubscriptionId: uuid.New(),
		// MARC: wir brauchen auch noch einen Zeitstempel, wann die Subscription erstellt wurde
		Follower:  models.User{Username: follower}, // MARC: ich bin mir nicht sicher ob das so gut funktioniert, weil du erstellst hier einen neuen user, was ist wenn der user nicht existiert?
		Following: models.User{Username: following},
	}

	// MARC: außerdem müssen noch weitere Fehlerfälle (siehe Postman) überprüft werden
	// User Not Found (s. o.)
	// Self follow

	// Speichern der Subscription in der Datenbank
	err := service.subscriptionRepo.CreateSubscription(&newSubscription)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	// Erstellen des Antwortobjekts

	return &models.SubscriptionPostResponseDTO{
		SubscriptionId: uuid.New(), // MARC: hier dann auch created at zurückgeben
		Follower:       follower,
		Following:      following,
	}, nil, http.StatusCreated
}

func (service *SubscriptionService) DeleteSubscription(subscriptionId uuid.UUID) (*models.SubscriptionDeleteResponseDTO, *customerrors.CustomError, int) {
	// MARC: am besten erst subscription auslesen (falls nciht vorhanden, dann 404 Not Found)
	// dann überprüfen, ob der eingeloggte nutzer der follower ist, wenn nicht darf er nicht löschen
	// dann löschen

	// Löschen der Subscription aus der Datenbank
	err := service.subscriptionRepo.DeleteSubscription(subscriptionId)
	if err != nil {
		return nil, customerrors.InternalServerError, http.StatusInternalServerError
	}

	return &models.SubscriptionDeleteResponseDTO{ // MARC: es gibt bei delete kein response body, deswegen auch 204 no content
		SubscriptionId: subscriptionId,
		Follower:       "exampleFollower",
		Following:      "exampleFollowing",
	}, nil, http.StatusNoContent

	// MARC: weiterhin hast du jetzt einen mock für den subscription service geschrieben und dann nur den controller getestet und den service weggemockt
	// bis jetzt habe ich immer den service gar nicht "weggemockt" sondern nur das repository, und dann beim controller test den service gleich mitgetestet
	// Und bei den Tests fehlen dann die tests für die einzelnen Fehlerfälle

}
