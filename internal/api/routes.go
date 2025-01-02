package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	handlers "github.com/vkukul/messaging-system/internal/api/handlers"
	service "github.com/vkukul/messaging-system/internal/service"
)

func SetupRoutes(r *gin.Engine) {
	// Create message service
	messageService := service.NewMessageService()

	// Create handlers
	messageHandlers := handlers.NewMessageHandlers(messageService)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		messages := v1.Group("/messages")
		{
			messages.POST("/start", messageHandlers.StartProcessing)
			messages.POST("/stop", messageHandlers.StopProcessing)
			messages.GET("/sent", messageHandlers.GetSentMessages)
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
