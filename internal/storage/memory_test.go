package storage

import (
	"context"
	"testing"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/models"
	"github.com/google/uuid"
)

func TestMemoryStorage_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &models.User{
				ID:        uuid.New(),
				Username:  "testuser",
				Password:  "hashedpassword",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "user with empty username",
			user: &models.User{
				ID:        uuid.New(),
				Username:  "",
				Password:  "hashedpassword",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemoryStorage()
			err := storage.CreateUser(context.Background(), tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				retrievedUser, err := storage.GetUserByUsername(context.Background(), tt.user.Username)
				if err != nil {
					t.Errorf("Failed to get user: %v", err)
					return
				}
				if retrievedUser.Username != tt.user.Username {
					t.Errorf("Expected username %s, got %s", tt.user.Username, retrievedUser.Username)
				}
			}
		})
	}
}

func TestMemoryStorage_CreateUser_Duplicate(t *testing.T) {
	storage := NewMemoryStorage()
	user := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	duplicateUser := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Password:  "anotherpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.CreateUser(context.Background(), duplicateUser)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestMemoryStorage_GetUserByUsername(t *testing.T) {
	storage := NewMemoryStorage()
	user := &models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "existing user",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "non-existing user",
			username: "nonexistent",
			wantErr:  true,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.GetUserByUsername(context.Background(), tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_GetUserByID(t *testing.T) {
	storage := NewMemoryStorage()
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Username:  "testuser",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      userID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.GetUserByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_CreateData(t *testing.T) {
	storage := NewMemoryStorage()
	userID := uuid.New()

	tests := []struct {
		name    string
		data    *models.Data
		wantErr bool
	}{
		{
			name: "valid data",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        models.DataTypeLoginPassword,
				Name:        "Test Login",
				Description: "Test description",
				Data:        []byte("test data"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "data with empty name",
			data: &models.Data{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        models.DataTypeText,
				Name:        "",
				Description: "Test description",
				Data:        []byte("test data"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.CreateData(context.Background(), tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				retrievedData, err := storage.GetDataByID(context.Background(), tt.data.ID)
				if err != nil {
					t.Errorf("Failed to get data: %v", err)
					return
				}
				if retrievedData.Name != tt.data.Name {
					t.Errorf("Expected name %s, got %s", tt.data.Name, retrievedData.Name)
				}
			}
		})
	}
}

func TestMemoryStorage_GetDataByUserID(t *testing.T) {
	storage := NewMemoryStorage()
	userID1 := uuid.New()
	userID2 := uuid.New()

	for i := 0; i < 3; i++ {
		data := &models.Data{
			ID:          uuid.New(),
			UserID:      userID1,
			Type:        models.DataTypeText,
			Name:        "Test Data " + string(rune(i)),
			Description: "Test description",
			Data:        []byte("test data"),
			Metadata:    "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := storage.CreateData(context.Background(), data)
		if err != nil {
			t.Fatalf("Failed to create data: %v", err)
		}
	}

	data2 := &models.Data{
		ID:          uuid.New(),
		UserID:      userID2,
		Type:        models.DataTypeText,
		Name:        "Other User Data",
		Description: "Test description",
		Data:        []byte("test data"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.CreateData(context.Background(), data2)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	tests := []struct {
		name      string
		userID    uuid.UUID
		wantCount int
	}{
		{
			name:      "user with data",
			userID:    userID1,
			wantCount: 3,
		},
		{
			name:      "user with one data",
			userID:    userID2,
			wantCount: 1,
		},
		{
			name:      "user with no data",
			userID:    uuid.New(),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := storage.GetDataByUserID(context.Background(), tt.userID)
			if err != nil {
				t.Errorf("GetDataByUserID() error = %v", err)
				return
			}

			if len(data) != tt.wantCount {
				t.Errorf("Expected %d data items, got %d", tt.wantCount, len(data))
			}
		})
	}
}

func TestMemoryStorage_UpdateData(t *testing.T) {
	storage := NewMemoryStorage()
	userID := uuid.New()
	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        models.DataTypeText,
		Name:        "Original Name",
		Description: "Original description",
		Data:        []byte("original data"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.CreateData(context.Background(), data)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	tests := []struct {
		name       string
		updateData *models.Data
		wantErr    bool
	}{
		{
			name: "valid update",
			updateData: &models.Data{
				ID:          data.ID,
				UserID:      userID,
				Type:        models.DataTypeText,
				Name:        "Updated Name",
				Description: "Updated description",
				Data:        []byte("updated data"),
				Metadata:    "{}",
				CreatedAt:   data.CreatedAt,
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "update non-existing data",
			updateData: &models.Data{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        models.DataTypeText,
				Name:        "Non-existing",
				Description: "Test description",
				Data:        []byte("test data"),
				Metadata:    "{}",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.UpdateData(context.Background(), tt.updateData)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				retrievedData, err := storage.GetDataByID(context.Background(), tt.updateData.ID)
				if err != nil {
					t.Errorf("Failed to get updated data: %v", err)
					return
				}
				if retrievedData.Name != tt.updateData.Name {
					t.Errorf("Expected updated name %s, got %s", tt.updateData.Name, retrievedData.Name)
				}
			}
		})
	}
}

func TestMemoryStorage_DeleteData(t *testing.T) {
	storage := NewMemoryStorage()
	userID := uuid.New()
	data := &models.Data{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        models.DataTypeText,
		Name:        "Test Data",
		Description: "Test description",
		Data:        []byte("test data"),
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.CreateData(context.Background(), data)
	if err != nil {
		t.Fatalf("Failed to create data: %v", err)
	}

	tests := []struct {
		name    string
		dataID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "existing data",
			dataID:  data.ID,
			wantErr: false,
		},
		{
			name:    "non-existing data",
			dataID:  uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DeleteData(context.Background(), tt.dataID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				_, err = storage.GetDataByID(context.Background(), tt.dataID)
				if err != ErrDataNotFound {
					t.Errorf("Expected ErrDataNotFound, got %v", err)
				}
			}
		})
	}
}
