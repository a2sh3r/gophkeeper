package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/models"
)

// DisplayStructuredData displays structured data in a user-friendly format
func DisplayStructuredData(data *models.Data, cryptoManager *crypto.CryptoManager) error {
	decryptedData, err := cryptoManager.Decrypt(data.Data)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	fmt.Printf("ID: %s\n", data.ID.String())
	fmt.Printf("Type: %s\n", data.Type)
	fmt.Printf("Name: %s\n", CleanQuotes(data.Name))
	if data.Description != "" {
		fmt.Printf("Description: %s\n", CleanQuotes(data.Description))
	}
	fmt.Printf("Created: %s\n", data.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", data.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("---")

	switch data.Type {
	case "login_password":
		var loginPasswordData models.LoginPasswordData
		if err := json.Unmarshal(decryptedData, &loginPasswordData); err == nil {
			fmt.Printf("Login: %s\n", loginPasswordData.Login)
			fmt.Printf("Password: %s\n", loginPasswordData.Password)
			if loginPasswordData.URL != "" {
				fmt.Printf("URL: %s\n", loginPasswordData.URL)
			}
			if loginPasswordData.Notes != "" {
				fmt.Printf("Notes: %s\n", loginPasswordData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(decryptedData))
		}
	case "text":
		var textData models.TextData
		if err := json.Unmarshal(decryptedData, &textData); err == nil {
			fmt.Printf("Content: %s\n", textData.Content)
			if textData.Notes != "" {
				fmt.Printf("Notes: %s\n", textData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(decryptedData))
		}
	case "binary":
		var binaryData models.BinaryData
		if err := json.Unmarshal(decryptedData, &binaryData); err == nil {
			fmt.Printf("File: %s\n", binaryData.FileName)
			fmt.Printf("Size: %d bytes\n", binaryData.Size)
			fmt.Printf("MIME Type: %s\n", binaryData.MimeType)
			if binaryData.Notes != "" {
				fmt.Printf("Notes: %s\n", binaryData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(decryptedData))
		}
	case "bank_card":
		var bankCardData models.BankCardData
		if err := json.Unmarshal(decryptedData, &bankCardData); err == nil {
			fmt.Printf("Card Number: %s\n", bankCardData.CardNumber)
			fmt.Printf("Expiry Date: %s\n", bankCardData.ExpiryDate)
			fmt.Printf("CVV: %s\n", bankCardData.CVV)
			fmt.Printf("Cardholder: %s\n", bankCardData.Cardholder)
			if bankCardData.Bank != "" {
				fmt.Printf("Bank: %s\n", bankCardData.Bank)
			}
			if bankCardData.Notes != "" {
				fmt.Printf("Notes: %s\n", bankCardData.Notes)
			}
		} else {
			fmt.Printf("Data: %s\n", string(decryptedData))
		}
	default:
		fmt.Printf("Data: %s\n", string(decryptedData))
	}

	return nil
}

// CleanQuotes removes quotes from string
func CleanQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
