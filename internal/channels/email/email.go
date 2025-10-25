package email

import (
	"fmt"
	"log"

	"github.com/retail-core/notification-service/internal/dispatcher"
)

type EmailHandler struct{}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (*EmailHandler) Send(notification dispatcher.Notification) error {
	log.Printf("Sending email to %s using template %s with data %v", notification.To, notification.Template, notification.Data)
	// Implement email sending logic here
	fmt.Printf("Sending Email ...")
	return nil
}