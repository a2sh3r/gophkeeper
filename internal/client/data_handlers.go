package client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/models"
)

// CreateLoginPasswordData creates login/password data from user input
func CreateLoginPasswordData() ([]byte, string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter login: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read login")
	}
	login := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter password: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read password")
	}
	password := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter URL (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read URL")
	}
	url := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter notes (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read notes")
	}
	notes := strings.TrimSpace(scanner.Text())

	loginPasswordData := models.LoginPasswordData{
		Login:    login,
		Password: password,
		URL:      url,
		Notes:    notes,
	}

	data, err := json.Marshal(loginPasswordData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal login password data: %w", err)
	}

	metadata := fmt.Sprintf("Login: %s, URL: %s", login, url)
	return data, metadata, nil
}

// CreateTextData creates text data from user input
func CreateTextData() ([]byte, string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter text content: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read text content")
	}
	content := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter notes (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read notes")
	}
	notes := strings.TrimSpace(scanner.Text())

	textData := models.TextData{
		Content: content,
		Notes:   notes,
	}

	data, err := json.Marshal(textData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal text data: %w", err)
	}

	metadata := fmt.Sprintf("Length: %d characters", len(content))
	return data, metadata, nil
}

// CreateBinaryData creates binary data from file
func CreateBinaryData() ([]byte, string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter file path: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read file path")
	}
	filePath := strings.TrimSpace(scanner.Text())

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file info: %w", err)
	}

	fileName := fileInfo.Name()
	fileExt := filepath.Ext(fileName)
	mimeType := getMimeType(fileExt)

	fmt.Print("Enter notes (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read notes")
	}
	notes := strings.TrimSpace(scanner.Text())

	binaryData := models.BinaryData{
		FileName: fileName,
		Size:     int64(len(fileData)),
		MimeType: mimeType,
		Notes:    notes,
	}

	metadataBytes, err := json.Marshal(binaryData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal binary metadata: %w", err)
	}

	encodedData := base64.StdEncoding.EncodeToString(fileData)
	return []byte(encodedData), string(metadataBytes), nil
}

// CreateBankCardData creates bank card data from user input
func CreateBankCardData() ([]byte, string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter card number: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read card number")
	}
	cardNumber := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter expiry date (MM/YY): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read expiry date")
	}
	expiryDate := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter CVV: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read CVV")
	}
	cvv := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter cardholder name: ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read cardholder name")
	}
	cardholder := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter bank name (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read bank name")
	}
	bank := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter notes (optional): ")
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("failed to read notes")
	}
	notes := strings.TrimSpace(scanner.Text())

	bankCardData := models.BankCardData{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		Cardholder: cardholder,
		Bank:       bank,
		Notes:      notes,
	}

	data, err := json.Marshal(bankCardData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal bank card data: %w", err)
	}

	metadata := fmt.Sprintf("Card: %s, Bank: %s", cardNumber, bank)
	return data, metadata, nil
}

// getMimeType returns MIME type based on file extension
func getMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".zip":
		return "application/zip"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream"
	}
}
