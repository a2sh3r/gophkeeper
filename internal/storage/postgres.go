package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// PostgresStorage implements PostgreSQL storage
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates new PostgreSQL storage
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// CreateUser creates a new user in PostgreSQL
func (s *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, username, password, master_password, salt, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Username, user.Password, user.MasterPassword, user.Salt, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if err.Error() == `duplicate key value violates unique constraint "users_username_key"` {
			logger.Log.Warn("User already exists", zap.String("username", user.Username))
			return ErrUserExists
		}
		logger.Log.Error("Failed to create user in database", zap.Error(err), zap.String("username", user.Username))
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByUsername gets user by username
func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE username = $1`

	row := s.db.QueryRowContext(ctx, query, username)
	user := &models.User{}

	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.MasterPassword, &user.Salt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debug("User not found by username", zap.String("username", username))
			return nil, ErrUserNotFound
		}
		logger.Log.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID gets user by ID
func (s *PostgresStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `SELECT id, username, password, master_password, salt, created_at, updated_at FROM users WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, userID)
	user := &models.User{}

	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.MasterPassword, &user.Salt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debug("User not found by ID", zap.String("user_id", userID.String()))
			return nil, ErrUserNotFound
		}
		logger.Log.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateData creates new data
func (s *PostgresStorage) CreateData(ctx context.Context, data *models.Data) error {
	query := `INSERT INTO data (id, user_id, type, name, description, data, metadata, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.db.ExecContext(ctx, query, data.ID, data.UserID, data.Type, data.Name, data.Description,
		data.Data, data.Metadata, data.CreatedAt, data.UpdatedAt)
	if err != nil {
		logger.Log.Error("Failed to create data in database", zap.Error(err),
			zap.String("data_id", data.ID.String()), zap.String("user_id", data.UserID.String()))
		return fmt.Errorf("failed to create data: %w", err)
	}
	return nil
}

// GetDataByID gets data by ID
func (s *PostgresStorage) GetDataByID(ctx context.Context, dataID uuid.UUID) (*models.Data, error) {
	query := `SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at 
			  FROM data WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, dataID)
	data := &models.Data{}

	err := row.Scan(&data.ID, &data.UserID, &data.Type, &data.Name, &data.Description,
		&data.Data, &data.Metadata, &data.CreatedAt, &data.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debug("Data not found by ID", zap.String("data_id", dataID.String()))
			return nil, ErrDataNotFound
		}
		logger.Log.Error("Failed to get data by ID", zap.Error(err), zap.String("data_id", dataID.String()))
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	return data, nil
}

// GetDataByUserID gets all data for a user
func (s *PostgresStorage) GetDataByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Data, error) {
	query := `SELECT id, user_id, type, name, description, data, metadata, created_at, updated_at 
			  FROM data WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Log.Error("Failed to query user data", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to query data: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Log.Error("Failed to close database", zap.Error(err))
		}
	}()

	var dataList []*models.Data
	for rows.Next() {
		data := &models.Data{}
		err := rows.Scan(&data.ID, &data.UserID, &data.Type, &data.Name, &data.Description,
			&data.Data, &data.Metadata, &data.CreatedAt, &data.UpdatedAt)
		if err != nil {
			logger.Log.Error("Failed to scan data row", zap.Error(err), zap.String("user_id", userID.String()))
			return nil, fmt.Errorf("failed to scan data: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error("Rows iteration error", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return dataList, nil
}

// UpdateData updates data
func (s *PostgresStorage) UpdateData(ctx context.Context, data *models.Data) error {
	query := `UPDATE data SET type = $2, name = $3, description = $4, data = $5, metadata = $6, updated_at = $7 
			  WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, data.ID, data.Type, data.Name, data.Description,
		data.Data, data.Metadata, data.UpdatedAt)
	if err != nil {
		logger.Log.Error("Failed to update data in database", zap.Error(err),
			zap.String("data_id", data.ID.String()))
		return fmt.Errorf("failed to update data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected for update", zap.Error(err),
			zap.String("data_id", data.ID.String()))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Log.Debug("Data not found for update", zap.String("data_id", data.ID.String()))
		return ErrDataNotFound
	}

	return nil
}

// DeleteData deletes data
func (s *PostgresStorage) DeleteData(ctx context.Context, dataID uuid.UUID) error {
	query := `DELETE FROM data WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, dataID)
	if err != nil {
		logger.Log.Error("Failed to delete data from database", zap.Error(err),
			zap.String("data_id", dataID.String()))
		return fmt.Errorf("failed to delete data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("Failed to get rows affected for delete", zap.Error(err),
			zap.String("data_id", dataID.String()))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Log.Debug("Data not found for deletion", zap.String("data_id", dataID.String()))
		return ErrDataNotFound
	}

	return nil
}
