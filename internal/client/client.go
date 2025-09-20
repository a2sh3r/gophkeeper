package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/models"
	"go.uber.org/zap"
)

// Client represents client for server interaction
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates new client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

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

// GetData gets all user data
func (c *Client) GetData(ctx context.Context) ([]models.Data, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/data", nil)
	if err != nil {
		logger.Log.Error("Failed to create GET data request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("GET data request failed", zap.Error(err))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read GET data response", zap.Error(err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			logger.Log.Warn("GET data failed with server error", zap.Int("status_code", resp.StatusCode),
				zap.String("error", errResp.Error))
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		logger.Log.Warn("GET data failed with unknown error", zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var dataResp models.DataListResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		logger.Log.Error("Failed to unmarshal GET data response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return dataResp.Data, nil
}

// CreateData creates new data
func (c *Client) CreateData(ctx context.Context, dataReq models.DataRequest) (*models.Data, error) {
	jsonData, err := json.Marshal(dataReq)
	if err != nil {
		logger.Log.Error("Failed to marshal create data request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/data", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.Error("Failed to create POST data request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("POST data request failed", zap.Error(err))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read POST data response", zap.Error(err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			logger.Log.Warn("POST data failed with server error", zap.Int("status_code", resp.StatusCode),
				zap.String("error", errResp.Error))
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		logger.Log.Warn("POST data failed with unknown error", zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var dataResp models.DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		logger.Log.Error("Failed to unmarshal POST data response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &dataResp.Data, nil
}

// GetDataByID gets data by ID
func (c *Client) GetDataByID(ctx context.Context, id string) (*models.Data, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/data/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var dataResp models.DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &dataResp.Data, nil
}

// UpdateData updates data
func (c *Client) UpdateData(ctx context.Context, id string, dataReq models.DataRequest) (*models.Data, error) {
	jsonData, err := json.Marshal(dataReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/api/v1/data/"+id, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var dataResp models.DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &dataResp.Data, nil
}

// DeleteData deletes data
func (c *Client) DeleteData(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/api/v1/data/"+id, nil)
	if err != nil {
		logger.Log.Error("Failed to create DELETE data request", zap.Error(err), zap.String("data_id", id))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("DELETE data request failed", zap.Error(err), zap.String("data_id", id))
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error("Failed to close body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Log.Error("Failed to read DELETE data response", zap.Error(err), zap.String("data_id", id))
			return fmt.Errorf("failed to read response: %w", err)
		}

		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			logger.Log.Warn("DELETE data failed with server error", zap.Int("status_code", resp.StatusCode),
				zap.String("data_id", id), zap.String("error", errResp.Error))
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		logger.Log.Warn("DELETE data failed with unknown error", zap.Int("status_code", resp.StatusCode),
			zap.String("data_id", id), zap.String("response", string(body)))
		return fmt.Errorf("server error: %s", string(body))
	}

	return nil
}
