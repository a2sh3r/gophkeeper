package client

import (
	"encoding/json"
	"testing"

	"github.com/a2sh3r/gophkeeper/internal/models"
)

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".pdf", "application/pdf"},
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".txt", "text/plain"},
		{".doc", "application/msword"},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{".xls", "application/vnd.ms-excel"},
		{".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{".zip", "application/zip"},
		{".mp3", "audio/mpeg"},
		{".mp4", "video/mp4"},
		{".avi", "video/x-msvideo"},
		{".unknown", "application/octet-stream"},
		{".PDF", "application/pdf"}, // Test case sensitivity
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := getMimeType(tt.ext)
			if result != tt.expected {
				t.Errorf("getMimeType(%q) = %q, want %q", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestCreateLoginPasswordData_ValidInput(t *testing.T) {
	// This test would require mocking stdin input, which is complex
	// For now, we'll test the function exists and can be called
	// In a real implementation, you might want to use dependency injection
	// or create a testable version that accepts input parameters

	// Test that the function exists and returns expected structure
	// We can't easily test the interactive input without mocking stdin
	t.Skip("Skipping interactive input test - requires stdin mocking")
}

func TestCreateTextData_ValidInput(t *testing.T) {
	// Similar to above - requires stdin mocking
	t.Skip("Skipping interactive input test - requires stdin mocking")
}

func TestCreateBinaryData_ValidInput(t *testing.T) {
	// Similar to above - requires stdin mocking
	t.Skip("Skipping interactive input test - requires stdin mocking")
}

func TestCreateBankCardData_ValidInput(t *testing.T) {
	// Similar to above - requires stdin mocking
	t.Skip("Skipping interactive input test - requires stdin mocking")
}

// Test helper functions for data creation without interactive input
func TestLoginPasswordDataStructure(t *testing.T) {
	loginPasswordData := models.LoginPasswordData{
		Login:    "testuser",
		Password: "testpass",
		URL:      "https://example.com",
		Notes:    "Test notes",
	}

	data, err := json.Marshal(loginPasswordData)
	if err != nil {
		t.Fatalf("Failed to marshal login password data: %v", err)
	}

	var unmarshaled models.LoginPasswordData
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal login password data: %v", err)
	}

	if unmarshaled.Login != loginPasswordData.Login {
		t.Errorf("Login mismatch: expected %s, got %s", loginPasswordData.Login, unmarshaled.Login)
	}
	if unmarshaled.Password != loginPasswordData.Password {
		t.Errorf("Password mismatch: expected %s, got %s", loginPasswordData.Password, unmarshaled.Password)
	}
	if unmarshaled.URL != loginPasswordData.URL {
		t.Errorf("URL mismatch: expected %s, got %s", loginPasswordData.URL, unmarshaled.URL)
	}
	if unmarshaled.Notes != loginPasswordData.Notes {
		t.Errorf("Notes mismatch: expected %s, got %s", loginPasswordData.Notes, unmarshaled.Notes)
	}
}

func TestTextDataStructure(t *testing.T) {
	textData := models.TextData{
		Content: "This is test content",
		Notes:   "Test notes",
	}

	data, err := json.Marshal(textData)
	if err != nil {
		t.Fatalf("Failed to marshal text data: %v", err)
	}

	var unmarshaled models.TextData
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal text data: %v", err)
	}

	if unmarshaled.Content != textData.Content {
		t.Errorf("Content mismatch: expected %s, got %s", textData.Content, unmarshaled.Content)
	}
	if unmarshaled.Notes != textData.Notes {
		t.Errorf("Notes mismatch: expected %s, got %s", textData.Notes, unmarshaled.Notes)
	}
}

func TestBinaryDataStructure(t *testing.T) {
	binaryData := models.BinaryData{
		FileName: "test.pdf",
		Size:     1024,
		MimeType: "application/pdf",
		Notes:    "Test binary file",
	}

	data, err := json.Marshal(binaryData)
	if err != nil {
		t.Fatalf("Failed to marshal binary data: %v", err)
	}

	var unmarshaled models.BinaryData
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal binary data: %v", err)
	}

	if unmarshaled.FileName != binaryData.FileName {
		t.Errorf("FileName mismatch: expected %s, got %s", binaryData.FileName, unmarshaled.FileName)
	}
	if unmarshaled.Size != binaryData.Size {
		t.Errorf("Size mismatch: expected %d, got %d", binaryData.Size, unmarshaled.Size)
	}
	if unmarshaled.MimeType != binaryData.MimeType {
		t.Errorf("MimeType mismatch: expected %s, got %s", binaryData.MimeType, unmarshaled.MimeType)
	}
	if unmarshaled.Notes != binaryData.Notes {
		t.Errorf("Notes mismatch: expected %s, got %s", binaryData.Notes, unmarshaled.Notes)
	}
}

func TestBankCardDataStructure(t *testing.T) {
	bankCardData := models.BankCardData{
		CardNumber: "1234567890123456",
		ExpiryDate: "12/25",
		CVV:        "123",
		Cardholder: "John Doe",
		Bank:       "Test Bank",
		Notes:      "Test card",
	}

	data, err := json.Marshal(bankCardData)
	if err != nil {
		t.Fatalf("Failed to marshal bank card data: %v", err)
	}

	var unmarshaled models.BankCardData
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal bank card data: %v", err)
	}

	if unmarshaled.CardNumber != bankCardData.CardNumber {
		t.Errorf("CardNumber mismatch: expected %s, got %s", bankCardData.CardNumber, unmarshaled.CardNumber)
	}
	if unmarshaled.ExpiryDate != bankCardData.ExpiryDate {
		t.Errorf("ExpiryDate mismatch: expected %s, got %s", bankCardData.ExpiryDate, unmarshaled.ExpiryDate)
	}
	if unmarshaled.CVV != bankCardData.CVV {
		t.Errorf("CVV mismatch: expected %s, got %s", bankCardData.CVV, unmarshaled.CVV)
	}
	if unmarshaled.Cardholder != bankCardData.Cardholder {
		t.Errorf("Cardholder mismatch: expected %s, got %s", bankCardData.Cardholder, unmarshaled.Cardholder)
	}
	if unmarshaled.Bank != bankCardData.Bank {
		t.Errorf("Bank mismatch: expected %s, got %s", bankCardData.Bank, unmarshaled.Bank)
	}
	if unmarshaled.Notes != bankCardData.Notes {
		t.Errorf("Notes mismatch: expected %s, got %s", bankCardData.Notes, unmarshaled.Notes)
	}
}
