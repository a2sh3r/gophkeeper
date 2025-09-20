package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestClient_SetToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "valid token",
			token: "test-token",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "long token",
			token: "very-long-token-with-many-characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("http://localhost:8080")
			client.SetToken(tt.token)

			if client.token != tt.token {
				t.Errorf("Expected token %s, got %s", tt.token, client.token)
			}
		})
	}
}

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		serverCode int
		serverResp models.AuthResponse
		wantErr    bool
	}{
		{
			name:       "successful registration",
			username:   "testuser",
			password:   "password123",
			serverCode: http.StatusOK,
			serverResp: models.AuthResponse{
				Token: "test-token",
				User: models.User{
					ID:       uuid.New(),
					Username: "testuser",
				},
			},
			wantErr: false,
		},
		{
			name:       "server error",
			username:   "testuser",
			password:   "password123",
			serverCode: http.StatusConflict,
			serverResp: models.AuthResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/register" {
					t.Errorf("Expected path /api/v1/register, got %s", r.URL.Path)
				}

				var req models.UserRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				if req.Username != tt.username {
					t.Errorf("Expected username %s, got %s", tt.username, req.Username)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.Register(tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp.Token != tt.serverResp.Token {
					t.Errorf("Expected token %s, got %s", tt.serverResp.Token, resp.Token)
				}
				if resp.User.Username != tt.serverResp.User.Username {
					t.Errorf("Expected username %s, got %s", tt.serverResp.User.Username, resp.User.Username)
				}
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		serverCode int
		serverResp models.AuthResponse
		wantErr    bool
	}{
		{
			name:       "successful login",
			username:   "testuser",
			password:   "password123",
			serverCode: http.StatusOK,
			serverResp: models.AuthResponse{
				Token: "test-token",
				User: models.User{
					ID:       uuid.New(),
					Username: "testuser",
				},
			},
			wantErr: false,
		},
		{
			name:       "invalid credentials",
			username:   "testuser",
			password:   "wrongpassword",
			serverCode: http.StatusUnauthorized,
			serverResp: models.AuthResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/login" {
					t.Errorf("Expected path /api/v1/login, got %s", r.URL.Path)
				}

				var req models.LoginRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				if req.Username != tt.username {
					t.Errorf("Expected username %s, got %s", tt.username, req.Username)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.Login(tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp.Token != tt.serverResp.Token {
					t.Errorf("Expected token %s, got %s", tt.serverResp.Token, resp.Token)
				}
				if resp.User.Username != tt.serverResp.User.Username {
					t.Errorf("Expected username %s, got %s", tt.serverResp.User.Username, resp.User.Username)
				}
			}
		})
	}
}

func TestClient_GetData(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		serverCode int
		serverResp models.DataListResponse
		wantErr    bool
	}{
		{
			name:       "successful get data",
			token:      "test-token",
			serverCode: http.StatusOK,
			serverResp: models.DataListResponse{
				Data: []models.Data{
					{
						ID:   uuid.New(),
						Type: models.DataTypeText,
						Name: "Test Data",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			token:      "invalid-token",
			serverCode: http.StatusUnauthorized,
			serverResp: models.DataListResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			data, err := client.GetData()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(data) != len(tt.serverResp.Data) {
					t.Errorf("Expected %d data items, got %d", len(tt.serverResp.Data), len(data))
				}
			}
		})
	}
}

func TestClient_CreateData(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataReq    models.DataRequest
		serverCode int
		serverResp models.DataResponse
		wantErr    bool
	}{
		{
			name:  "successful create data",
			token: "test-token",
			dataReq: models.DataRequest{
				Type: models.DataTypeText,
				Name: "Test Data",
				Data: []byte("test content"),
			},
			serverCode: http.StatusCreated,
			serverResp: models.DataResponse{
				Data: models.Data{
					ID:   uuid.New(),
					Type: models.DataTypeText,
					Name: "Test Data",
				},
			},
			wantErr: false,
		},
		{
			name:  "unauthorized",
			token: "invalid-token",
			dataReq: models.DataRequest{
				Type: models.DataTypeText,
				Name: "Test Data",
				Data: []byte("test content"),
			},
			serverCode: http.StatusUnauthorized,
			serverResp: models.DataResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				if r.Method != "POST" {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				var req models.DataRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				if req.Name != tt.dataReq.Name {
					t.Errorf("Expected name %s, got %s", tt.dataReq.Name, req.Name)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			data, err := client.CreateData(tt.dataReq)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if data.Name != tt.serverResp.Data.Name {
					t.Errorf("Expected name %s, got %s", tt.serverResp.Data.Name, data.Name)
				}
			}
		})
	}
}

func TestClient_GetDataByID(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataID     string
		serverCode int
		serverResp models.DataResponse
		wantErr    bool
	}{
		{
			name:       "successful get data by ID",
			token:      "test-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusOK,
			serverResp: models.DataResponse{
				Data: models.Data{
					ID:   uuid.New(),
					Type: models.DataTypeText,
					Name: "Test Data",
				},
			},
			wantErr: false,
		},
		{
			name:       "data not found",
			token:      "test-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusNotFound,
			serverResp: models.DataResponse{},
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			token:      "invalid-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusUnauthorized,
			serverResp: models.DataResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/data/" + tt.dataID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			data, err := client.GetDataByID(tt.dataID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if data.Name != tt.serverResp.Data.Name {
					t.Errorf("Expected name %s, got %s", tt.serverResp.Data.Name, data.Name)
				}
			}
		})
	}
}

func TestClient_UpdateData(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataID     string
		dataReq    models.DataRequest
		serverCode int
		serverResp models.DataResponse
		wantErr    bool
	}{
		{
			name:   "successful update data",
			token:  "test-token",
			dataID: uuid.New().String(),
			dataReq: models.DataRequest{
				Type: models.DataTypeText,
				Name: "Updated Data",
				Data: []byte("updated content"),
			},
			serverCode: http.StatusOK,
			serverResp: models.DataResponse{
				Data: models.Data{
					ID:   uuid.New(),
					Type: models.DataTypeText,
					Name: "Updated Data",
				},
			},
			wantErr: false,
		},
		{
			name:   "data not found",
			token:  "test-token",
			dataID: uuid.New().String(),
			dataReq: models.DataRequest{
				Type: models.DataTypeText,
				Name: "Updated Data",
				Data: []byte("updated content"),
			},
			serverCode: http.StatusNotFound,
			serverResp: models.DataResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/data/" + tt.dataID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				if r.Method != "PUT" {
					t.Errorf("Expected method PUT, got %s", r.Method)
				}

				var req models.DataRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				if req.Name != tt.dataReq.Name {
					t.Errorf("Expected name %s, got %s", tt.dataReq.Name, req.Name)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
					logger.Log.Error("Failed to encode response", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			data, err := client.UpdateData(tt.dataID, tt.dataReq)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if data.Name != tt.serverResp.Data.Name {
					t.Errorf("Expected name %s, got %s", tt.serverResp.Data.Name, data.Name)
				}
			}
		})
	}
}

func TestClient_DeleteData(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataID     string
		serverCode int
		wantErr    bool
	}{
		{
			name:       "successful delete data",
			token:      "test-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "data not found",
			token:      "test-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			token:      "invalid-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "server error",
			token:      "test-token",
			dataID:     uuid.New().String(),
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "bad request",
			token:      "test-token",
			dataID:     "invalid-uuid",
			serverCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/data/" + tt.dataID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				if r.Method != "DELETE" {
					t.Errorf("Expected method DELETE, got %s", r.Method)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.serverCode)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			err := client.DeleteData(tt.dataID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_authRequest_ServerError(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		req        interface{}
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:       "server error with JSON response",
			endpoint:   "/api/v1/register",
			req:        models.UserRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusInternalServerError,
			serverResp: `{"error": "internal server error"}`,
			wantErr:    true,
		},
		{
			name:       "server error with plain text response",
			endpoint:   "/api/v1/login",
			req:        models.LoginRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusBadRequest,
			serverResp: "bad request",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverCode)
				if _, err := w.Write([]byte(tt.serverResp)); err != nil {
					logger.Log.Error("Failed to write header", zap.Error(err))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			_, err := client.authRequest(tt.endpoint, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("authRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_authRequest_Additional(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		req        interface{}
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:       "successful auth request",
			endpoint:   "/api/v1/register",
			req:        models.UserRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusOK,
			serverResp: `{"token": "test-token", "user": {"id": "550e8400-e29b-41d4-a716-446655440000", "username": "test"}}`,
			wantErr:    false,
		},
		{
			name:       "created status code",
			endpoint:   "/api/v1/register",
			req:        models.UserRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusCreated,
			serverResp: `{"token": "test-token", "user": {"id": "550e8400-e29b-41d4-a716-446655440000", "username": "test"}}`,
			wantErr:    false,
		},
		{
			name:       "invalid JSON response",
			endpoint:   "/api/v1/login",
			req:        models.LoginRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusOK,
			serverResp: `{"invalid": json}`,
			wantErr:    true,
		},
		{
			name:       "empty response body",
			endpoint:   "/api/v1/login",
			req:        models.LoginRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusOK,
			serverResp: ``,
			wantErr:    true,
		},
		{
			name:       "unauthorized with error response",
			endpoint:   "/api/v1/login",
			req:        models.LoginRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusUnauthorized,
			serverResp: `{"error": "invalid credentials"}`,
			wantErr:    true,
		},
		{
			name:       "forbidden with error response",
			endpoint:   "/api/v1/login",
			req:        models.LoginRequest{Username: "test", Password: "pass"},
			serverCode: http.StatusForbidden,
			serverResp: `{"error": "access denied"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.endpoint {
					t.Errorf("Expected path %s, got %s", tt.endpoint, r.URL.Path)
				}

				if r.Method != "POST" {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				w.WriteHeader(tt.serverCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			client := NewClient(server.URL)

			_, err := client.authRequest(tt.endpoint, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("authRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetData_Additional(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:       "successful get data",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": [{"id": "550e8400-e29b-41d4-a716-446655440000", "type": "text", "name": "test"}]}`,
			wantErr:    false,
		},
		{
			name:       "empty data list",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": []}`,
			wantErr:    false,
		},
		{
			name:       "invalid JSON response",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": [invalid json]}`,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			token:      "invalid-token",
			serverCode: http.StatusUnauthorized,
			serverResp: `{"error": "unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "forbidden",
			token:      "expired-token",
			serverCode: http.StatusForbidden,
			serverResp: `{"error": "forbidden"}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			token:      "valid-token",
			serverCode: http.StatusInternalServerError,
			serverResp: `{"error": "internal server error"}`,
			wantErr:    true,
		},
		{
			name:       "bad gateway",
			token:      "valid-token",
			serverCode: http.StatusBadGateway,
			serverResp: `{"error": "bad gateway"}`,
			wantErr:    true,
		},
		{
			name:       "service unavailable",
			token:      "valid-token",
			serverCode: http.StatusServiceUnavailable,
			serverResp: `{"error": "service unavailable"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				if r.Method != "GET" {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.serverCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			_, err := client.GetData()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreateData_Additional(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataReq    models.DataRequest
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:  "successful create data",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "test data",
				Description: "test description",
				Data:        []byte("test content"),
				Metadata:    "test metadata",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440000", "type": "text", "name": "test data"}}`,
			wantErr:    false,
		},
		{
			name:  "create login password data",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "login_password",
				Name:        "GitHub",
				Description: "Main GitHub account",
				Data:        []byte(`{"login": "user", "password": "pass"}`),
				Metadata:    "github.com",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440001", "type": "login_password", "name": "GitHub"}}`,
			wantErr:    false,
		},
		{
			name:  "create bank card data",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "bank_card",
				Name:        "Visa Card",
				Description: "Main credit card",
				Data:        []byte(`{"number": "1234567890123456", "expiry": "12/25", "cvv": "123"}`),
				Metadata:    "Bank of America",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440002", "type": "bank_card", "name": "Visa Card"}}`,
			wantErr:    false,
		},
		{
			name:  "create binary data",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "binary",
				Name:        "Document",
				Description: "Important document",
				Data:        []byte("base64encodeddata"),
				Metadata:    "pdf",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440003", "type": "binary", "name": "Document"}}`,
			wantErr:    false,
		},
		{
			name:  "invalid JSON response",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "test",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {invalid json}}`,
			wantErr:    true,
		},
		{
			name:  "unauthorized",
			token: "invalid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "test",
			},
			serverCode: http.StatusUnauthorized,
			serverResp: `{"error": "unauthorized"}`,
			wantErr:    true,
		},
		{
			name:  "forbidden",
			token: "expired-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "test",
			},
			serverCode: http.StatusForbidden,
			serverResp: `{"error": "forbidden"}`,
			wantErr:    true,
		},
		{
			name:  "bad request",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "invalid_type",
				Name: "test",
			},
			serverCode: http.StatusBadRequest,
			serverResp: `{"error": "invalid data type"}`,
			wantErr:    true,
		},
		{
			name:  "conflict",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "duplicate",
			},
			serverCode: http.StatusConflict,
			serverResp: `{"error": "data already exists"}`,
			wantErr:    true,
		},
		{
			name:  "server error",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "test",
			},
			serverCode: http.StatusInternalServerError,
			serverResp: `{"error": "internal server error"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				if r.Method != "POST" {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.serverCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			_, err := client.CreateData(tt.dataReq)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetData_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:       "large data response",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": [
				{"id": "550e8400-e29b-41d4-a716-446655440000", "type": "text", "name": "item1", "description": "desc1", "data": "ZGF0YTE=", "metadata": "meta1", "created_at": "2023-01-01T00:00:00Z", "updated_at": "2023-01-01T00:00:00Z"},
				{"id": "550e8400-e29b-41d4-a716-446655440001", "type": "login_password", "name": "item2", "description": "desc2", "data": "ZGF0YTI=", "metadata": "meta2", "created_at": "2023-01-02T00:00:00Z", "updated_at": "2023-01-02T00:00:00Z"},
				{"id": "550e8400-e29b-41d4-a716-446655440002", "type": "bank_card", "name": "item3", "description": "desc3", "data": "ZGF0YTM=", "metadata": "meta3", "created_at": "2023-01-03T00:00:00Z", "updated_at": "2023-01-03T00:00:00Z"},
				{"id": "550e8400-e29b-41d4-a716-446655440003", "type": "binary", "name": "item4", "description": "desc4", "data": "ZGF0YTQ=", "metadata": "meta4", "created_at": "2023-01-04T00:00:00Z", "updated_at": "2023-01-04T00:00:00Z"}
			]}`,
			wantErr: false,
		},
		{
			name:       "malformed JSON array",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": [{"id": "550e8400-e29b-41d4-a716-446655440000", "type": "text", "name": "test", "missing": "field"}]}`,
			wantErr:    false, // This will actually succeed as missing fields are optional
		},
		{
			name:       "null data field",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"data": null}`,
			wantErr:    false, // This will succeed as null is valid
		},
		{
			name:       "missing data field",
			token:      "valid-token",
			serverCode: http.StatusOK,
			serverResp: `{"items": []}`,
			wantErr:    false, // This will succeed as the response is valid JSON
		},
		{
			name:       "gateway timeout",
			token:      "valid-token",
			serverCode: http.StatusGatewayTimeout,
			serverResp: `{"error": "gateway timeout"}`,
			wantErr:    true,
		},
		{
			name:       "request timeout",
			token:      "valid-token",
			serverCode: http.StatusRequestTimeout,
			serverResp: `{"error": "request timeout"}`,
			wantErr:    true,
		},
		{
			name:       "too many requests",
			token:      "valid-token",
			serverCode: http.StatusTooManyRequests,
			serverResp: `{"error": "rate limit exceeded"}`,
			wantErr:    true,
		},
		{
			name:       "not implemented",
			token:      "valid-token",
			serverCode: http.StatusNotImplemented,
			serverResp: `{"error": "not implemented"}`,
			wantErr:    true,
		},
		{
			name:       "bad gateway",
			token:      "valid-token",
			serverCode: http.StatusBadGateway,
			serverResp: `{"error": "bad gateway"}`,
			wantErr:    true,
		},
		{
			name:       "service unavailable",
			token:      "valid-token",
			serverCode: http.StatusServiceUnavailable,
			serverResp: `{"error": "service unavailable"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				if r.Method != "GET" {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.serverCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			_, err := client.GetData()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreateData_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		dataReq    models.DataRequest
		serverCode int
		serverResp string
		wantErr    bool
	}{
		{
			name:  "create data with all fields",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "Complete Test Data",
				Description: "This is a complete test with all fields filled",
				Data:        []byte("This is the actual data content"),
				Metadata:    "Additional metadata information",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440000", "type": "text", "name": "Complete Test Data", "description": "This is a complete test with all fields filled", "data": "VGhpcyBpcyB0aGUgYWN0dWFsIGRhdGEgY29udGVudA==", "metadata": "Additional metadata information", "created_at": "2023-01-01T00:00:00Z", "updated_at": "2023-01-01T00:00:00Z"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with minimal fields",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Minimal",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440001", "type": "text", "name": "Minimal"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with empty name",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440002", "type": "text", "name": ""}}`,
			wantErr:    false,
		},
		{
			name:  "create data with empty description",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "No Description",
				Description: "",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440003", "type": "text", "name": "No Description", "description": ""}}`,
			wantErr:    false,
		},
		{
			name:  "create data with empty metadata",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:     "text",
				Name:     "No Metadata",
				Metadata: "",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440004", "type": "text", "name": "No Metadata", "metadata": ""}}`,
			wantErr:    false,
		},
		{
			name:  "create data with empty data field",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Empty Data",
				Data: []byte(""),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440005", "type": "text", "name": "Empty Data", "data": ""}}`,
			wantErr:    false,
		},
		{
			name:  "create data with large data field",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Large Data",
				Data: []byte(strings.Repeat("This is a very long string that contains a lot of data. ", 100)),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440006", "type": "text", "name": "Large Data"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with special characters in name",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Special Chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440007", "type": "text", "name": "Special Chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with unicode characters",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "Unicode: 你好世界 🌍",
				Description: "Description with émojis and ñ characters",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440008", "type": "text", "name": "Unicode: 你好世界 🌍"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with JSON data field",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "JSON Data",
				Data: []byte(`{"nested": {"object": {"with": ["array", "of", "values"]}}}`),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440009", "type": "text", "name": "JSON Data"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with binary data",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "binary",
				Name: "Binary File",
				Data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440010", "type": "binary", "name": "Binary File"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with very long name",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: strings.Repeat("Very Long Name ", 50),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440011", "type": "text", "name": "Very Long Name"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with very long description",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "Long Description",
				Description: strings.Repeat("This is a very long description that goes on and on. ", 20),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440012", "type": "text", "name": "Long Description"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with very long metadata",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:     "text",
				Name:     "Long Metadata",
				Metadata: strings.Repeat("This is metadata information that is very long. ", 30),
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440013", "type": "text", "name": "Long Metadata"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with all data types",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "login_password",
				Name:        "All Types Test",
				Description: "Testing all supported data types",
				Data:        []byte(`{"login": "user@example.com", "password": "securePassword123!", "otp": "123456"}`),
				Metadata:    "github.com,google.com,facebook.com",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440014", "type": "login_password", "name": "All Types Test"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with bank card details",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "bank_card",
				Name:        "Credit Card",
				Description: "Main credit card",
				Data:        []byte(`{"number": "4111 1111 1111 1111", "expiry": "12/25", "cvv": "123", "holder": "John Doe", "bank": "Chase Bank"}`),
				Metadata:    "Chase Bank, Credit Card, Primary",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440015", "type": "bank_card", "name": "Credit Card"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with OTP information",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type:        "text",
				Name:        "OTP Codes",
				Description: "One-time passwords for various services",
				Data:        []byte(`{"google": "123456", "github": "789012", "facebook": "345678", "expires": "2023-12-31T23:59:59Z"}`),
				Metadata:    "OTP, 2FA, Authentication",
			},
			serverCode: http.StatusCreated,
			serverResp: `{"data": {"id": "550e8400-e29b-41d4-a716-446655440016", "type": "text", "name": "OTP Codes"}}`,
			wantErr:    false,
		},
		{
			name:  "create data with network error",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Network Error Test",
			},
			serverCode: http.StatusInternalServerError,
			serverResp: `{"error": "network error"}`,
			wantErr:    true,
		},
		{
			name:  "create data with database error",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Database Error Test",
			},
			serverCode: http.StatusInternalServerError,
			serverResp: `{"error": "database connection failed"}`,
			wantErr:    true,
		},
		{
			name:  "create data with validation error",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "invalid_type",
				Name: "Validation Error Test",
			},
			serverCode: http.StatusBadRequest,
			serverResp: `{"error": "invalid data type: invalid_type"}`,
			wantErr:    true,
		},
		{
			name:  "create data with duplicate name error",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Duplicate Name",
			},
			serverCode: http.StatusConflict,
			serverResp: `{"error": "data with this name already exists"}`,
			wantErr:    true,
		},
		{
			name:  "create data with quota exceeded",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Quota Exceeded Test",
			},
			serverCode: http.StatusTooManyRequests,
			serverResp: `{"error": "storage quota exceeded"}`,
			wantErr:    true,
		},
		{
			name:  "create data with maintenance mode",
			token: "valid-token",
			dataReq: models.DataRequest{
				Type: "text",
				Name: "Maintenance Test",
			},
			serverCode: http.StatusServiceUnavailable,
			serverResp: `{"error": "service under maintenance"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/data" {
					t.Errorf("Expected path /api/v1/data, got %s", r.URL.Path)
				}

				if r.Method != "POST" {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				expectedAuth := "Bearer " + tt.token
				if r.Header.Get("Authorization") != expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuth, r.Header.Get("Authorization"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.serverCode)
				_, _ = w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			client.SetToken(tt.token)

			_, err := client.CreateData(tt.dataReq)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
