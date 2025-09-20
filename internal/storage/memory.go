package storage

import (
	"errors"
	"sync"

	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrDataNotFound = errors.New("data not found")
)

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	users map[string]*models.User
	data  map[uuid.UUID]*models.Data
	mutex sync.RWMutex
}

// NewMemoryStorage creates new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users: make(map[string]*models.User),
		data:  make(map[uuid.UUID]*models.Data),
	}
}

// CreateUser creates new user
func (s *MemoryStorage) CreateUser(user *models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.users[user.Username]; exists {
		return ErrUserExists
	}

	s.users[user.Username] = user
	return nil
}

// GetUserByUsername gets user by username
func (s *MemoryStorage) GetUserByUsername(username string) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByID gets user by ID
func (s *MemoryStorage) GetUserByID(userID uuid.UUID) (*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// CreateData creates new data
func (s *MemoryStorage) CreateData(data *models.Data) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[data.ID] = data
	return nil
}

// GetDataByID gets data by ID
func (s *MemoryStorage) GetDataByID(dataID uuid.UUID) (*models.Data, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, exists := s.data[dataID]
	if !exists {
		return nil, ErrDataNotFound
	}

	return data, nil
}

// GetDataByUserID gets all user data
func (s *MemoryStorage) GetDataByUserID(userID uuid.UUID) ([]*models.Data, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var userData []*models.Data
	for _, data := range s.data {
		if data.UserID == userID {
			userData = append(userData, data)
		}
	}

	return userData, nil
}

// UpdateData updates data
func (s *MemoryStorage) UpdateData(data *models.Data) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.data[data.ID]; !exists {
		return ErrDataNotFound
	}

	s.data[data.ID] = data
	return nil
}

// DeleteData deletes data
func (s *MemoryStorage) DeleteData(dataID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.data[dataID]; !exists {
		return ErrDataNotFound
	}

	delete(s.data, dataID)
	return nil
}
