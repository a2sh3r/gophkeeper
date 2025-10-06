package client

import (
	"context"
	"testing"

	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/models"
)

func TestClientSession_NewClientSession(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	if session == nil {
		t.Fatal("NewClientSession returned nil")
	}

	if session.cli != cli {
		t.Error("Client not set correctly")
	}

	if session.cryptoManager != nil {
		t.Error("CryptoManager should be nil initially")
	}

	if session.masterPassword != "" {
		t.Error("MasterPassword should be empty initially")
	}
}

func TestClientSession_SetCryptoManager(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	session.SetCryptoManager(cryptoManager, "testpassword123")

	if session.cryptoManager != cryptoManager {
		t.Error("CryptoManager not set correctly")
	}

	if session.masterPassword != "testpassword123" {
		t.Error("MasterPassword not set correctly")
	}
}

func TestClientSession_IsAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	// Initially not authenticated
	if session.IsAuthenticated() {
		t.Error("Session should not be authenticated initially")
	}

	// Set crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	session.SetCryptoManager(cryptoManager, "testpassword123")

	// Now should be authenticated
	if !session.IsAuthenticated() {
		t.Error("Session should be authenticated after setting crypto manager")
	}
}

func TestClientSession_GetClient(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	if session.GetClient() != cli {
		t.Error("GetClient returned wrong client")
	}
}

func TestClientSession_GetCryptoManager(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	// Initially nil
	if session.GetCryptoManager() != nil {
		t.Error("CryptoManager should be nil initially")
	}

	// Set crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	session.SetCryptoManager(cryptoManager, "testpassword123")

	// Now should return the manager
	if session.GetCryptoManager() != cryptoManager {
		t.Error("GetCryptoManager returned wrong crypto manager")
	}
}

func TestClientSession_Register(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	// This will fail because there's no server, but we're testing the method exists
	_, err := session.Register(context.Background(), "testuser", "testpass", "masterpass123")
	if err == nil {
		t.Error("Expected error for register without server")
	}
}

func TestClientSession_Login(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	// This will fail because there's no server, but we're testing the method exists
	_, err := session.Login(context.Background(), "testuser", "testpass")
	if err == nil {
		t.Error("Expected error for login without server")
	}
}

func TestClientSession_List_NotAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	_, err := session.List(context.Background())
	if err != ErrNotAuthenticated {
		t.Errorf("Expected ErrNotAuthenticated, got %v", err)
	}
}

func TestClientSession_Get_NotAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	_, err := session.Get(context.Background(), "test-id")
	if err != ErrNotAuthenticated {
		t.Errorf("Expected ErrNotAuthenticated, got %v", err)
	}
}

func TestClientSession_Create_NotAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	dataReq := models.DataRequest{
		Type: "text",
		Name: "test",
		Data: []byte("test data"),
	}

	_, err := session.Create(context.Background(), dataReq)
	if err != ErrNotAuthenticated {
		t.Errorf("Expected ErrNotAuthenticated, got %v", err)
	}
}

func TestClientSession_Update_NotAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	dataReq := models.DataRequest{
		Type: "text",
		Name: "test",
		Data: []byte("test data"),
	}

	_, err := session.Update(context.Background(), "test-id", dataReq)
	if err != ErrNotAuthenticated {
		t.Errorf("Expected ErrNotAuthenticated, got %v", err)
	}
}

func TestClientSession_Delete_NotAuthenticated(t *testing.T) {
	cli := NewClient("http://localhost:8080")
	session := NewClientSession(cli)

	err := session.Delete(context.Background(), "test-id")
	if err != ErrNotAuthenticated {
		t.Errorf("Expected ErrNotAuthenticated, got %v", err)
	}
}
