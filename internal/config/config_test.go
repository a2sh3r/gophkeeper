package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"go.uber.org/zap"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Database: DatabaseConfig{
					Type:     "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "gophkeeper",
					User:     "postgres",
					Password: "password",
					SSLMode:  "disable",
				},
				JWT: JWTConfig{
					Secret:      "your-secret-key",
					TokenExpiry: 24 * time.Hour,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			config := Load()

			if config.Server.Host != tt.expected.Server.Host {
				t.Errorf("Expected Server.Host %s, got %s", tt.expected.Server.Host, config.Server.Host)
			}
			if config.Server.Port != tt.expected.Server.Port {
				t.Errorf("Expected Server.Port %d, got %d", tt.expected.Server.Port, config.Server.Port)
			}
			if config.Database.Type != tt.expected.Database.Type {
				t.Errorf("Expected Database.Type %s, got %s", tt.expected.Database.Type, config.Database.Type)
			}
			if config.Database.Host != tt.expected.Database.Host {
				t.Errorf("Expected Database.Host %s, got %s", tt.expected.Database.Host, config.Database.Host)
			}
			if config.Database.Port != tt.expected.Database.Port {
				t.Errorf("Expected Database.Port %d, got %d", tt.expected.Database.Port, config.Database.Port)
			}
			if config.Database.Name != tt.expected.Database.Name {
				t.Errorf("Expected Database.Name %s, got %s", tt.expected.Database.Name, config.Database.Name)
			}
			if config.Database.User != tt.expected.Database.User {
				t.Errorf("Expected Database.User %s, got %s", tt.expected.Database.User, config.Database.User)
			}
			if config.Database.Password != tt.expected.Database.Password {
				t.Errorf("Expected Database.Password %s, got %s", tt.expected.Database.Password, config.Database.Password)
			}
			if config.Database.SSLMode != tt.expected.Database.SSLMode {
				t.Errorf("Expected Database.SSLMode %s, got %s", tt.expected.Database.SSLMode, config.Database.SSLMode)
			}
			if config.JWT.Secret != tt.expected.JWT.Secret {
				t.Errorf("Expected JWT.Secret %s, got %s", tt.expected.JWT.Secret, config.JWT.Secret)
			}
			if config.JWT.TokenExpiry != tt.expected.JWT.TokenExpiry {
				t.Errorf("Expected JWT.TokenExpiry %v, got %v", tt.expected.JWT.TokenExpiry, config.JWT.TokenExpiry)
			}
		})
	}
}

func TestConfig_GetDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "postgres DSN",
			config: Config{
				Database: DatabaseConfig{
					Type:     "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "gophkeeper",
					User:     "postgres",
					Password: "password",
					SSLMode:  "disable",
				},
			},
			expected: "postgres://postgres:password@localhost:5432/gophkeeper?sslmode=disable",
		},
		{
			name: "postgres DSN with SSL",
			config: Config{
				Database: DatabaseConfig{
					Type:     "postgres",
					Host:     "db.example.com",
					Port:     5433,
					Name:     "testdb",
					User:     "testuser",
					Password: "testpass",
					SSLMode:  "require",
				},
			},
			expected: "postgres://testuser:testpass@db.example.com:5433/testdb?sslmode=require",
		},
		{
			name: "memory DSN",
			config: Config{
				Database: DatabaseConfig{
					Type: "memory",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.GetDSN()
			if dsn != tt.expected {
				t.Errorf("Expected DSN %s, got %s", tt.expected, dsn)
			}
		})
	}
}

func TestConfig_GetServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "default server address",
			config: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			expected: "localhost:8080",
		},
		{
			name: "custom server address",
			config: Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 9090,
				},
			},
			expected: "0.0.0.0:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := tt.config.GetServerAddr()
			if addr != tt.expected {
				t.Errorf("Expected server address %s, got %s", tt.expected, addr)
			}
		})
	}
}

func TestNetAddress_String(t *testing.T) {
	tests := []struct {
		name     string
		address  NetAddress
		expected string
	}{
		{
			name: "localhost address",
			address: NetAddress{
				Host: "localhost",
				Port: 8080,
			},
			expected: "localhost:8080",
		},
		{
			name: "IP address",
			address: NetAddress{
				Host: "192.168.1.1",
				Port: 9090,
			},
			expected: "192.168.1.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.address.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNetAddress_Set(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  NetAddress
		wantError bool
	}{
		{
			name:  "valid address",
			input: "localhost:8080",
			expected: NetAddress{
				Host: "localhost",
				Port: 8080,
			},
			wantError: false,
		},
		{
			name:  "IP address",
			input: "192.168.1.1:9090",
			expected: NetAddress{
				Host: "192.168.1.1",
				Port: 9090,
			},
			wantError: false,
		},
		{
			name:      "invalid format",
			input:     "localhost",
			wantError: true,
		},
		{
			name:      "invalid port",
			input:     "localhost:invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := &NetAddress{}
			err := addr.Set(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("Set() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if addr.Host != tt.expected.Host {
					t.Errorf("Expected Host %s, got %s", tt.expected.Host, addr.Host)
				}
				if addr.Port != tt.expected.Port {
					t.Errorf("Expected Port %d, got %d", tt.expected.Port, addr.Port)
				}
			}
		})
	}
}

func TestLoadConfigFile(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		configData interface{}
		wantError  bool
		setupFile  bool
	}{
		{
			name:       "empty config path",
			configPath: "",
			wantError:  false,
			setupFile:  false,
		},
		{
			name:       "non-existent file",
			configPath: "/tmp/nonexistent.json",
			wantError:  true,
			setupFile:  false,
		},
		{
			name:       "valid config file",
			configPath: "/tmp/test_config.json",
			configData: Config{
				Server: ServerConfig{
					Host: "testhost",
					Port: 9090,
				},
				Database: DatabaseConfig{
					Type: "memory",
				},
			},
			wantError: false,
			setupFile: true,
		},
		{
			name:       "invalid JSON file",
			configPath: "/tmp/invalid_config.json",
			configData: "invalid json",
			wantError:  true,
			setupFile:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				var fileData []byte
				var err error

				if tt.configData != nil {
					if config, ok := tt.configData.(Config); ok {
						fileData, err = json.Marshal(config)
						if err != nil {
							t.Fatalf("Failed to marshal config: %v", err)
						}
					} else {
						fileData = []byte(tt.configData.(string))
					}
				}

				err = os.WriteFile(tt.configPath, fileData, 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				defer func() {
					_ = os.Remove(tt.configPath)
				}()
			}

			cfg := &Config{}
			err := loadConfigFile(tt.configPath, cfg)

			if (err != nil) != tt.wantError {
				t.Errorf("loadConfigFile() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "no env var set",
			envVar:   "",
			expected: "",
		},
		{
			name:     "env var set",
			envVar:   "/path/to/config.json",
			expected: "/path/to/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				_ = os.Setenv("CONFIG", tt.envVar)
			} else {
				_ = os.Unsetenv("CONFIG")
			}

			defer func() {
				_ = os.Unsetenv("CONFIG")
			}()
			result := getConfigPath()
			if result != tt.expected {
				t.Errorf("getConfigPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewServerConfig(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name:      "default config",
			envVars:   map[string]string{},
			wantError: false,
		},
		{
			name: "custom env vars",
			envVars: map[string]string{
				"SERVER_HOST": "0.0.0.0",
				"SERVER_PORT": "9090",
				"DB_TYPE":     "memory",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					logger.Log.Error("Failed to set env", zap.Error(err))
				}
			}

			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			config, err := NewServerConfig()

			if (err != nil) != tt.wantError {
				t.Errorf("NewServerConfig() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && config == nil {
				t.Error("NewServerConfig() returned nil config")
			}
		})
	}
}

func TestConfig_ParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "parse server address flag",
			args: []string{"-a", "localhost:9090"},
			expected: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 9090,
				},
			},
		},
		{
			name: "parse database flags",
			args: []string{"-db-type", "memory", "-db-host", "testhost", "-db-port", "5433"},
			expected: Config{
				Database: DatabaseConfig{
					Type: "memory",
					Host: "testhost",
					Port: 5433,
				},
			},
		},
		{
			name: "parse JWT flags",
			args: []string{"-jwt-secret", "test-secret", "-jwt-expiry", "1h"},
			expected: Config{
				JWT: JWTConfig{
					Secret:      "test-secret",
					TokenExpiry: time.Hour,
				},
			},
		},
		{
			name: "parse log level flag",
			args: []string{"-log-level", "debug"},
			expected: Config{
				Server: ServerConfig{
					LogLevel: "debug",
				},
			},
		},
		{
			name: "parse multiple flags",
			args: []string{"-a", "localhost:9090", "-db-type", "memory", "-log-level", "warn"},
			expected: Config{
				Server: ServerConfig{
					Host:     "localhost",
					Port:     9090,
					LogLevel: "warn",
				},
				Database: DatabaseConfig{
					Type: "memory",
				},
			},
		},
		{
			name: "parse invalid port",
			args: []string{"-a", "localhost:invalid"},
			expected: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080, // Default port when parsing fails
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Host:     "localhost",
					Port:     8080,
					LogLevel: "info",
				},
				Database: DatabaseConfig{
					Type:     "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "gophkeeper",
					User:     "postgres",
					Password: "password",
					SSLMode:  "disable",
				},
				JWT: JWTConfig{
					Secret:      "your-secret-key",
					TokenExpiry: 24 * time.Hour,
				},
			}

			originalArgs := os.Args
			defer func() {
				os.Args = originalArgs
			}()

			os.Args = append([]string{"config"}, tt.args...)

			config.ParseFlags()

			if tt.expected.Server.Host != "" && config.Server.Host != tt.expected.Server.Host {
				t.Errorf("ParseFlags() Server.Host = %v, want %v", config.Server.Host, tt.expected.Server.Host)
			}
			if tt.expected.Server.Port != 0 && config.Server.Port != tt.expected.Server.Port {
				t.Errorf("ParseFlags() Server.Port = %v, want %v", config.Server.Port, tt.expected.Server.Port)
			}
			if tt.expected.Server.LogLevel != "" && config.Server.LogLevel != tt.expected.Server.LogLevel {
				t.Errorf("ParseFlags() Server.LogLevel = %v, want %v", config.Server.LogLevel, tt.expected.Server.LogLevel)
			}

			if tt.expected.Database.Type != "" && config.Database.Type != tt.expected.Database.Type {
				t.Errorf("ParseFlags() Database.Type = %v, want %v", config.Database.Type, tt.expected.Database.Type)
			}
			if tt.expected.Database.Host != "" && config.Database.Host != tt.expected.Database.Host {
				t.Errorf("ParseFlags() Database.Host = %v, want %v", config.Database.Host, tt.expected.Database.Host)
			}
			if tt.expected.Database.Port != 0 && config.Database.Port != tt.expected.Database.Port {
				t.Errorf("ParseFlags() Database.Port = %v, want %v", config.Database.Port, tt.expected.Database.Port)
			}

			if tt.expected.JWT.Secret != "" && config.JWT.Secret != tt.expected.JWT.Secret {
				t.Errorf("ParseFlags() JWT.Secret = %v, want %v", config.JWT.Secret, tt.expected.JWT.Secret)
			}
			if tt.expected.JWT.TokenExpiry != 0 && config.JWT.TokenExpiry != tt.expected.JWT.TokenExpiry {
				t.Errorf("ParseFlags() JWT.TokenExpiry = %v, want %v", config.JWT.TokenExpiry, tt.expected.JWT.TokenExpiry)
			}
		})
	}
}
