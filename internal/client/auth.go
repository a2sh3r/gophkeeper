package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"go.uber.org/zap"
)

// Register registers new user
func (c *Client) Register(ctx context.Context, username, password, masterPassword string) (*models.AuthResponse, error) {
	req := models.UserRequest{
		Username:       username,
		Password:       password,
		MasterPassword: masterPassword,
	}

	return c.authRequest(ctx, "/api/v1/register", req)
}

// Login authenticates user
func (c *Client) Login(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	req := models.LoginRequest{
		Username: username,
		Password: password,
	}

	return c.authRequest(ctx, "/api/v1/login", req)
}

// authRequest performs authentication request
func (c *Client) authRequest(ctx context.Context, endpoint string, req interface{}) (*models.AuthResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to marshal auth request", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.Error("Failed to create auth request", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Log.Error("Auth request failed", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read auth response", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			logger.Log.Warn("Auth request failed with server error", zap.String("endpoint", endpoint),
				zap.Int("status_code", resp.StatusCode), zap.String("error", errResp.Error))
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		logger.Log.Warn("Auth request failed with unknown error", zap.String("endpoint", endpoint),
			zap.Int("status_code", resp.StatusCode), zap.String("response", string(body)))
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var authResp models.AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		logger.Log.Error("Failed to unmarshal auth response", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &authResp, nil
}

