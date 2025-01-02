package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vkukul/messaging-system/internal/models"
)

func TestInitRedis(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Successfully connect to Redis",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitRedis()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCacheMessage(t *testing.T) {
	// Initialize Redis for tests
	if err := InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		msg     *models.Message
		wantErr bool
	}{
		{
			name: "Successfully cache message",
			msg: &models.Message{
				ID:        1,
				To:        "+905551234567",
				Content:   "Test message",
				Sent:      true,
				SentAt:    time.Now(),
				MessageID: "test-message-id",
			},
			wantErr: false,
		},
		{
			name:    "Fail with nil message",
			msg:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CacheMessage(ctx, tt.msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the message was cached correctly
				cached, err := GetCachedMessage(ctx, tt.msg.MessageID)
				assert.NoError(t, err)
				assert.NotNil(t, cached)
				assert.Equal(t, tt.msg.ID, cached.ID)
				assert.Equal(t, tt.msg.To, cached.To)
				assert.Equal(t, tt.msg.Content, cached.Content)
			}
		})
	}
}

func TestGetCachedMessage(t *testing.T) {
	// Initialize Redis for tests
	if err := InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}

	ctx := context.Background()
	testMsg := &models.Message{
		ID:        1,
		To:        "+905551234567",
		Content:   "Test message",
		Sent:      true,
		SentAt:    time.Now(),
		MessageID: "test-get-message-id",
	}

	// Cache the test message
	if err := CacheMessage(ctx, testMsg); err != nil {
		t.Fatalf("Failed to cache test message: %v", err)
	}

	tests := []struct {
		name      string
		messageID string
		want      *models.Message
		wantErr   bool
	}{
		{
			name:      "Successfully get cached message",
			messageID: testMsg.MessageID,
			want:      testMsg,
			wantErr:   false,
		},
		{
			name:      "Message not found",
			messageID: "non-existent-id",
			want:      nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCachedMessage(ctx, tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.To, got.To)
					assert.Equal(t, tt.want.Content, got.Content)
				}
			}
		})
	}
}

func TestCheckRateLimit(t *testing.T) {
	// Initialize Redis for tests
	if err := InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}

	ctx := context.Background()
	recipient := "+905551234567"

	// Clear any existing rate limit
	if err := ClearRateLimit(ctx, recipient); err != nil {
		t.Fatalf("Failed to clear rate limit: %v", err)
	}

	tests := []struct {
		name       string
		recipient  string
		iterations int
		want       bool
	}{
		{
			name:       "Under rate limit",
			recipient:  recipient,
			iterations: 5,
			want:       true,
		},
		{
			name:       "Exceed rate limit",
			recipient:  recipient,
			iterations: 15,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear rate limit before each test
			assert.NoError(t, ClearRateLimit(ctx, tt.recipient))

			var lastResult bool
			for i := 0; i < tt.iterations; i++ {
				var err error
				lastResult, err = CheckRateLimit(ctx, tt.recipient)
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, lastResult)
		})
	}
}

func TestClearRateLimit(t *testing.T) {
	// Initialize Redis for tests
	if err := InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}

	ctx := context.Background()
	recipient := "+905551234567"

	tests := []struct {
		name      string
		recipient string
		wantErr   bool
	}{
		{
			name:      "Successfully clear rate limit",
			recipient: recipient,
			wantErr:   false,
		},
		{
			name:      "Clear non-existent rate limit",
			recipient: "non-existent",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a rate limit first
			_, err := CheckRateLimit(ctx, tt.recipient)
			assert.NoError(t, err)

			// Clear the rate limit
			err = ClearRateLimit(ctx, tt.recipient)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify rate limit was cleared
				count, err := CheckRateLimit(ctx, tt.recipient)
				assert.NoError(t, err)
				assert.True(t, count)
			}
		})
	}
}
