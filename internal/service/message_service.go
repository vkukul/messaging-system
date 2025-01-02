package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vkukul/messaging-system/internal/models"
	"github.com/vkukul/messaging-system/pkg/database"
	"github.com/vkukul/messaging-system/pkg/redis"
)

const (
	webhookURL      = "https://httpbin.org/post"
	batchSize       = 2
	processInterval = 2 * time.Minute
	maxWorkers      = 5
	maxRetries      = 3
)

type MessageService struct {
	processing bool
	client     *http.Client
	mu         sync.RWMutex
	workers    chan struct{}
}

func NewMessageService() *MessageService {
	return &MessageService{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		workers: make(chan struct{}, maxWorkers),
	}
}

func (s *MessageService) StartProcessing() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.processing {
		return fmt.Errorf("message processing is already running")
	}

	s.processing = true
	go s.processMessages()
	return nil
}

func (s *MessageService) StopProcessing() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processing = false
}

func (s *MessageService) isProcessing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.processing
}

func (s *MessageService) processMessages() {
	ticker := time.NewTicker(processInterval)
	defer ticker.Stop()

	for s.isProcessing() {
		var messages []models.Message
		if err := database.DB.Where("sent = ?", false).Limit(batchSize).Find(&messages).Error; err != nil {
			log.Printf("Error fetching messages: %v", err)
			<-ticker.C
			continue
		}

		if len(messages) == 0 {
			<-ticker.C
			continue
		}

		// Process messages in parallel with worker pool
		var wg sync.WaitGroup
		for i := range messages {
			msg := &messages[i]
			s.workers <- struct{}{}
			wg.Add(1)
			go func() {
				defer func() {
					<-s.workers
					wg.Done()
				}()

				if err := s.sendMessageWithRetry(msg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}()
		}
		wg.Wait()

		<-ticker.C
	}
}

func (s *MessageService) sendMessageWithRetry(msg *models.Message) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := s.sendMessage(msg); err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

func (s *MessageService) sendMessage(msg *models.Message) error {
	ctx := context.Background()

	// Check rate limit before sending
	canSend, err := redis.CheckRateLimit(ctx, msg.To)
	if err != nil {
		log.Printf("Warning: Rate limit check failed: %v", err)
	} else if !canSend {
		return fmt.Errorf("rate limit exceeded for recipient %s", msg.To)
	}

	payload := map[string]string{
		"to":      msg.To,
		"content": msg.Content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	msg.MessageID = uuid.New().String()
	msg.Sent = true
	msg.SentAt = time.Now()

	// Cache the sent message
	if err := redis.CacheMessage(ctx, msg); err != nil {
		log.Printf("Warning: Failed to cache message: %v", err)
	}

	// Update the message in the database
	if err := database.DB.Save(msg).Error; err != nil {
		return fmt.Errorf("error updating message status: %v", err)
	}

	return nil
}

func (s *MessageService) GetSentMessages() ([]models.Message, error) {
	var messages []models.Message

	// Try to get from database
	if err := database.DB.Where("sent = ?", true).Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("error fetching sent messages: %v", err)
	}

	// For each message, check if we have a cached version
	ctx := context.Background()
	var wg sync.WaitGroup
	messagesChan := make(chan *models.Message, len(messages))

	for i := range messages {
		msg := &messages[i]
		if msg.MessageID == "" {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			cachedMsg, err := redis.GetCachedMessage(ctx, msg.MessageID)
			if err != nil {
				log.Printf("Warning: Failed to get cached message %s: %v", msg.MessageID, err)
				return
			}
			if cachedMsg != nil {
				messagesChan <- cachedMsg
			}
		}()
	}

	go func() {
		wg.Wait()
		close(messagesChan)
	}()

	result := make([]models.Message, 0, len(messages))
	for msg := range messagesChan {
		result = append(result, *msg)
	}

	return result, nil
}
