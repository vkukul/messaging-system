package main

import (
	"log"

	"github.com/gin-gonic/gin"

	_ "github.com/vkukul/messaging-system/docs"
	"github.com/vkukul/messaging-system/internal/api"
	"github.com/vkukul/messaging-system/pkg/database"
	"github.com/vkukul/messaging-system/pkg/redis"
)

// @title           Messaging System API
// @version         1.0
// @description     An automatic message sending system that processes messages every 2 minutes.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize Redis (bonus feature)
	if err := redis.InitRedis(); err != nil {
		log.Printf("Warning: Failed to initialize Redis (bonus feature): %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup API routes
	api.SetupRoutes(r)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
