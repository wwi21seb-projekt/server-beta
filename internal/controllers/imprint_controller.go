package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ImprintControllerInterface interface {
	GetImprint(c *gin.Context)
}

type ImprintController struct {
}

// NewImprintController can be used as a constructor to return a new ImprintController "object"
func NewImprintController() *ImprintController {
	return &ImprintController{}
}

// GetImprint returns imprint information
func (controller ImprintController) GetImprint(c *gin.Context) {

	responseBody := struct {
		Text string `json:"text"`
	}{
		Text: "Diese Website wird im Rahmen eines Studienprojekts angeboten von:" +
			"\n\nKurs WWI21SEB" +
			"\nDuale Hochschule Baden-Württemberg Mannheim" +
			"\nCoblitzallee 1-9, 68163 Mannheim" +
			"\n\nKontakt:\nEmail:projekt.serverbeta@gmail.com" +
			"\n\nDie Nutzung von auf dieser Website veröffentlichten Kontaktdaten durch Dritte zur Übersendung von nicht ausdrücklich angeforderter Werbung und Informationsmaterialien wird hiermit ausdrücklich untersagt. Die Betreiber der Seiten behalten sich ausdrücklich rechtliche Schritte im Falle der unverlangten Zusendung von Werbeinformationen, etwa durch Spam-Mails, vor." +
			"\n\nDiese Webseite wurde im Rahmen eines Studienprojekts erstellt und dient ausschließlich zu nicht-kommerziellen und zu Lernzwecken. Es wird keine Garantie für die Richtigkeit, Vollständigkeit und Aktualität der bereitgestellten Inhalte übernommen. Jegliche Haftung ist ausgeschlossen.",
	}

	// Respond
	c.JSON(http.StatusOK, responseBody)
}
