package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"go.uber.org/zap"
)

const (
	configFile = ".gophkeeper_config"
)

// Config represents client configuration
type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
	Salt      string `json:"salt"`
}

// LoadConfig loads configuration from file
func LoadConfig() *Config {
	config := &Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to get home directory", zap.Error(err))
		return config
	}

	configPath := fmt.Sprintf("%s/%s", homeDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Log.Error("Failed to read config file", zap.Error(err))
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		logger.Log.Error("Failed to unmarshal config", zap.Error(err))
		return config
	}
	return config
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to get home directory", zap.Error(err))
		return err
	}

	configPath := fmt.Sprintf("%s/%s", homeDir, configFile)
	data, err := json.Marshal(config)
	if err != nil {
		logger.Log.Error("Failed to marshal config", zap.Error(err))
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return configFile
	}
	return fmt.Sprintf("%s/%s", homeDir, configFile)
}
