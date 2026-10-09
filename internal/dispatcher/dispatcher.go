package dispatcher

import (
	"encoding/json"
	"fmt"

	"github.com/retail-core/notification-service/internal/logger"
	"go.uber.org/zap"
)

type NotChannel string

// PermanentError marks a failure that retrying cannot fix.
type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }


const (
	EmailChannel NotChannel = "email"
	SMSChannel   NotChannel = "sms"
	PushChannel  NotChannel = "push"
	InAppChannel NotChannel = "inapp"
	InventoryUpdateChannel NotChannel = "inventory_update"
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
}

func (d *Dispatcher) DispatchOld(rawMsg []byte)  {
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
}


func (d *Dispatcher) Dispatch(rawMsg []byte) error {
	var notification Notification
	if err := json.Unmarshal(rawMsg, &notification); err != nil {
		return &PermanentError{Err: fmt.Errorf("invalid payload: %w", err)}
	}

	handler, ok := d.handlers[notification.Channel]
	if !ok {
		return &PermanentError{Err: fmt.Errorf("no handler for channel %s", notification.Channel)}
	}

	return handler.Send(notification)
}


func (d *Dispatcher) Dispatch_Old(rawMsg []byte) error {
	var notification Notification
	if err := json.Unmarshal(rawMsg, &notification); err != nil {
		logger.L().Error("Failed to unmarshal notification", zap.Error(err))
		return err // bad payload — don't requeue, see note below
	}

	handler, ok := d.handlers[notification.Channel]
	if !ok {
		logger.L().Error("No handler registered for channel", zap.String("channel", string(notification.Channel)))
		return fmt.Errorf("no handler for channel %s", notification.Channel)
	}

	if err := handler.Send(notification); err != nil {
		logger.L().Error("Failed to send notification", zap.Error(err))
		return err
	}

	return nil
}