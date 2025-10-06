package models

import (
	"time"

	"github.com/google/uuid"
)

// DataType defines the type of stored data
type DataType string

const (
	DataTypeLoginPassword DataType = "login_password"
	DataTypeText          DataType = "text"
	DataTypeBinary        DataType = "binary"
	DataTypeBankCard      DataType = "bank_card"
)

// Data represents user's private data
type Data struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Type        DataType  `json:"type" db:"type"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Data        []byte    `json:"data" db:"data"`
	Metadata    string    `json:"metadata" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DataRequest represents create/update data request
type DataRequest struct {
	Type        DataType `json:"type" validate:"required,oneof=login_password text binary bank_card"`
	Name        string   `json:"name" validate:"required,max=255"`
	Description string   `json:"description" validate:"max=1000"`
	Data        []byte   `json:"data" validate:"required"`
	Metadata    string   `json:"metadata" validate:"max=2000"`
}

// LoginPasswordData represents login/password data
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// BankCardData represents bank card data
type BankCardData struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Cardholder string `json:"cardholder"`
	Bank       string `json:"bank,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// TextData represents arbitrary text data
type TextData struct {
	Content string `json:"content"`
	Notes   string `json:"notes,omitempty"`
}

// BinaryData represents binary data
type BinaryData struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Notes    string `json:"notes,omitempty"`
}
