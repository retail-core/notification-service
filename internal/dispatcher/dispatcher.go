package dispatcher

import (
	"encoding/json"
	"github.com/retail-core/notification-service/internal/logger"
	"go.uber.org/zap"
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

// Register binds a channel name (e.g. "email") to its handler
func (d *Dispatcher) Register(channel NotChannel, handler Handler) {
	d.handlers[channel] = handler
	logger.L().Info("Registered handler for channel", zap.String("channel", string(channel)))
}

func (d *Dispatcher) Dispatch(rawMsg []byte)  {
	var notification Notification
	if err := json.Unmarshal(rawMsg, &notification); err != nil {
		logger.L().Error("Failed to unmarshal notification", zap.Error(err))
		return
	}

	handler, ok := d.handlers[notification.Channel]
	if !ok {
		logger.L().Error("No handler registered for channel", zap.String("channel", string(notification.Channel)))
		return
	}

	if err := handler.Send(notification); err != nil {
		logger.L().Error("Failed to send notification", zap.Error(err))
	}
	// } else {
	// 	logger.L().Info("Notification sent", zap.String("channel", strings.ToUpper(string(notification.Channel))), zap.String("to", notification.To))
	// }
}
