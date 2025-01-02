package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vkukul/messaging-system/internal/service"
)

type MessageHandlers struct {
	messageService *service.MessageService
}

// Response represents a standard API response
type Response struct {
	Message string `json:"message"`
}

// Message represents a message in the system
type Message struct {
	ID        uint   `json:"id"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Sent      bool   `json:"sent"`
	SentAt    string `json:"sent_at,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

func NewMessageHandlers(messageService *service.MessageService) *MessageHandlers {
	return &MessageHandlers{
		messageService: messageService,
	}
}

// StartProcessing godoc
// @Summary      Start message processing
// @Description  Start the automatic message sending process that sends messages every 2 minutes
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Router       /messages/start [post]
func (h *MessageHandlers) StartProcessing(c *gin.Context) {
	if err := h.messageService.StartProcessing(); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Message: "Message processing started"})
}

// StopProcessing godoc
// @Summary      Stop message processing
// @Description  Stop the automatic message sending process
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Router       /messages/stop [post]
func (h *MessageHandlers) StopProcessing(c *gin.Context) {
	h.messageService.StopProcessing()
	c.JSON(http.StatusOK, Response{Message: "Message processing stopped"})
}

// GetSentMessages godoc
// @Summary      Get sent messages
// @Description  Get a list of all messages that have been sent
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Success      200  {array}   Message
// @Failure      500  {object}  Response
// @Router       /messages/sent [get]
func (h *MessageHandlers) GetSentMessages(c *gin.Context) {
	messages, err := h.messageService.GetSentMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}
