package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   int
	Username string
}

type Hub struct {
	Clients     map[*Client]bool
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	UserClients map[int]*Client // Map user ID to client for direct messaging
	DB          *sql.DB
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key-change-this-in-production")

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		UserClients: make(map[int]*Client),
		DB:          db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			h.UserClients[client.UserID] = client
			log.Printf("User %s (ID: %d) connected via WebSocket", client.Username, client.UserID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserClients, client.UserID)
				close(client.Send)
				log.Printf("User %s (ID: %d) disconnected from WebSocket", client.Username, client.UserID)
			}

		case message := <-h.Broadcast:
			// For broadcast messages, send to all connected clients
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
					delete(h.UserClients, client.UserID)
				}
			}
		}
	}
}

// NotifyNewMessage sends a notification to specific users about a new message
func (h *Hub) NotifyNewMessage(message Message, recipientIDs []int) {
	wsMsg := WebSocketMessage{
		Type: "new_message",
		Data: message,
	}

	msgData, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal message notification: %v", err)
		return
	}

	// Send to specific recipients if they are connected
	for _, recipientID := range recipientIDs {
		if client, exists := h.UserClients[recipientID]; exists {
			select {
			case client.Send <- msgData:
			default:
				// Client's send channel is full, skip
				log.Printf("Failed to send message notification to user %d: channel full", recipientID)
			}
		}
	}

	// Also send to sender if connected (for confirmation)
	if client, exists := h.UserClients[message.SenderID]; exists {
		select {
		case client.Send <- msgData:
		default:
		}
	}
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameters
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Token required", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "ping":
			// Handle ping for connection keepalive
			pongMsg := WebSocketMessage{Type: "pong"}
			if pongData, err := json.Marshal(pongMsg); err == nil {
				select {
				case c.Send <- pongData:
				default:
				}
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
