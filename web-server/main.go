package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer CloseDB()


	// Create Gin router
	r := gin.Default()


	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, ApiResponse{
			Success: true,
			Message: "Server is running",
		})
	})

	// Root endpoint to redirect to frontend
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, ApiResponse{
			Success: true,
			Message: "Chat API Server",
			Data: map[string]string{
				"api_docs":     "http://localhost:8080/api"
			},
		})
	})

	api := r.Group("/api")
	{
		
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(AuthMiddleware())
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.GET("/users", authHandler.GetUsers)

			protected.POST("/messages", messageHandler.SendMessage)
			protected.GET("/messages", messageHandler.GetMessageHistory)
			protected.GET("/conversations/:user_id", messageHandler.GetConversation)
			protected.PUT("/messages/read", messageHandler.MarkAsRead)

			protected.POST("/media/upload", mediaHandler.UploadMedia)
			protected.GET("/media", mediaHandler.GetUserMedia)
		}

		api.GET("/media/:user_dir/:filename", mediaHandler.ServeMedia)
	}

	port := getEnvOrDefault("PORT", "8080")

	log.Printf("API Server starting on port %s", port)
	log.Printf("API base URL: http://localhost:%s/api", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
