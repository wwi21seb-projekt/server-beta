package services

import (
	"fmt"
	"net/smtp"
	"os"
)

type MailServiceInterface interface {
	SendMail(receiver string, subject string, body string) error
}

type MailService struct {
}

// NewMailService can be used as a constructor to generate a new MailService "object"
func NewMailService() *MailService {
	return &MailService{}
}

// SendMail sends a mail to a receiver with the given subject and body text
func (service *MailService) SendMail(receiver string, subject string, body string) error {

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")
	address := host + ":" + port

	to := []string{receiver}

	// Construct the email message with headers
	header := ""
	header += fmt.Sprintf("From: %s\r\n", from)
	header += fmt.Sprintf("To: %s\r\n", receiver)
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += "MIME-version: 1.0;\r\n"
	header += "Content-Type: text/plain; charset=\"UTF-8\";\r\n"
	header += "\r\n" // Blank line to separate headers from body

	message := []byte(header + body)

	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(address, auth, from, to, message)

	return err
}
