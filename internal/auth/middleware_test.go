package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", time.Hour)
	userID := uuid.New()
	username := "testuser"

	token, err := jwtManager.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectHandler  bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name:           "no authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:           "invalid format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:           "missing bearer",
			authHeader:     token,
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthMiddleware(jwtManager)
			handlerCalled := false

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if tt.expectHandler {
					if r.Header.Get("X-User-ID") != userID.String() {
						t.Errorf("Expected X-User-ID %s, got %s", userID.String(), r.Header.Get("X-User-ID"))
					}
					if r.Header.Get("X-Username") != username {
						t.Errorf("Expected X-Username %s, got %s", username, r.Header.Get("X-Username"))
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			middleware(w, req, handler.ServeHTTP)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if handlerCalled != tt.expectHandler {
				t.Errorf("Expected handler called %v, got %v", tt.expectHandler, handlerCalled)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "unauthorized error",
			status:  http.StatusUnauthorized,
			message: "Unauthorized",
		},
		{
			name:    "forbidden error",
			status:  http.StatusForbidden,
			message: "Forbidden",
		},
		{
			name:    "internal server error",
			status:  http.StatusInternalServerError,
			message: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeError(w, tt.status, tt.message)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}
