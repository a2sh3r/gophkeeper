package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Nonce []byte `json:"nonce"`
	Salt  []byte `json:"salt"`
	Data  []byte `json:"data"`
}

// CryptoManager handles encryption and decryption operations
type CryptoManager struct {
	masterPassword string
	key            []byte
	salt           []byte
}

// NewCryptoManager creates a new crypto manager with master password
func NewCryptoManager(masterPassword string) (*CryptoManager, error) {
	if masterPassword == "" {
		return nil, fmt.Errorf("master password cannot be empty")
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)

	return &CryptoManager{
		masterPassword: masterPassword,
		key:            key,
		salt:           salt,
	}, nil
}

// NewCryptoManagerWithSalt creates a new crypto manager with existing salt
func NewCryptoManagerWithSalt(masterPassword string, salt []byte) (*CryptoManager, error) {
	if masterPassword == "" {
		return nil, fmt.Errorf("master password cannot be empty")
	}
	if len(salt) != 32 {
		return nil, fmt.Errorf("invalid salt length: expected 32 bytes, got %d", len(salt))
	}

	key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)

	return &CryptoManager{
		masterPassword: masterPassword,
		key:            key,
		salt:           salt,
	}, nil
}

// Encrypt encrypts data using AES-256-GCM
func (cm *CryptoManager) Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	block, err := aes.NewCipher(cm.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	encData := EncryptedData{
		Nonce: nonce,
		Salt:  cm.salt,
		Data:  encryptedData[len(nonce):],
	}

	jsonData, err := json.Marshal(encData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encrypted data: %w", err)
	}

	return jsonData, nil
}

// Decrypt decrypts data using AES-256-GCM
func (cm *CryptoManager) Decrypt(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("encrypted data cannot be empty")
	}

	var encData EncryptedData
	if err := json.Unmarshal(encryptedData, &encData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted data: %w", err)
	}

	if len(encData.Salt) != 32 {
		return nil, fmt.Errorf("invalid salt length in encrypted data")
	}

	key := pbkdf2.Key([]byte(cm.masterPassword), encData.Salt, 100000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	decryptedData, err := gcm.Open(nil, encData.Nonce, encData.Data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return decryptedData, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (cm *CryptoManager) EncryptString(data string) (string, error) {
	encrypted, err := cm.Encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 encoded string
func (cm *CryptoManager) DecryptString(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	decrypted, err := cm.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GetSalt returns the salt used for key derivation
func (cm *CryptoManager) GetSalt() []byte {
	return cm.salt
}

// GetSaltBase64 returns the salt as base64 encoded string
func (cm *CryptoManager) GetSaltBase64() string {
	return base64.StdEncoding.EncodeToString(cm.salt)
}

// VerifyMasterPassword verifies if the master password is correct for given salt
func VerifyMasterPassword(masterPassword string, salt []byte) bool {
	if masterPassword == "" || len(salt) != 32 {
		return false
	}

	key := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return false
	}

	_, err = cipher.NewGCM(block)
	return err == nil
}
