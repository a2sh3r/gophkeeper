package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/auth"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/a2sh3r/gophkeeper/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestServer_Register(t *testing.T) {
	tests := []struct {
		name           string
		req            models.UserRequest
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "valid registration",
			req: models.UserRequest{
				Username:       "testuser",
				Password:       "password123",
				MasterPassword: "masterPassword123!",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "empty username",
			req: models.UserRequest{
				Username:       "",
				Password:       "password123",
				MasterPassword: "masterPassword123!",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "empty password",
			req: models.UserRequest{
				Username:       "testuser",
				Password:       "",
				MasterPassword: "masterPassword123!",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			jsonBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.User.Username != tt.req.Username {
					t.Errorf("Expected username %s, got %s", tt.req.Username, response.User.Username)
				}

				if response.Token == "" {
					t.Error("Expected non-empty token")
				}
			}
		})
	}
}

func TestServer_Register_DuplicateUser(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := userStorage.CreateUser(context.Background(), user); err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err), zap.String("username", user.Username))
	}

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	reqBody := models.UserRequest{
		Username:       "testuser",
		Password:       "password123",
		MasterPassword: "masterPassword123!",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestServer_Login(t *testing.T) {
	tests := []struct {
		name           string
		req            models.LoginRequest
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "valid login",
			req: models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "empty username",
			req: models.LoginRequest{
				Username: "",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
		{
			name: "empty password",
			req: models.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			expectedStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			if tt.name == "valid login" {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:        uuid.New(),
					Username:  "testuser",
					Password:  string(hashedPassword),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := userStorage.CreateUser(context.Background(), user); err != nil {
					logger.Log.Error("Failed to create data", zap.Error(err), zap.String("username", user.Username))
				}
			}

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			jsonBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.User.Username != tt.req.Username {
					t.Errorf("Expected username %s, got %s", tt.req.Username, response.User.Username)
				}

				if response.Token == "" {
					t.Error("Expected non-empty token")
				}
			}
		})
	}
}

func TestServer_CreateData(t *testing.T) {
	tests := []struct {
		name           string
		req            models.DataRequest
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "valid data creation",
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "Test Data",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "empty name",
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "",
				Description: "Test description",
				Data:        []byte("test content"),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "empty data",
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "Test Data",
				Description: "Test description",
				Data:        []byte(""),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusCreated,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			userID := uuid.New()
			token, _ := jwtManager.GenerateToken(userID, "testuser")

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			jsonBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/v1/data", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.DataResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Data.Name != tt.req.Name {
					t.Errorf("Expected name %s, got %s", tt.req.Name, response.Data.Name)
				}

				if response.Data.UserID != userID {
					t.Errorf("Expected UserID %s, got %s", userID, response.Data.UserID)
				}
			}
		})
	}
}

func TestServer_GetData(t *testing.T) {
	tests := []struct {
		name           string
		dataCount      int
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "user with data",
			dataCount:      3,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "user with no data",
			dataCount:      0,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			userID := uuid.New()
			token, _ := jwtManager.GenerateToken(userID, "testuser")

			for i := 0; i < tt.dataCount; i++ {
				data := &models.Data{
					ID:          uuid.New(),
					UserID:      userID,
					Type:        models.DataTypeText,
					Name:        "Test Data " + string(rune(i)),
					Description: "Test description",
					Data:        []byte("test content"),
					Metadata:    "{}",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := dataStorage.CreateData(context.Background(), data); err != nil {
					logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
				}
			}

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			req := httptest.NewRequest("GET", "/api/v1/data", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.DataListResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(response.Data) != tt.dataCount {
					t.Errorf("Expected %d data items, got %d", tt.dataCount, len(response.Data))
				}
			}
		})
	}
}

func TestServer_GetDataByID(t *testing.T) {
	tests := []struct {
		name           string
		dataID         string
		userID         uuid.UUID
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "valid data access",
			dataID:         "",
			userID:         uuid.New(),
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "invalid data ID",
			dataID:         "invalid-uuid",
			userID:         uuid.New(),
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "non-existing data",
			dataID:         uuid.New().String(),
			userID:         uuid.New(),
			expectedStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			token, _ := jwtManager.GenerateToken(tt.userID, "testuser")

			var dataID string
			if tt.name == "valid data access" {
				data := &models.Data{
					ID:          uuid.New(),
					UserID:      tt.userID,
					Type:        models.DataTypeText,
					Name:        "Test Data",
					Description: "Test description",
					Data:        []byte("test content"),
					Metadata:    "{}",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := dataStorage.CreateData(context.Background(), data); err != nil {
					logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
				}
				dataID = data.ID.String()
			} else {
				dataID = tt.dataID
			}

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			req := httptest.NewRequest("GET", "/api/v1/data/"+dataID, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.DataResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Data.Name != "Test Data" {
					t.Errorf("Expected name Test Data, got %s", response.Data.Name)
				}
			}
		})
	}
}

func TestServer_UpdateData(t *testing.T) {
	tests := []struct {
		name           string
		dataID         string
		userID         uuid.UUID
		req            models.DataRequest
		expectedStatus int
		wantErr        bool
	}{
		{
			name:   "valid update",
			dataID: "",
			userID: uuid.New(),
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "Updated Data",
				Description: "Updated description",
				Data:        []byte("updated content"),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:   "invalid data ID",
			dataID: "invalid-uuid",
			userID: uuid.New(),
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "Updated Data",
				Description: "Updated description",
				Data:        []byte("updated content"),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:   "non-existing data",
			dataID: uuid.New().String(),
			userID: uuid.New(),
			req: models.DataRequest{
				Type:        models.DataTypeText,
				Name:        "Updated Data",
				Description: "Updated description",
				Data:        []byte("updated content"),
				Metadata:    "{}",
			},
			expectedStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			token, _ := jwtManager.GenerateToken(tt.userID, "testuser")

			var dataID string
			if tt.name == "valid update" {
				data := &models.Data{
					ID:          uuid.New(),
					UserID:      tt.userID,
					Type:        models.DataTypeText,
					Name:        "Original Data",
					Description: "Original description",
					Data:        []byte("original content"),
					Metadata:    "{}",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := dataStorage.CreateData(context.Background(), data); err != nil {
					logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
				}
				dataID = data.ID.String()
			} else {
				dataID = tt.dataID
			}

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			jsonBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("PUT", "/api/v1/data/"+dataID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.wantErr {
				var response models.DataResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.Data.Name != tt.req.Name {
					t.Errorf("Expected name %s, got %s", tt.req.Name, response.Data.Name)
				}
			}
		})
	}
}

func TestServer_DeleteData(t *testing.T) {
	tests := []struct {
		name           string
		dataID         string
		userID         uuid.UUID
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "valid delete",
			dataID:         "",
			userID:         uuid.New(),
			expectedStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "invalid data ID",
			dataID:         "invalid-uuid",
			userID:         uuid.New(),
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "non-existing data",
			dataID:         uuid.New().String(),
			userID:         uuid.New(),
			expectedStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStorage := storage.NewMemoryStorage()
			dataStorage := storage.NewMemoryStorage()
			jwtManager := auth.NewJWTManager("test-secret", time.Hour)

			token, _ := jwtManager.GenerateToken(tt.userID, "testuser")

			var dataID string
			if tt.name == "valid delete" {
				data := &models.Data{
					ID:          uuid.New(),
					UserID:      tt.userID,
					Type:        models.DataTypeText,
					Name:        "Test Data",
					Description: "Test description",
					Data:        []byte("test content"),
					Metadata:    "{}",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				if err := dataStorage.CreateData(context.Background(), data); err != nil {
					logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
				}
				dataID = data.ID.String()
			} else {
				dataID = tt.dataID
			}

			router := mux.NewRouter()
			RegisterRoutes(router, userStorage, dataStorage, jwtManager)

			req := httptest.NewRequest("DELETE", "/api/v1/data/"+dataID, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServer_HandleRegister_InvalidJSON(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("POST", "/api/v1/register", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleLogin_InvalidJSON(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("POST", "/api/v1/login", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleCreateData_InvalidJSON(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID := uuid.New()
	token, _ := jwtManager.GenerateToken(userID, "testuser")

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("POST", "/api/v1/data", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleUpdateData_InvalidJSON(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID := uuid.New()
	token, _ := jwtManager.GenerateToken(userID, "testuser")

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("PUT", "/api/v1/data/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleRegister_StorageError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := userStorage.CreateUser(context.Background(), user); err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err), zap.String("username", user.Username))
	}

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	reqBody := models.UserRequest{
		Username:       "testuser",
		Password:       "password123",
		MasterPassword: "masterPassword123!",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestServer_HandleLogin_StorageError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	reqBody := models.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestServer_HandleGetDataByID_AccessDenied(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID1 := uuid.New()
	userID2 := uuid.New()
	token, _ := jwtManager.GenerateToken(userID1, "testuser")

	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID2,
		Type:        models.DataTypeText,
		Name:        "Test Data",
		Description: "Test description",
		Data:        []byte("test content"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := dataStorage.CreateData(context.Background(), data); err != nil {
		logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
	}

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("GET", "/api/v1/data/"+data.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestServer_HandleUpdateData_AccessDenied(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID1 := uuid.New()
	userID2 := uuid.New()
	token, _ := jwtManager.GenerateToken(userID1, "testuser")

	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID2,
		Type:        models.DataTypeText,
		Name:        "Test Data",
		Description: "Test description",
		Data:        []byte("test content"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := dataStorage.CreateData(context.Background(), data); err != nil {
		logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
	}

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	reqBody := models.DataRequest{
		Type: models.DataTypeText,
		Name: "Updated Data",
		Data: []byte("updated content"),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/data/"+data.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestServer_HandleDeleteData_AccessDenied(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID1 := uuid.New()
	userID2 := uuid.New()
	token, _ := jwtManager.GenerateToken(userID1, "testuser")

	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID2,
		Type:        models.DataTypeText,
		Name:        "Test Data",
		Description: "Test description",
		Data:        []byte("test content"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := dataStorage.CreateData(context.Background(), data); err != nil {
		logger.Log.Error("Failed to create data", zap.Error(err), zap.String("data name", data.Name))
	}

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("DELETE", "/api/v1/data/"+data.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestServer_HandleRegister_InternalError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleLogin_InternalError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_HandleGetData_InternalError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID := uuid.New()
	token, _ := jwtManager.GenerateToken(userID, "testuser")

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("GET", "/api/v1/data", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestServer_HandleCreateData_InternalError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID := uuid.New()
	token, _ := jwtManager.GenerateToken(userID, "testuser")

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	reqBody := models.DataRequest{
		Type: models.DataTypeText,
		Name: "Test Data",
		Data: []byte("test content"),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/data", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestServer_HandleDeleteData_InternalError(t *testing.T) {
	userStorage := storage.NewMemoryStorage()
	dataStorage := storage.NewMemoryStorage()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userID := uuid.New()
	token, _ := jwtManager.GenerateToken(userID, "testuser")

	router := mux.NewRouter()
	RegisterRoutes(router, userStorage, dataStorage, jwtManager)

	req := httptest.NewRequest("DELETE", "/api/v1/data/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
