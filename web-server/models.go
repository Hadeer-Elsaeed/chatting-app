package main

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID          int       `json:"id"`
	SenderID    int       `json:"sender_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"` // ("direct", "broadcast")
	MediaURL    *string   `json:"media_url"`
	MediaType   *string   `json:"media_type"`
	CreatedAt   time.Time `json:"created_at"`
	
	SenderUsername string              `json:"sender_username,omitempty"`
	Recipients     []MessageRecipient  `json:"recipients,omitempty"`
}

type MessageRecipient struct {
	ID          int       `json:"id"`
	MessageID   int       `json:"message_id"`
	RecipientID int       `json:"recipient_id"`
	IsRead      bool      `json:"is_read"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time `json:"created_at"`
	
	// Additional fields for API responses
	RecipientUsername string `json:"recipient_username,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type SendMessageRequest struct {
	Content     string   `json:"content" binding:"required"`
	MessageType string   `json:"message_type" binding:"required,oneof=direct broadcast"`
	Recipients  []int    `json:"recipients"` // For direct messages, will be empty for broadcast
	MediaURL    *string  `json:"media_url"`
	MediaType   *string  `json:"media_type"`
}

type MessageHistoryResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type UserListResponse struct {
	Users []User `json:"users"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
