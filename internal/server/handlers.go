package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/auth"
	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type DataStorage interface {
	GetDataByID(ctx context.Context, dataID uuid.UUID) (*models.Data, error)
	GetDataByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Data, error)
	CreateData(ctx context.Context, data *models.Data) error
	UpdateData(ctx context.Context, data *models.Data) error
	DeleteData(ctx context.Context, dataID uuid.UUID) error
}

func RegisterRoutes(r *mux.Router, userStorage UserStorage, dataStorage DataStorage, jwtManager *auth.JWTManager) {
	r.HandleFunc("/api/v1/register", handleRegister(userStorage, jwtManager)).Methods("POST")
	r.HandleFunc("/api/v1/login", handleLogin(userStorage, jwtManager)).Methods("POST")

	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth.AuthMiddleware(jwtManager)(w, r, next.ServeHTTP)
		})
	})

	protected.HandleFunc("/data", handleGetData(dataStorage)).Methods("GET")
	protected.HandleFunc("/data", handleCreateData(dataStorage)).Methods("POST")
	protected.HandleFunc("/data/{id}", handleGetDataByID(dataStorage)).Methods("GET")
	protected.HandleFunc("/data/{id}", handleUpdateData(dataStorage)).Methods("PUT")
	protected.HandleFunc("/data/{id}", handleDeleteData(dataStorage)).Methods("DELETE")
}

func handleRegister(userStorage UserStorage, jwtManager *auth.JWTManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		cryptoManager, err := crypto.NewCryptoManager(req.MasterPassword)
		if err != nil {
			logger.Log.Error("Failed to create crypto manager", zap.Error(err))
			http.Error(w, "Failed to initialize encryption", http.StatusInternalServerError)
			return
		}

		hashedMasterPassword, err := bcrypt.GenerateFromPassword([]byte(req.MasterPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Log.Error("Failed to hash master password", zap.Error(err))
			http.Error(w, "Failed to hash master password", http.StatusInternalServerError)
			return
		}

		user := &models.User{
			ID:             uuid.New(),
			Username:       req.Username,
			Password:       string(hashedPassword),
			MasterPassword: string(hashedMasterPassword),
			Salt:           cryptoManager.GetSaltBase64(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := userStorage.CreateUser(r.Context(), user); err != nil {
			if err.Error() == "user already exists" {
				logger.Log.Warn("User already exists", zap.String("username", req.Username))
				http.Error(w, "User already exists", http.StatusConflict)
				return
			}
			logger.Log.Error("Failed to create user", zap.Error(err), zap.String("username", req.Username))
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		logger.Log.Info("User registered successfully", zap.String("username", req.Username), zap.String("user_id", user.ID.String()))

		token, err := jwtManager.GenerateToken(user.ID, user.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		response := models.AuthResponse{
			Token: token,
			User:  *user,
			Salt:  user.Salt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Log.Error("Failed to encode response", zap.Error(err))
		}
	}
}

func handleLogin(userStorage UserStorage, jwtManager *auth.JWTManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log.Warn("Invalid login request", zap.Error(err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logger.Log.Info("User login attempt", zap.String("username", req.Username))

		user, err := userStorage.GetUserByUsername(r.Context(), req.Username)
		if err != nil {
			if err.Error() == "user not found" {
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

		token, err := jwtManager.GenerateToken(user.ID, user.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		response := models.AuthResponse{
			Token: token,
			User:  *user,
			Salt:  user.Salt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Log.Error("Failed to encode response", zap.Error(err))
		}
	}
}

func handleGetData(dataStorage DataStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		data, err := dataStorage.GetDataByUserID(r.Context(), userID)
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
}

func handleCreateData(dataStorage DataStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := dataStorage.CreateData(r.Context(), data); err != nil {
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
}

func handleGetDataByID(dataStorage DataStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		data, err := dataStorage.GetDataByID(r.Context(), dataID)
		if err != nil {
			if err.Error() == "data not found" {
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
}

func handleUpdateData(dataStorage DataStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		data, err := dataStorage.GetDataByID(r.Context(), dataID)
		if err != nil {
			if err.Error() == "data not found" {
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

		if err := dataStorage.UpdateData(r.Context(), data); err != nil {
			http.Error(w, "Failed to update data", http.StatusInternalServerError)
			return
		}

		response := models.DataResponse{Data: *data}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Log.Error("Failed to encode response", zap.Error(err))
		}
	}
}

func handleDeleteData(dataStorage DataStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		data, err := dataStorage.GetDataByID(r.Context(), dataID)
		if err != nil {
			if err.Error() == "data not found" {
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

		if err := dataStorage.DeleteData(r.Context(), dataID); err != nil {
			http.Error(w, "Failed to delete data", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
