package ws

import (
    "github.com/gorilla/websocket"
    "sync"
)

type Client struct {
    ID   string          // user ID
    Conn *websocket.Conn // websocket connection
    Send chan []byte     // channel for outgoing messages
    mu   sync.Mutex      // protects write to websocket
}

// Write safely to the websocket
func (c *Client) Write(message []byte) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.Conn.WriteMessage(websocket.TextMessage, message)
}
