package dispatcher

import (
	"encoding/json"
	"log"
	"strings"
)

type NotChannel string

const (
	EmailChannel NotChannel = "email"
	SMSChannel   NotChannel = "sms"
	PushChannel  NotChannel = "push"
)

type Notification struct {
	Channel  NotChannel         `json:"channel"`
	To       string             `json:"to"`
	Template string             `json:"template"`
	Data     map[string]string  `json:"data"`
}

type Handler interface {
	Send(notification Notification) error
}

type Dispatcher struct {
	handlers map[NotChannel]Handler
}

func New() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[NotChannel]Handler),
	}
}

func (d *Dispatcher) Dispatch(rawMsg []byte)  {
	var notification Notification
	if err := json.Unmarshal(rawMsg, &notification); err != nil {
		log.Printf("failed to unmarshal notification: %v", err)
		return
	}

	handler, ok := d.handlers[notification.Channel]
	if !ok {
		log.Printf("no handler found for channel: %s", notification.Channel)
		return
	}

	if err := handler.Send(notification); err != nil {
		log.Printf("failed to send notification: %v", err)
	} else {
		log.Printf("notification sent via %s to %s", strings.ToUpper(string(notification.Channel)), notification.To)
	}
}