package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type WebSocketNotification struct {
	Type        string `json:"type"`
	Message     Message `json:"message"`
	RecipientIDs []int  `json:"recipient_ids"`
}

// NotifyWebSocketServer sends a notification to the WebSocket server about a new message
func NotifyWebSocketServer(message Message, recipientIDs []int) {
	wsServerURL := getEnvOrDefault("WEBSOCKET_SERVER_URL", "http://websocket-server:8081")
	
	notification := WebSocketNotification{
		Type:         "new_message",
		Message:      message,
		RecipientIDs: recipientIDs,
	}
	
	jsonData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal WebSocket notification: %v", err)
		return
	}
	
	// Send HTTP request to WebSocket server's notification endpoint
	resp, err := http.Post(wsServerURL+"/notify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to notify WebSocket server: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("WebSocket server returned error status: %d", resp.StatusCode)
	}
}
