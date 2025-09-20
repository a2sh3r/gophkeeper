package db

import (
	"database/sql"
	"fmt"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// DB represents database connection and operations
type DB struct {
	conn *sql.DB
}

// New creates new database connection and runs migrations
func New(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Log.Error("Failed to open database connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		logger.Log.Error("Failed to run database migrations", zap.Error(err))
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Log.Info("Database connection established and migrations completed")
	return db, nil
}

// Close closes database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns database connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// migrate runs database migrations
func (db *DB) migrate() error {
	driver, err := postgres.WithInstance(db.conn, &postgres.Config{})
	if err != nil {
		logger.Log.Error("Failed to create migration driver", zap.Error(err))
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		logger.Log.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Log.Info("Database migrations completed successfully")
	return nil
}
