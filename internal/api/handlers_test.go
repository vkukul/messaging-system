package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vkukul/messaging-system/pkg/database"
	"github.com/vkukul/messaging-system/pkg/redis"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router)
	return router
}

func TestStartProcessingHandler(t *testing.T) {
	if err := redis.InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}
	if err := database.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	router := setupTestRouter()

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantResponse map[string]string
	}{
		{
			name:       "Successfully start processing",
			method:     "POST",
			path:       "/api/v1/messages/start",
			wantStatus: http.StatusOK,
			wantResponse: map[string]string{
				"message": "Message processing started",
			},
		},
		{
			name:       "Invalid method",
			method:     "GET",
			path:       "/api/v1/messages/start",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantResponse != nil {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResponse, response)
			}
		})
	}
}

func TestStopProcessingHandler(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantResponse map[string]string
	}{
		{
			name:       "Successfully stop processing",
			method:     "POST",
			path:       "/api/v1/messages/stop",
			wantStatus: http.StatusOK,
			wantResponse: map[string]string{
				"message": "Message processing stopped",
			},
		},
		{
			name:       "Invalid method",
			method:     "GET",
			path:       "/api/v1/messages/stop",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantResponse != nil {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResponse, response)
			}
		})
	}
}

func TestGetSentMessagesHandler(t *testing.T) {
	if err := redis.InitRedis(); err != nil {
		t.Fatalf("Failed to initialize Redis: %v", err)
	}
	if err := database.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	router := setupTestRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "Successfully get sent messages",
			method:     "GET",
			path:       "/api/v1/messages/sent",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid method",
			method:     "POST",
			path:       "/api/v1/messages/sent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var messages []interface{}
				err := json.Unmarshal(w.Body.Bytes(), &messages)
				assert.NoError(t, err)
			}
		})
	}
}
