package main

import (
	"time"
)

type Message struct {
	ID             int       `json:"id"`
	SenderID       int       `json:"sender_id"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	MediaURL       *string   `json:"media_url"`
	MediaType      *string   `json:"media_type"`
	CreatedAt      time.Time `json:"created_at"`
	SenderUsername string    `json:"sender_username"`
}

type WebSocketMessage struct {
	Type        string      `json:"type"`
	RecipientID int         `json:"recipient_id,omitempty"`
	Content     string      `json:"content,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
