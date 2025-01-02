package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vkukul/messaging-system/internal/models"
	"github.com/vkukul/messaging-system/pkg/database"
	"github.com/vkukul/messaging-system/pkg/redis"
)

func setupTest(t *testing.T) {
	// Initialize Redis for tests
	if err := redis.InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
}

func TestNewMessageService(t *testing.T) {
	service := NewMessageService()
	assert.NotNil(t, service)
	assert.False(t, service.processing)
}

func TestStartProcessing(t *testing.T) {
	setupTest(t)
	service := NewMessageService()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Successfully start processing",
			setup: func() {
				service.processing = false
			},
			wantErr: false,
		},
		{
			name: "Already processing",
			setup: func() {
				service.processing = true
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.StartProcessing()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, service.processing)
				service.StopProcessing() // Clean up
			}
		})
	}
}

func TestStopProcessing(t *testing.T) {
	service := NewMessageService()
	service.processing = true
	service.StopProcessing()
	assert.False(t, service.processing)
}

func TestSendMessage(t *testing.T) {
	setupTest(t)
	service := NewMessageService()

	ctx := context.Background()
	msg := &models.Message{
		To:      "+905551234567",
		Content: "Test message",
	}

	// Create the message in the database first
	err := database.DB.Create(msg).Error
	assert.NoError(t, err)

	err = service.sendMessage(msg)
	assert.NoError(t, err)
	assert.True(t, msg.Sent)
	assert.NotEmpty(t, msg.MessageID)
	assert.NotZero(t, msg.SentAt)

	// Verify message was cached in Redis
	cached, err := redis.GetCachedMessage(ctx, msg.MessageID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, msg.ID, cached.ID)
	assert.Equal(t, msg.To, cached.To)
	assert.Equal(t, msg.Content, cached.Content)

	// Clean up
	database.DB.Unscoped().Delete(msg)
}

func TestGetSentMessages(t *testing.T) {
	setupTest(t)
	service := NewMessageService()

	// Create and send test messages
	testMessages := []*models.Message{
		{
			To:      "+905551234567",
			Content: "Test message 1",
		},
		{
			To:      "+905551234568",
			Content: "Test message 2",
		},
	}

	// Create messages in the database
	for _, msg := range testMessages {
		err := database.DB.Create(msg).Error
		assert.NoError(t, err)
		err = service.sendMessage(msg)
		assert.NoError(t, err)
	}

	// Get sent messages
	messages, err := service.GetSentMessages()
	assert.NoError(t, err)
	assert.NotEmpty(t, messages)

	// Verify messages are returned correctly
	for _, msg := range messages {
		assert.True(t, msg.Sent)
		assert.NotEmpty(t, msg.MessageID)
		assert.NotZero(t, msg.SentAt)
	}

	// Clean up
	for _, msg := range testMessages {
		database.DB.Unscoped().Delete(msg)
	}
}
