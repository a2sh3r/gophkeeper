package client

import (
	"context"

	"github.com/a2sh3r/gophkeeper/internal/crypto"
	"github.com/a2sh3r/gophkeeper/internal/models"
)

// ClientSession represents a client session with authentication and encryption
type ClientSession struct {
	cli            *Client
	cryptoManager  *crypto.CryptoManager
	masterPassword string
}

// NewClientSession creates a new client session
func NewClientSession(cli *Client) *ClientSession {
	return &ClientSession{
		cli: cli,
	}
}

// SetCryptoManager sets the crypto manager for the session
func (s *ClientSession) SetCryptoManager(cryptoManager *crypto.CryptoManager, masterPassword string) {
	s.cryptoManager = cryptoManager
	s.masterPassword = masterPassword
}

// IsAuthenticated checks if the session is authenticated with crypto manager
func (s *ClientSession) IsAuthenticated() bool {
	return s.cryptoManager != nil
}

// GetClient returns the underlying HTTP client
func (s *ClientSession) GetClient() *Client {
	return s.cli
}

// GetCryptoManager returns the crypto manager
func (s *ClientSession) GetCryptoManager() *crypto.CryptoManager {
	return s.cryptoManager
}

// Register registers a new user
func (s *ClientSession) Register(ctx context.Context, username, password, masterPassword string) (*models.AuthResponse, error) {
	return s.cli.Register(ctx, username, password, masterPassword)
}

// Login authenticates user
func (s *ClientSession) Login(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	return s.cli.Login(ctx, username, password)
}

// List gets all user data
func (s *ClientSession) List(ctx context.Context) ([]models.Data, error) {
	if !s.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	return s.cli.GetData(ctx)
}

// Get gets data by ID
func (s *ClientSession) Get(ctx context.Context, id string) (*models.Data, error) {
	if !s.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	return s.cli.GetDataByID(ctx, id)
}

// Create creates new data
func (s *ClientSession) Create(ctx context.Context, dataReq models.DataRequest) (*models.Data, error) {
	if !s.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	return s.cli.CreateData(ctx, dataReq)
}

// Update updates data
func (s *ClientSession) Update(ctx context.Context, id string, dataReq models.DataRequest) (*models.Data, error) {
	if !s.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}
	return s.cli.UpdateData(ctx, id, dataReq)
}

// Delete deletes data
func (s *ClientSession) Delete(ctx context.Context, id string) error {
	if !s.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	return s.cli.DeleteData(ctx, id)
}
