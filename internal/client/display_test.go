package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
)

func TestCleanQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted string"`, "quoted string"},
		{`"single word"`, "single word"},
		{`"multiple words here"`, "multiple words here"},
		{`unquoted string`, "unquoted string"},
		{`"`, `"`},
		{`""`, ""},
		{`"partial quote`, `"partial quote`},
		{`partial quote"`, `partial quote"`},
		{`"nested "quotes" here"`, `nested "quotes" here`},
		{`  "  spaced  "  `, "  spaced  "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CleanQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("CleanQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDisplayStructuredData_LoginPassword(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create test data
	loginPasswordData := models.LoginPasswordData{
		Login:    "testuser",
		Password: "testpass",
		URL:      "https://example.com",
		Notes:    "Test notes",
	}

	dataBytes, err := json.Marshal(loginPasswordData)
	if err != nil {
		t.Fatalf("Failed to marshal login password data: %v", err)
	}

	encryptedData, err := cryptoManager.Encrypt(dataBytes)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	data := &models.Data{
		ID:          uuid.New(),
		Type:        "login_password",
		Name:        "Test Login",
		Description: "Test Description",
		Data:        encryptedData,
		Metadata:    "Login: testuser, URL: https://example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display - this will print to stdout, but we're testing it doesn't error
	err = DisplayStructuredData(data, cryptoManager)
	if err != nil {
		t.Errorf("DisplayStructuredData failed: %v", err)
	}
}

func TestDisplayStructuredData_Text(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create test data
	textData := models.TextData{
		Content: "This is test content",
		Notes:   "Test notes",
	}

	dataBytes, err := json.Marshal(textData)
	if err != nil {
		t.Fatalf("Failed to marshal text data: %v", err)
	}

	encryptedData, err := cryptoManager.Encrypt(dataBytes)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	data := &models.Data{
		ID:          uuid.New(),
		Type:        "text",
		Name:        "Test Text",
		Description: "Test Description",
		Data:        encryptedData,
		Metadata:    "Length: 20 characters",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display
	err = DisplayStructuredData(data, cryptoManager)
	if err != nil {
		t.Errorf("DisplayStructuredData failed: %v", err)
	}
}

func TestDisplayStructuredData_Binary(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create test data
	binaryData := models.BinaryData{
		FileName: "test.pdf",
		Size:     1024,
		MimeType: "application/pdf",
		Notes:    "Test binary file",
	}

	dataBytes, err := json.Marshal(binaryData)
	if err != nil {
		t.Fatalf("Failed to marshal binary data: %v", err)
	}

	encryptedData, err := cryptoManager.Encrypt(dataBytes)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	data := &models.Data{
		ID:          uuid.New(),
		Type:        "binary",
		Name:        "Test Binary",
		Description: "Test Description",
		Data:        encryptedData,
		Metadata:    string(dataBytes),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display
	err = DisplayStructuredData(data, cryptoManager)
	if err != nil {
		t.Errorf("DisplayStructuredData failed: %v", err)
	}
}

func TestDisplayStructuredData_BankCard(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create test data
	bankCardData := models.BankCardData{
		CardNumber: "1234567890123456",
		ExpiryDate: "12/25",
		CVV:        "123",
		Cardholder: "John Doe",
		Bank:       "Test Bank",
		Notes:      "Test card",
	}

	dataBytes, err := json.Marshal(bankCardData)
	if err != nil {
		t.Fatalf("Failed to marshal bank card data: %v", err)
	}

	encryptedData, err := cryptoManager.Encrypt(dataBytes)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	data := &models.Data{
		ID:          uuid.New(),
		Type:        "bank_card",
		Name:        "Test Card",
		Description: "Test Description",
		Data:        encryptedData,
		Metadata:    "Card: 1234567890123456, Bank: Test Bank",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display
	err = DisplayStructuredData(data, cryptoManager)
	if err != nil {
		t.Errorf("DisplayStructuredData failed: %v", err)
	}
}

func TestDisplayStructuredData_UnknownType(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create test data with unknown type
	encryptedData, err := cryptoManager.Encrypt([]byte("raw data"))
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	data := &models.Data{
		ID:          uuid.New(),
		Type:        "unknown_type",
		Name:        "Test Unknown",
		Description: "Test Description",
		Data:        encryptedData,
		Metadata:    "Unknown metadata",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display
	err = DisplayStructuredData(data, cryptoManager)
	if err != nil {
		t.Errorf("DisplayStructuredData failed: %v", err)
	}
}

func TestDisplayStructuredData_DecryptionError(t *testing.T) {
	// Create crypto manager
	cryptoManager, err := crypto.NewCryptoManager("testpassword123")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Create data with invalid encrypted content
	data := &models.Data{
		ID:          uuid.New(),
		Type:        "text",
		Name:        "Test Text",
		Description: "Test Description",
		Data:        []byte("invalid encrypted data"),
		Metadata:    "Test metadata",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test display - should return error
	err = DisplayStructuredData(data, cryptoManager)
	if err == nil {
		t.Error("Expected error for invalid encrypted data")
	}
}
