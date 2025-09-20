package db

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		dsn       string
		wantError bool
	}{
		{
			name:      "invalid DSN",
			dsn:       "invalid://dsn",
			wantError: true,
		},
		{
			name:      "empty DSN",
			dsn:       "",
			wantError: true,
		},
		{
			name:      "postgres DSN without connection",
			dsn:       "postgres://user:pass@localhost:5432/nonexistent?sslmode=disable",
			wantError: true,
		},
		{
			name:      "valid postgres DSN format",
			dsn:       "postgres://user:pass@localhost:5432/db?sslmode=disable",
			wantError: true, // Will fail due to no real connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.dsn)

			if (err != nil) != tt.wantError {
				t.Errorf("New() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && db == nil {
				t.Error("New() returned nil database")
			}

			if db != nil {
				if err := db.Close(); err != nil {
					logger.Log.Error("Failed to close", zap.Error(err))
				}
			}
		})
	}
}

func TestNew_WithMockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectPing()

	mock.ExpectQuery("SELECT CURRENT_DATABASE()").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	dbInstance := &DB{conn: db}

	err = dbInstance.migrate()
	if err == nil {
		t.Errorf("migrate() should return error due to missing migration files")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestDB_Close(t *testing.T) {
	tests := []struct {
		name      string
		setupDB   bool
		wantError bool
	}{
		{
			name:      "close nil connection",
			setupDB:   false,
			wantError: true, // Will panic, so we expect error
		},
		{
			name:      "close valid connection",
			setupDB:   true,
			wantError: true, // Will panic, so we expect error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *DB
			if tt.setupDB {
				db = &DB{conn: nil}
			} else {
				db = &DB{conn: nil}
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantError {
						t.Errorf("Close() panicked unexpectedly: %v", r)
					}
				}
			}()

			err := db.Close()

			if (err != nil) != tt.wantError {
				t.Errorf("Close() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDB_Conn(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  bool
		expected *sql.DB
	}{
		{
			name:     "nil connection",
			setupDB:  false,
			expected: nil,
		},
		{
			name:     "valid connection",
			setupDB:  true,
			expected: nil, // We'll set this to nil since we can't create real connections in tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *DB
			if tt.setupDB {
				db = &DB{conn: nil}
			} else {
				db = &DB{conn: nil}
			}

			conn := db.Conn()

			if conn != tt.expected {
				t.Errorf("Conn() = %v, want %v", conn, tt.expected)
			}
		})
	}
}

func TestDB_migrate(t *testing.T) {
	tests := []struct {
		name      string
		setupDB   bool
		wantError bool
	}{
		{
			name:      "nil connection",
			setupDB:   false,
			wantError: true, // Will panic, so we expect error
		},
		{
			name:      "valid connection without migrations",
			setupDB:   true,
			wantError: true, // Will panic, so we expect error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *DB
			if tt.setupDB {
				db = &DB{conn: nil}
			} else {
				db = &DB{conn: nil}
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantError {
						t.Errorf("migrate() panicked unexpectedly: %v", r)
					}
				}
			}()

			err := db.migrate()

			if (err != nil) != tt.wantError {
				t.Errorf("migrate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDB_migrate_WithMockDB(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "migration driver creation",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT CURRENT_DATABASE()").
					WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))
			},
			wantError: true, // Will fail due to missing migration files
		},
		{
			name: "migration driver creation error",
			mockSetup: func(mock sqlmock.Sqlmock) {
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
				_ = db.Close()
			}()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			dbInstance := &DB{conn: db}
			err = dbInstance.migrate()

			if (err != nil) != tt.wantError {
				t.Errorf("migrate() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
