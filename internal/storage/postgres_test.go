package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	return db, mock
}

func TestNewPostgresStorage(t *testing.T) {
	tests := []struct {
		name string
		db   *sql.DB
	}{
		{
			name: "nil database",
			db:   nil,
		},
		{
			name: "valid database",
			db:   &sql.DB{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewPostgresStorage(tt.db)

			if storage == nil {
				t.Error("NewPostgresStorage() returned nil")
				return
			}

			if storage.db != tt.db {
				t.Errorf("NewPostgresStorage() db = %v, want %v", storage.db, tt.db)
			}
		})
	}
}

func TestPostgresStorage_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      *models.User
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful user creation",
			user: &models.User{
				ID:             uuid.New(),
				Username:       "testuser",
				Password:       "hashedpassword",
				MasterPassword: "hashedmasterpassword",
				Salt:           "salt123",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(sqlmock.AnyArg(), "testuser", "hashedpassword", "hashedmasterpassword", "salt123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantError: false,
		},
		{
			name: "duplicate user error",
			user: &models.User{
				ID:             uuid.New(),
				Username:       "existinguser",
				Password:       "hashedpassword",
				MasterPassword: "hashedmasterpassword",
				Salt:           "salt123",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(sqlmock.AnyArg(), "existinguser", "hashedpassword", "hashedmasterpassword", "salt123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(fmt.Errorf(`duplicate key value violates unique constraint "users_username_key"`))
			},
			wantError: true,
		},
		{
			name: "database error",
			user: &models.User{
				ID:             uuid.New(),
				Username:       "testuser",
				Password:       "hashedpassword",
				MasterPassword: "hashedmasterpassword",
				Salt:           "salt123",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(sqlmock.AnyArg(), "testuser", "hashedpassword", "hashedmasterpassword", "salt123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			err = storage.CreateUser(context.Background(), tt.user)

			if (err != nil) != tt.wantError {
				t.Errorf("CreateUser() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.name == "duplicate user error" && err != nil && err != ErrUserExists {
				t.Errorf("CreateUser() expected ErrUserExists, got %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_GetUserByUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:     "successful user retrieval",
			username: "testuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "master_password", "salt", "created_at", "updated_at"}).
					AddRow(uuid.New(), "testuser", "hashedpassword", "hashedmasterpassword", "salt123", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			wantError: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantError: true,
		},
		{
			name:     "database error",
			username: "testuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE username = \\$1").
					WithArgs("testuser").
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			user, err := storage.GetUserByUsername(context.Background(), tt.username)

			if (err != nil) != tt.wantError {
				t.Errorf("GetUserByUsername() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && user == nil {
				t.Error("GetUserByUsername() returned nil user")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_GetUserByID(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:   "successful user retrieval",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "master_password", "salt", "created_at", "updated_at"}).
					AddRow(userID, "testuser", "hashedpassword", "hashedmasterpassword", "salt123", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE id = \\$1").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantError: false,
		},
		{
			name:   "user not found",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE id = \\$1").
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			wantError: true,
		},
		{
			name:   "database error",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE id = \\$1").
					WithArgs(userID).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			user, err := storage.GetUserByID(context.Background(), tt.userID)

			if (err != nil) != tt.wantError {
				t.Errorf("GetUserByID() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && user == nil {
				t.Error("GetUserByID() returned nil user")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_CreateData(t *testing.T) {
	tests := []struct {
		name      string
		data      *models.Data
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful data creation",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        models.DataTypeText,
				Name:        "test data",
				Description: "test description",
				Data:        []byte("test content"),
				Metadata:    "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO data").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "text", "test data", "test description", []byte("test content"), "", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantError: false,
		},
		{
			name: "database error",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        models.DataTypeLoginPassword,
				Name:        "login data",
				Description: "login description",
				Data:        []byte("username:password"),
				Metadata:    "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO data").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "login_password", "login data", "login description", []byte("username:password"), "", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			err := storage.CreateData(context.Background(), tt.data)

			if (err != nil) != tt.wantError {
				t.Errorf("CreateData() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_GetDataByID(t *testing.T) {
	dataID := uuid.New()
	tests := []struct {
		name      string
		dataID    uuid.UUID
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:   "successful data retrieval",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "name", "description", "data", "metadata", "created_at", "updated_at"}).
					AddRow(dataID, uuid.New(), "text", "test data", "test description", []byte("test content"), "", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(dataID).
					WillReturnRows(rows)
			},
			wantError: false,
		},
		{
			name:   "data not found",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(dataID).
					WillReturnError(sql.ErrNoRows)
			},
			wantError: true,
		},
		{
			name:   "database error",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(dataID).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			data, err := storage.GetDataByID(context.Background(), tt.dataID)

			if (err != nil) != tt.wantError {
				t.Errorf("GetDataByID() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && data == nil {
				t.Error("GetDataByID() returned nil data")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_GetDataByUserID(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:   "successful data list retrieval",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "name", "description", "data", "metadata", "created_at", "updated_at"}).
					AddRow(uuid.New(), userID, "text", "test data 1", "description 1", []byte("content 1"), "", time.Now(), time.Now()).
					AddRow(uuid.New(), userID, "login_password", "test data 2", "description 2", []byte("content 2"), "", time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantError: false,
		},
		{
			name:   "no data found",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "type", "name", "description", "data", "metadata", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantError: false,
		},
		{
			name:   "database error",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at").
					WithArgs(userID).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			dataList, err := storage.GetDataByUserID(context.Background(), tt.userID)

			if (err != nil) != tt.wantError {
				t.Errorf("GetDataByUserID() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && dataList == nil && tt.name != "no data found" {
				t.Error("GetDataByUserID() returned nil data list")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_UpdateData(t *testing.T) {
	tests := []struct {
		name      string
		data      *models.Data
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful data update",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        models.DataTypeText,
				Name:        "updated data",
				Description: "updated description",
				Data:        []byte("updated content"),
				Metadata:    "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE data SET").
					WithArgs(sqlmock.AnyArg(), "text", "updated data", "updated description", []byte("updated content"), "", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantError: false,
		},
		{
			name: "data not found",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        models.DataTypeBankCard,
				Name:        "bank card",
				Description: "credit card",
				Data:        []byte("card number"),
				Metadata:    "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE data SET").
					WithArgs(sqlmock.AnyArg(), "bank_card", "bank card", "credit card", []byte("card number"), "", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantError: true,
		},
		{
			name: "database error",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Type:        models.DataTypeText,
				Name:        "test data",
				Description: "test description",
				Data:        []byte("test content"),
				Metadata:    "",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE data SET").
					WithArgs(sqlmock.AnyArg(), "text", "test data", "test description", []byte("test content"), "", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			err := storage.UpdateData(context.Background(), tt.data)

			if (err != nil) != tt.wantError {
				t.Errorf("UpdateData() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresStorage_DeleteData(t *testing.T) {
	dataID := uuid.New()
	tests := []struct {
		name      string
		dataID    uuid.UUID
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name:   "successful data deletion",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM data WHERE id = \\$1").
					WithArgs(dataID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantError: false,
		},
		{
			name:   "data not found",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM data WHERE id = \\$1").
					WithArgs(dataID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantError: true,
		},
		{
			name:   "database error",
			dataID: dataID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM data WHERE id = \\$1").
					WithArgs(dataID).
					WillReturnError(sql.ErrConnDone)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close database", zap.Error(err))
				}
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			storage := NewPostgresStorage(db)
			err := storage.DeleteData(context.Background(), tt.dataID)

			if (err != nil) != tt.wantError {
				t.Errorf("DeleteData() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
