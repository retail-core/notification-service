package ws

import (
	"github.com/retail-core/notification-service/internal/logger"
	"go.uber.org/zap"
)

type UserMessage struct {
	UserID string
	Data   []byte
}

type Hub struct {
	Clients    map[string]*Client // key: userID
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	SendToUser chan UserMessage
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
		SendToUser: make(chan UserMessage),
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	logger := logger.L()
	logger.Info("WebSocket Notification Hub started")
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ID] = client
			logger.Info("Client registered", zap.String("user_id", client.ID))

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ID]; ok {
				close(client.Send)
				delete(h.Clients, client.ID)
				logger.Info("Client unregistered", zap.String("user_id", client.ID))
			}

		case msg := <-h.SendToUser:
			if client, ok := h.Clients[msg.UserID]; ok {
				select {
				case client.Send <- msg.Data:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
					logger.Info("Client removed due to slow or blocked send channel", zap.String("user_id", client.ID))
				}
			}

		case message := <-h.Broadcast:
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					// Client is slow or blocked, remove it
					close(client.Send)
					delete(h.Clients, client.ID)
					logger.Info("Client removed due to slow or blocked send channel", zap.String("user_id", client.ID))
				}
			}
		}
	}
}
