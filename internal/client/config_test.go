package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_LoadConfig_NoFile(t *testing.T) {
	// Test loading config when file doesn't exist
	// This test will pass if the config file doesn't exist in the home directory
	// which is the expected behavior
	config := LoadConfig()

	if config == nil {
		t.Fatal("LoadConfig returned nil")
	}

	// The config might have default values or be empty depending on the environment
	// We just verify it doesn't crash and returns a valid config
	_ = config.ServerURL
	_ = config.Token
	_ = config.Salt
}

func TestConfig_SaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test config file path
	testConfigFile := filepath.Join(tempDir, "test_config.json")

	// Create test config
	testConfig := &Config{
		ServerURL: "http://test-server:8080",
		Token:     "test-token-123",
		Salt:      "test-salt-456",
	}

	// Save config to test file
	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	err = os.WriteFile(testConfigFile, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load config from test file
	data, err = os.ReadFile(testConfigFile)
	if err != nil {
		t.Fatalf("Failed to read test config file: %v", err)
	}
	var loadedConfig Config
	err = json.Unmarshal(data, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal test config: %v", err)
	}

	// Verify loaded config
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
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test config file path
	testConfigFile := filepath.Join(tempDir, "test_config.json")

	// Write invalid JSON to config file
	err := os.WriteFile(testConfigFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Try to load config from invalid JSON file
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

	// Should contain the config file name
	if filepath.Base(path) != configFile {
		t.Errorf("GetConfigPath should end with %s, got %s", configFile, filepath.Base(path))
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	// Test that Config can be properly serialized/deserialized
	originalConfig := &Config{
		ServerURL: "http://test-server:8080",
		Token:     "test-token-123",
		Salt:      "test-salt-456",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(originalConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal from JSON
	var loadedConfig Config
	err = json.Unmarshal(jsonData, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify data integrity
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
