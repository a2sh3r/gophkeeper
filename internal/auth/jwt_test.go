package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		userID   uuid.UUID
		username string
		wantErr  bool
	}{
		{
			name:     "valid user",
			userID:   uuid.New(),
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "empty username",
			userID:   uuid.New(),
			username: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager("test-secret", time.Hour)
			token, err := manager.GenerateToken(tt.userID, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("Generated token is empty")
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	userID := uuid.New()
	username := "testuser"

	token, err := manager.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims.UserID != userID {
					t.Errorf("Expected UserID %v, got %v", userID, claims.UserID)
				}
				if claims.Username != username {
					t.Errorf("Expected Username %s, got %s", username, claims.Username)
				}
			}
		})
	}
}

func TestJWTManager_ValidateToken_Expired(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Millisecond)
	userID := uuid.New()
	username := "testuser"

	token, err := manager.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = manager.ValidateToken(token)
	if err != ErrTokenExpired {
		t.Errorf("Expected ErrTokenExpired, got %v", err)
	}
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret1", time.Hour)
	manager2 := NewJWTManager("secret2", time.Hour)
	userID := uuid.New()
	username := "testuser"

	token, err := manager1.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = manager2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}
