package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/vkukul/messaging-system/internal/models"
)

const (
	MessageKeyPrefix = "message:"
	RateLimitPrefix  = "rate_limit:"
	CacheDuration    = 24 * time.Hour
	MaxRetries       = 3
	PoolSize         = 10
)

var (
	Client *redis.Client
	once   sync.Once
)

// InitRedis initializes the Redis client with connection pooling
func InitRedis() error {
	var initErr error
	once.Do(func() {
		host := getEnv("REDIS_HOST", "localhost")
		port := getEnv("REDIS_PORT", "6379")
		addr := fmt.Sprintf("%s:%s", host, port)

		Client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     "", // no password set
			DB:           0,  // use default DB
			PoolSize:     PoolSize,
			MinIdleConns: 2,
			MaxRetries:   MaxRetries,
		})

		ctx := context.Background()
		if err := Client.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("failed to connect to Redis: %v", err)
			return
		}
	})
	return initErr
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// withRetry executes a Redis operation with retries
func withRetry(operation func() error) error {
	var err error
	for i := 0; i < MaxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	return fmt.Errorf("operation failed after %d retries: %v", MaxRetries, err)
}

// CacheMessage stores a sent message in Redis with retries
func CacheMessage(ctx context.Context, msg *models.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	key := MessageKeyPrefix + msg.MessageID
	return withRetry(func() error {
		return Client.Set(ctx, key, string(data), CacheDuration).Err()
	})
}

// GetCachedMessage retrieves a message from Redis cache with retries
func GetCachedMessage(ctx context.Context, messageID string) (*models.Message, error) {
	if messageID == "" {
		return nil, fmt.Errorf("messageID cannot be empty")
	}

	key := MessageKeyPrefix + messageID
	var data string
	err := withRetry(func() error {
		var err error
		data, err = Client.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil
		}
		return err
	})

	if err == nil && data == "" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var msg models.Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}
	return &msg, nil
}

// CheckRateLimit checks if we can send more messages with retries
func CheckRateLimit(ctx context.Context, recipient string) (bool, error) {
	if recipient == "" {
		return false, fmt.Errorf("recipient cannot be empty")
	}

	key := RateLimitPrefix + recipient
	var count int64
	err := withRetry(func() error {
		var err error
		count, err = Client.Incr(ctx, key).Result()
		if err != nil {
			return err
		}

		if count == 1 {
			return Client.Expire(ctx, key, time.Minute).Err()
		}
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %v", err)
	}

	return count <= 10, nil
}

// ClearRateLimit clears the rate limit for a recipient with retries
func ClearRateLimit(ctx context.Context, recipient string) error {
	if recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}

	key := RateLimitPrefix + recipient
	return withRetry(func() error {
		return Client.Del(ctx, key).Err()
	})
}
