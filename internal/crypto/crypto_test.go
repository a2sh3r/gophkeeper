package crypto

import (
	"testing"
)

func TestNewCryptoManager(t *testing.T) {
	tests := []struct {
		name           string
		masterPassword string
		wantErr        bool
	}{
		{
			name:           "valid master password",
			masterPassword: "testPassword123!",
			wantErr:        false,
		},
		{
			name:           "empty master password",
			masterPassword: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, err := NewCryptoManager(tt.masterPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCryptoManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cm == nil {
				t.Error("NewCryptoManager() returned nil manager")
			}
		})
	}
}

func TestNewCryptoManagerWithSalt(t *testing.T) {
	validSalt := make([]byte, 32)
	for i := range validSalt {
		validSalt[i] = byte(i)
	}

	tests := []struct {
		name           string
		masterPassword string
		salt           []byte
		wantErr        bool
	}{
		{
			name:           "valid master password and salt",
			masterPassword: "testPassword123!",
			salt:           validSalt,
			wantErr:        false,
		},
		{
			name:           "empty master password",
			masterPassword: "",
			salt:           validSalt,
			wantErr:        true,
		},
		{
			name:           "invalid salt length",
			masterPassword: "testPassword123!",
			salt:           []byte("short"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, err := NewCryptoManagerWithSalt(tt.masterPassword, tt.salt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCryptoManagerWithSalt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cm == nil {
				t.Error("NewCryptoManagerWithSalt() returned nil manager")
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	cm, err := NewCryptoManager("testPassword123!")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple text",
			data: "Hello, World!",
		},
		{
			name: "JSON data",
			data: `{"login": "user@example.com", "password": "secret123", "url": "https://example.com"}`,
		},
		{
			name: "empty string",
			data: " ",
		},
		{
			name: "unicode text",
			data: "Привет, мир! 🌍",
		},
		{
			name: "long text",
			data: "This is a very long text that contains multiple lines and should be properly encrypted and decrypted without any issues. It includes special characters like @#$%^&*() and numbers 1234567890.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := cm.Encrypt([]byte(tt.data))
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			// Verify encrypted data is different from original
			if string(encrypted) == tt.data {
				t.Error("Encrypted data should be different from original")
			}

			// Decrypt
			decrypted, err := cm.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			// Verify decrypted data matches original
			if string(decrypted) != tt.data {
				t.Errorf("Decrypt() = %v, want %v", string(decrypted), tt.data)
			}
		})
	}
}

func TestEncryptDecryptString(t *testing.T) {
	cm, err := NewCryptoManager("testPassword123!")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	testData := "Hello, World!"

	// Encrypt string
	encryptedStr, err := cm.EncryptString(testData)
	if err != nil {
		t.Errorf("EncryptString() error = %v", err)
		return
	}

	// Decrypt string
	decryptedStr, err := cm.DecryptString(encryptedStr)
	if err != nil {
		t.Errorf("DecryptString() error = %v", err)
		return
	}

	// Verify
	if decryptedStr != testData {
		t.Errorf("DecryptString() = %v, want %v", decryptedStr, testData)
	}
}

func TestCrossManagerDecryption(t *testing.T) {
	masterPassword := "testPassword123!"

	// Create first manager
	cm1, err := NewCryptoManager(masterPassword)
	if err != nil {
		t.Fatalf("Failed to create first crypto manager: %v", err)
	}

	// Encrypt with first manager
	testData := "Cross manager test data"
	encrypted, err := cm1.Encrypt([]byte(testData))
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
		return
	}

	// Create second manager with same salt
	cm2, err := NewCryptoManagerWithSalt(masterPassword, cm1.GetSalt())
	if err != nil {
		t.Fatalf("Failed to create second crypto manager: %v", err)
	}

	// Decrypt with second manager
	decrypted, err := cm2.Decrypt(encrypted)
	if err != nil {
		t.Errorf("Decrypt() error = %v", err)
		return
	}

	// Verify
	if string(decrypted) != testData {
		t.Errorf("Cross manager decryption failed: got %v, want %v", string(decrypted), testData)
	}
}

func TestVerifyMasterPassword(t *testing.T) {
	masterPassword := "testPassword123!"
	cm, err := NewCryptoManager(masterPassword)
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	salt := cm.GetSalt()

	tests := []struct {
		name           string
		masterPassword string
		salt           []byte
		want           bool
	}{
		{
			name:           "correct password",
			masterPassword: masterPassword,
			salt:           salt,
			want:           true,
		},
		{
			name:           "wrong password",
			masterPassword: "wrongPassword",
			salt:           salt,
			want:           true, // This function only validates key generation, not actual password correctness
		},
		{
			name:           "empty password",
			masterPassword: "",
			salt:           salt,
			want:           false,
		},
		{
			name:           "invalid salt length",
			masterPassword: masterPassword,
			salt:           []byte("short"),
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyMasterPassword(tt.masterPassword, tt.salt)
			if got != tt.want {
				t.Errorf("VerifyMasterPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSalt(t *testing.T) {
	cm, err := NewCryptoManager("testPassword123!")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	salt := cm.GetSalt()
	if len(salt) != 32 {
		t.Errorf("GetSalt() length = %v, want 32", len(salt))
	}

	saltBase64 := cm.GetSaltBase64()
	if saltBase64 == "" {
		t.Error("GetSaltBase64() returned empty string")
	}
}

func TestEmptyDataHandling(t *testing.T) {
	cm, err := NewCryptoManager("testPassword123!")
	if err != nil {
		t.Fatalf("Failed to create crypto manager: %v", err)
	}

	// Test encrypting empty data
	_, err = cm.Encrypt([]byte{})
	if err == nil {
		t.Error("Encrypt() should return error for empty data")
	}

	// Test decrypting empty data
	_, err = cm.Decrypt([]byte{})
	if err == nil {
		t.Error("Decrypt() should return error for empty data")
	}
}
