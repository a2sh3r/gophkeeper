package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/models"
)

// ErrNotAuthenticated is returned when session is not authenticated
var ErrNotAuthenticated = fmt.Errorf("session not authenticated - please login first")

// RegisterCommand handles user registration
func (s *ClientSession) RegisterCommand(ctx context.Context, username, password string, config *Config) error {
	if len(username) == 0 || len(password) == 0 {
		return fmt.Errorf("username and password are required")
	}

	fmt.Print("Enter master password for data encryption (min 8 characters): ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read master password")
	}
	masterPassword := scanner.Text()

	if len(masterPassword) < 8 {
		return fmt.Errorf("master password must be at least 8 characters long")
	}

	resp, err := s.Register(ctx, username, password, masterPassword)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	saltBytes, err := base64.StdEncoding.DecodeString(resp.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	cryptoManager, err := crypto.NewCryptoManagerWithSalt(masterPassword, saltBytes)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	s.SetCryptoManager(cryptoManager, masterPassword)

	config.Token = resp.Token
	config.Salt = resp.Salt
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.cli.SetToken(resp.Token)

	fmt.Printf("Successfully registered user: %s\n", resp.User.Username)
	fmt.Println("Master password set for data encryption")
	return nil
}

// LoginCommand handles user login
func (s *ClientSession) LoginCommand(ctx context.Context, username, password string, config *Config) error {
	if len(username) == 0 || len(password) == 0 {
		return fmt.Errorf("username and password are required")
	}

	resp, err := s.Login(ctx, username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Print("Enter master password for data decryption: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read master password")
	}
	masterPassword := scanner.Text()

	saltBytes, err := base64.StdEncoding.DecodeString(resp.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	cryptoManager, err := crypto.NewCryptoManagerWithSalt(masterPassword, saltBytes)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	s.SetCryptoManager(cryptoManager, masterPassword)

	config.Token = resp.Token
	config.Salt = resp.Salt
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.cli.SetToken(resp.Token)

	fmt.Printf("Successfully logged in as: %s\n", resp.User.Username)
	fmt.Println("Master password verified for data decryption")
	return nil
}

// ListCommand handles listing all data
func (s *ClientSession) ListCommand(ctx context.Context) error {
	data, err := s.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	if len(data) == 0 {
		fmt.Println("No data found")
		return nil
	}

	fmt.Printf("Found %d items:\n", len(data))
	for _, item := range data {
		fmt.Printf("  %s [%s] - %s", item.ID.String(), item.Type, CleanQuotes(item.Name))
		if item.Description != "" {
			fmt.Printf(" - %s", CleanQuotes(item.Description))
		}
		fmt.Printf("\n")
	}
	return nil
}

// GetCommand handles getting data by ID
func (s *ClientSession) GetCommand(ctx context.Context, id string) error {
	if len(id) == 0 {
		return fmt.Errorf("data ID is required")
	}

	data, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	return DisplayStructuredData(data, s.cryptoManager)
}

// CreateCommand handles creating new data
func (s *ClientSession) CreateCommand(ctx context.Context, dataType, name, description string) error {
	if !s.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	if len(dataType) == 0 || len(name) == 0 {
		return fmt.Errorf("data type and name are required")
	}

	var dataContent []byte
	var metadata string
	var err error

	switch dataType {
	case "login_password":
		dataContent, metadata, err = CreateLoginPasswordData()
	case "text":
		dataContent, metadata, err = CreateTextData()
	case "binary":
		dataContent, metadata, err = CreateBinaryData()
	case "bank_card":
		dataContent, metadata, err = CreateBankCardData()
	default:
		return fmt.Errorf("unknown data type: %s", dataType)
	}

	if err != nil {
		return fmt.Errorf("failed to create data content: %w", err)
	}

	encryptedData, err := s.cryptoManager.Encrypt(dataContent)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	dataReq := models.DataRequest{
		Type:        models.DataType(dataType),
		Name:        name,
		Description: description,
		Data:        encryptedData,
		Metadata:    metadata,
	}

	data, err := s.Create(ctx, dataReq)
	if err != nil {
		return fmt.Errorf("failed to create data: %w", err)
	}

	fmt.Printf("Successfully created encrypted data with ID: %s\n", data.ID)
	return nil
}

// UpdateCommand handles updating existing data
func (s *ClientSession) UpdateCommand(ctx context.Context, id string) error {
	if !s.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	if len(id) == 0 {
		return fmt.Errorf("data ID is required")
	}

	data, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	decryptedData, err := s.cryptoManager.Decrypt(data.Data)
	if err != nil {
		return fmt.Errorf("failed to decrypt current data: %w", err)
	}

	fmt.Printf("Current data: %s\n", string(decryptedData))
	fmt.Print("Enter new data content: ")
	scanner := bufio.NewScanner(os.Stdin)
	var newContent string
	if scanner.Scan() {
		newContent = scanner.Text()
	}

	encryptedContent, err := s.cryptoManager.Encrypt([]byte(newContent))
	if err != nil {
		return fmt.Errorf("failed to encrypt new data: %w", err)
	}

	dataReq := models.DataRequest{
		Type:        data.Type,
		Name:        data.Name,
		Description: data.Description,
		Data:        encryptedContent,
		Metadata:    data.Metadata,
	}

	updatedData, err := s.Update(ctx, id, dataReq)
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	fmt.Printf("Successfully updated encrypted data: %s\n", updatedData.ID)
	return nil
}

// DeleteCommand handles deleting data
func (s *ClientSession) DeleteCommand(ctx context.Context, id string) error {
	if len(id) == 0 {
		return fmt.Errorf("data ID is required")
	}

	fmt.Printf("Are you sure you want to delete data with ID %s? (y/N): ", id)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read confirmation")
	}
	confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("Deletion cancelled")
		return nil
	}

	err := s.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	fmt.Printf("Successfully deleted data: %s\n", id)
	return nil
}

// SaveCommand handles saving binary data to file
func (s *ClientSession) SaveCommand(ctx context.Context, id, outputPath string) error {
	if !s.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	if len(id) == 0 {
		return fmt.Errorf("data ID is required")
	}

	data, err := s.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	if data.Type != "binary" {
		return fmt.Errorf("data with ID %s is not binary type (type: %s)", id, data.Type)
	}

	var binaryData models.BinaryData
	if err := json.Unmarshal([]byte(data.Metadata), &binaryData); err != nil {
		return fmt.Errorf("failed to parse binary metadata: %w", err)
	}

	if outputPath == "" {
		outputPath = binaryData.FileName
	}

	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/N): ", outputPath)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("failed to read overwrite confirmation")
		}
		confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if confirmation != "y" && confirmation != "yes" {
			fmt.Println("Save cancelled")
			return nil
		}
	}

	decryptedData, err := s.cryptoManager.Decrypt(data.Data)
	if err != nil {
		return fmt.Errorf("failed to decrypt binary data: %w", err)
	}

	fileData, err := base64.StdEncoding.DecodeString(string(decryptedData))
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	err = os.WriteFile(outputPath, fileData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Successfully saved decrypted binary data to: %s\n", outputPath)
	fmt.Printf("File: %s\n", binaryData.FileName)
	fmt.Printf("Size: %d bytes\n", binaryData.Size)
	fmt.Printf("MIME Type: %s\n", binaryData.MimeType)
	if binaryData.Notes != "" {
		fmt.Printf("Notes: %s\n", binaryData.Notes)
	}
	return nil
}
