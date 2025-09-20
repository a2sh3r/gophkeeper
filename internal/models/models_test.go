package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser(t *testing.T) {
	tests := []struct {
		name string
		user User
	}{
		{
			name: "valid user",
			user: User{
				ID:        uuid.New(),
				Username:  "testuser",
				Password:  "hashedpassword",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "user with empty username",
			user: User{
				ID:        uuid.New(),
				Username:  "",
				Password:  "hashedpassword",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user.ID == uuid.Nil {
				t.Error("User ID should not be nil")
			}
			if tt.user.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if tt.user.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should not be zero")
			}
		})
	}
}

func TestUserRequest(t *testing.T) {
	tests := []struct {
		name string
		req  UserRequest
	}{
		{
			name: "valid request",
			req: UserRequest{
				Username: "testuser",
				Password: "password123",
			},
		},
		{
			name: "request with empty username",
			req: UserRequest{
				Username: "",
				Password: "password123",
			},
		},
		{
			name: "request with empty password",
			req: UserRequest{
				Username: "testuser",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req.Username == "" && tt.req.Password == "" {
				t.Log("Both username and password are empty")
			}
		})
	}
}

func TestLoginRequest(t *testing.T) {
	tests := []struct {
		name string
		req  LoginRequest
	}{
		{
			name: "valid request",
			req: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
		},
		{
			name: "request with empty username",
			req: LoginRequest{
				Username: "",
				Password: "password123",
			},
		},
		{
			name: "request with empty password",
			req: LoginRequest{
				Username: "testuser",
				Password: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req.Username == "" && tt.req.Password == "" {
				t.Log("Both username and password are empty")
			}
		})
	}
}

func TestAuthResponse(t *testing.T) {
	tests := []struct {
		name string
		resp AuthResponse
	}{
		{
			name: "valid response",
			resp: AuthResponse{
				Token: "jwt-token",
				User: User{
					ID:        uuid.New(),
					Username:  "testuser",
					Password:  "hashedpassword",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		},
		{
			name: "response with empty token",
			resp: AuthResponse{
				Token: "",
				User: User{
					ID:        uuid.New(),
					Username:  "testuser",
					Password:  "hashedpassword",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.Token == "" {
				t.Log("Token is empty")
			}
			if tt.resp.User.ID == uuid.Nil {
				t.Error("User ID should not be nil")
			}
		})
	}
}

func TestData(t *testing.T) {
	tests := []struct {
		name string
		data Data
	}{
		{
			name: "valid data",
			data: Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        DataTypeLoginPassword,
				Name:        "Test Data",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		{
			name: "data with empty name",
			data: Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        DataTypeText,
				Name:        "",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		{
			name: "data with empty description",
			data: Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        DataTypeBinary,
				Name:        "Test Data",
				Description: "",
				Data:        []byte("test content"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.ID == uuid.Nil {
				t.Error("Data ID should not be nil")
			}
			if tt.data.UserID == uuid.Nil {
				t.Error("UserID should not be nil")
			}
			if tt.data.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if tt.data.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should not be zero")
			}
		})
	}
}

func TestDataRequest(t *testing.T) {
	tests := []struct {
		name string
		req  DataRequest
	}{
		{
			name: "valid request",
			req: DataRequest{
				Type:        DataTypeLoginPassword,
				Name:        "Test Data",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
			},
		},
		{
			name: "request with empty name",
			req: DataRequest{
				Type:        DataTypeText,
				Name:        "",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
			},
		},
		{
			name: "request with empty data",
			req: DataRequest{
				Type:        DataTypeBinary,
				Name:        "Test Data",
				Description: "Test description",
				Data:        []byte(""),
				Metadata:    "{}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req.Name == "" {
				t.Log("Name is empty")
			}
			if len(tt.req.Data) == 0 {
				t.Log("Data is empty")
			}
		})
	}
}

func TestDataType(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
	}{
		{
			name:     "login password type",
			dataType: DataTypeLoginPassword,
		},
		{
			name:     "text type",
			dataType: DataTypeText,
		},
		{
			name:     "binary type",
			dataType: DataTypeBinary,
		},
		{
			name:     "bank card type",
			dataType: DataTypeBankCard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dataType == "" {
				t.Error("DataType should not be empty")
			}
		})
	}
}

func TestLoginPasswordData(t *testing.T) {
	tests := []struct {
		name string
		data LoginPasswordData
	}{
		{
			name: "valid login password data",
			data: LoginPasswordData{
				Login:    "testuser",
				Password: "password123",
				URL:      "https://example.com",
				Notes:    "Test notes",
			},
		},
		{
			name: "data with empty fields",
			data: LoginPasswordData{
				Login:    "",
				Password: "",
				URL:      "",
				Notes:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.Login == "" && tt.data.Password == "" {
				t.Log("Both login and password are empty")
			}
		})
	}
}

func TestBankCardData(t *testing.T) {
	tests := []struct {
		name string
		data BankCardData
	}{
		{
			name: "valid bank card data",
			data: BankCardData{
				CardNumber: "4111111111111111",
				ExpiryDate: "12/25",
				CVV:        "123",
				Cardholder: "John Doe",
				Bank:       "Chase Bank",
				Notes:      "Main credit card",
			},
		},
		{
			name: "data with empty fields",
			data: BankCardData{
				CardNumber: "",
				ExpiryDate: "",
				CVV:        "",
				Cardholder: "",
				Bank:       "",
				Notes:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.CardNumber == "" && tt.data.CVV == "" {
				t.Log("Both card number and CVV are empty")
			}
		})
	}
}

func TestTextData(t *testing.T) {
	tests := []struct {
		name string
		data TextData
	}{
		{
			name: "valid text data",
			data: TextData{
				Content: "This is test content",
				Notes:   "Test notes",
			},
		},
		{
			name: "data with empty content",
			data: TextData{
				Content: "",
				Notes:   "Test notes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.Content == "" {
				t.Log("Content is empty")
			}
		})
	}
}

func TestBinaryData(t *testing.T) {
	tests := []struct {
		name string
		data BinaryData
	}{
		{
			name: "valid binary data",
			data: BinaryData{
				FileName: "document.pdf",
				MimeType: "application/pdf",
				Size:     1024,
				Notes:    "Important document",
			},
		},
		{
			name: "data with empty filename",
			data: BinaryData{
				FileName: "",
				MimeType: "application/pdf",
				Size:     0,
				Notes:    "Important document",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.FileName == "" {
				t.Log("FileName is empty")
			}
			if tt.data.Size == 0 {
				t.Log("Size is zero")
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name string
		resp ErrorResponse
	}{
		{
			name: "error with message",
			resp: ErrorResponse{
				Error:   "validation error",
				Message: "Invalid input",
			},
		},
		{
			name: "error without message",
			resp: ErrorResponse{
				Error:   "internal error",
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.Error == "" {
				t.Error("Error should not be empty")
			}
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		name string
		resp SuccessResponse
	}{
		{
			name: "success with message",
			resp: SuccessResponse{
				Message: "Operation completed successfully",
			},
		},
		{
			name: "success without message",
			resp: SuccessResponse{
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.Message == "" {
				t.Log("Message is empty")
			}
		})
	}
}

func TestDataListResponse(t *testing.T) {
	tests := []struct {
		name string
		resp DataListResponse
	}{
		{
			name: "response with data",
			resp: DataListResponse{
				Data: []Data{
					{
						ID:   uuid.New(),
						Type: DataTypeText,
						Name: "Test Data",
					},
				},
			},
		},
		{
			name: "response with empty data",
			resp: DataListResponse{
				Data: []Data{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.resp.Data) == 0 {
				t.Log("Data list is empty")
			}
		})
	}
}

func TestDataResponse(t *testing.T) {
	tests := []struct {
		name string
		resp DataResponse
	}{
		{
			name: "response with data",
			resp: DataResponse{
				Data: Data{
					ID:   uuid.New(),
					Type: DataTypeText,
					Name: "Test Data",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.Data.ID == uuid.Nil {
				t.Error("Data ID should not be nil")
			}
		})
	}
}
