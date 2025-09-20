package server

import (
	"encoding/json"
	"net/http"
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

// UserReader defines read operations for users
type UserReader interface {
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(userID uuid.UUID) (*models.User, error)
}

// UserWriter defines write operations for users
type UserWriter interface {
	CreateUser(user *models.User) error
}

// UserStorage combines user read and write operations
type UserStorage interface {
	UserReader
	UserWriter
}

// DataReader defines read operations for data
type DataReader interface {
	GetDataByID(dataID uuid.UUID) (*models.Data, error)
	GetDataByUserID(userID uuid.UUID) ([]*models.Data, error)
}

// DataWriter defines write operations for data
type DataWriter interface {
	CreateData(data *models.Data) error
	UpdateData(data *models.Data) error
	DeleteData(dataID uuid.UUID) error
}

// DataStorage combines data read and write operations
type DataStorage interface {
	DataReader
	DataWriter
}

// Storage combines all storage operations
type Storage interface {
	UserStorage
	DataStorage
}

// Server represents HTTP server
type Server struct {
	storage    Storage
	jwtManager *auth.JWTManager
}

// NewServer creates new server
func NewServer(storage Storage, jwtManager *auth.JWTManager) *Server {
	return &Server{
		storage:    storage,
		jwtManager: jwtManager,
	}
}

// RegisterRoutes registers routes
func (s *Server) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/api/v1/login", s.handleLogin).Methods("POST")

	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth.AuthMiddleware(s.jwtManager)(w, r, next.ServeHTTP)
		})
	})

	protected.HandleFunc("/data", s.handleGetData).Methods("GET")
	protected.HandleFunc("/data", s.handleCreateData).Methods("POST")
	protected.HandleFunc("/data/{id}", s.handleGetDataByID).Methods("GET")
	protected.HandleFunc("/data/{id}", s.handleUpdateData).Methods("PUT")
	protected.HandleFunc("/data/{id}", s.handleDeleteData).Methods("DELETE")
}

// handleRegister handles user registration
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warn("Invalid registration request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Log.Info("User registration attempt", zap.String("username", req.Username))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.storage.CreateUser(user); err != nil {
		if err == storage.ErrUserExists {
			logger.Log.Warn("User already exists", zap.String("username", req.Username))
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		logger.Log.Error("Failed to create user", zap.Error(err), zap.String("username", req.Username))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	logger.Log.Info("User registered successfully", zap.String("username", req.Username), zap.String("user_id", user.ID.String()))

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleLogin handles user authentication
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warn("Invalid login request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Log.Info("User login attempt", zap.String("username", req.Username))

	user, err := s.storage.GetUserByUsername(req.Username)
	if err != nil {
		if err == storage.ErrUserNotFound {
			logger.Log.Warn("Login failed - user not found", zap.String("username", req.Username))
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		logger.Log.Error("Failed to get user", zap.Error(err), zap.String("username", req.Username))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Log.Warn("Login failed - invalid password", zap.String("username", req.Username))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	logger.Log.Info("User logged in successfully", zap.String("username", req.Username), zap.String("user_id", user.ID.String()))

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleGetData gets all user data
func (s *Server) handleGetData(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	data, err := s.storage.GetDataByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	response := models.DataListResponse{Data: make([]models.Data, len(data))}
	for i, d := range data {
		response.Data[i] = *d
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleCreateData creates new data
func (s *Server) handleCreateData(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.DataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Data:        req.Data,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.storage.CreateData(data); err != nil {
		http.Error(w, "Failed to create data", http.StatusInternalServerError)
		return
	}

	response := models.DataResponse{Data: *data}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleGetDataByID gets data by ID
func (s *Server) handleGetDataByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid data ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	data, err := s.storage.GetDataByID(dataID)
	if err != nil {
		if err == storage.ErrDataNotFound {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	if data.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	response := models.DataResponse{Data: *data}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleUpdateData updates data
func (s *Server) handleUpdateData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid data ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.DataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	data, err := s.storage.GetDataByID(dataID)
	if err != nil {
		if err == storage.ErrDataNotFound {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	if data.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	data.Type = req.Type
	data.Name = req.Name
	data.Description = req.Description
	data.Data = req.Data
	data.Metadata = req.Metadata
	data.UpdatedAt = time.Now()

	if err := s.storage.UpdateData(data); err != nil {
		http.Error(w, "Failed to update data", http.StatusInternalServerError)
		return
	}

	response := models.DataResponse{Data: *data}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

// handleDeleteData deletes data
func (s *Server) handleDeleteData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid data ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	data, err := s.storage.GetDataByID(dataID)
	if err != nil {
		if err == storage.ErrDataNotFound {
			http.Error(w, "Data not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	if data.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := s.storage.DeleteData(dataID); err != nil {
		http.Error(w, "Failed to delete data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
