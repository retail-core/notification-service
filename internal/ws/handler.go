package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/retail-core/notification-service/internal/errors"
	"github.com/retail-core/notification-service/internal/httpx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // allow all origins for now
}

func WebSocketHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id") // simple way to identify user, could be JWT in headers

		if userID == "" {
			httpx.WriteError(w, errors.BadRequest("user_id query parameter is required"))
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			httpx.WriteError(w, err)
			return
		}

		client := &Client{
			ID:   userID,
			Conn: conn,
			Send: make(chan []byte, 256), // buffered
		}

		hub.Register <- client

		// Start write pump
		go func() {
			for msg := range client.Send {
				client.Write(msg)
			}
		}()

		// Start read pump (optional, if you want to handle client messages)
		go func() {
			defer func() {
				hub.Unregister <- client
				conn.Close()
			}()
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
				// handle client messages here if needed
			}
		}()
	}
}
