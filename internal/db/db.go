package db

import (
	"database/sql"
	"fmt"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// DB represents database connection and operations
type DB struct {
	conn *sql.DB
}

// New creates new database connection
func New(dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Error("Failed to open database connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("Database connection established")
	return &DB{conn: conn}, nil
}

// Close closes database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns database connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}
