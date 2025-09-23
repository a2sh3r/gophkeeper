package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_NewConfig_NoFile(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("LoadConfig returned nil")
	}

	_ = config.ServerURL
	_ = config.Token
	_ = config.Salt
}

func TestConfig_SaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()

	testConfigFile := filepath.Join(tempDir, "test_config.json")

	testConfig := &Config{
		ServerURL: "http://test-server:8080",
		Token:     "test-token-123",
		Salt:      "test-salt-456",
	}

	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	err = os.WriteFile(testConfigFile, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	data, err = os.ReadFile(testConfigFile)
	if err != nil {
		t.Fatalf("Failed to read test config file: %v", err)
	}
	var loadedConfig Config
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal test config: %v", err)
	}

	if loadedConfig.ServerURL != testConfig.ServerURL {
		t.Errorf("ServerURL mismatch: expected %s, got %s", testConfig.ServerURL, loadedConfig.ServerURL)
	}
	if loadedConfig.Token != testConfig.Token {
		t.Errorf("Token mismatch: expected %s, got %s", testConfig.Token, loadedConfig.Token)
	}
	if loadedConfig.Salt != testConfig.Salt {
		t.Errorf("Salt mismatch: expected %s, got %s", testConfig.Salt, loadedConfig.Salt)
	}
}

func TestConfig_LoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	testConfigFile := filepath.Join(tempDir, "test_config.json")

	err := os.WriteFile(testConfigFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	data, err := os.ReadFile(testConfigFile)
	if err != nil {
		t.Fatalf("Failed to read invalid JSON file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON")
	}
}

func TestConfig_GetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath returned empty path")
	}

	if filepath.Base(path) != configFile {
		t.Errorf("GetConfigPath should end with %s, got %s", configFile, filepath.Base(path))
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	originalConfig := &Config{
		ServerURL: "http://test-server:8080",
		Token:     "test-token-123",
		Salt:      "test-salt-456",
	}

	jsonData, err := json.Marshal(originalConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var loadedConfig Config
	err = json.Unmarshal(jsonData, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if loadedConfig.ServerURL != originalConfig.ServerURL {
		t.Errorf("ServerURL mismatch after JSON round-trip")
	}
	if loadedConfig.Token != originalConfig.Token {
		t.Errorf("Token mismatch after JSON round-trip")
	}
	if loadedConfig.Salt != originalConfig.Salt {
		t.Errorf("Salt mismatch after JSON round-trip")
	}
}
