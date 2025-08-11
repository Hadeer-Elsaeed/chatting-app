package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize hub
	hub := NewHub(db)
	go hub.Run()

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Notification endpoint for REST API to notify about new messages
	r.POST("/notify", func(c *gin.Context) {
		var notification struct {
			Type         string  `json:"type"`
			Message      Message `json:"message"`
			RecipientIDs []int   `json:"recipient_ids"`
		}

		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if notification.Type == "new_message" {
			hub.NotifyNewMessage(notification.Message, notification.RecipientIDs)
		}

		c.JSON(http.StatusOK, gin.H{"status": "notification sent"})
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("WebSocket Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
