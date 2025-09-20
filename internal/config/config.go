// Package config provides configuration structures and functions for the server and client.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// ServerConfig holds configuration for the server.
type ServerConfig struct {
	Host     string `env:"SERVER_HOST" envDefault:"localhost" json:"host,omitempty"`
	Port     int    `env:"SERVER_PORT" envDefault:"8080" json:"port,omitempty"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info" json:"log_level,omitempty"`
}

// DatabaseConfig holds configuration for the database.
type DatabaseConfig struct {
	Type     string `env:"DB_TYPE" envDefault:"postgres" json:"type,omitempty"`
	Host     string `env:"DB_HOST" envDefault:"localhost" json:"host,omitempty"`
	Port     int    `env:"DB_PORT" envDefault:"5432" json:"port,omitempty"`
	Name     string `env:"DB_NAME" envDefault:"gophkeeper" json:"name,omitempty"`
	User     string `env:"DB_USER" envDefault:"postgres" json:"user,omitempty"`
	Password string `env:"DB_PASSWORD" envDefault:"password" json:"password,omitempty"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable" json:"ssl_mode,omitempty"`
}

// JWTConfig holds configuration for JWT authentication.
type JWTConfig struct {
	Secret      string        `env:"JWT_SECRET" envDefault:"your-secret-key" json:"secret,omitempty"`
	TokenExpiry time.Duration `env:"JWT_TOKEN_EXPIRY" envDefault:"24h" json:"token_expiry,omitempty"`
}

// Config represents application configuration.
type Config struct {
	Server   ServerConfig   `json:"server,omitempty"`
	Database DatabaseConfig `json:"database,omitempty"`
	JWT      JWTConfig      `json:"jwt,omitempty"`
}

// NetAddress represents a network address with host and port.
type NetAddress struct {
	Host string
	Port int
}

// String returns the string representation of the network address.
func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// Set parses and sets the network address from a string.
func (n *NetAddress) Set(flagValue string) error {
	parts := strings.Split(flagValue, ":")
	if len(parts) != 2 {
		return fmt.Errorf("address must be in format host:port")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	n.Host = parts[0]
	n.Port = port
	return nil
}

func loadConfigFile(configPath string, config interface{}) error {
	if configPath == "" {
		return nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close config file: %v\n", closeErr)
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

func getConfigPath() string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	}

	var configPath string
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&configPath, "c", "", "path to config file")
	fs.StringVar(&configPath, "config", "", "path to config file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return ""
	}

	return configPath
}

// NewServerConfig creates a new ServerConfig with priority: flags > env > config file.
func NewServerConfig() (*Config, error) {
	configPath := getConfigPath()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := loadConfigFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Parse command line flags to override values
	cfg.ParseFlags()

	return cfg, nil
}

// ParseFlags parses command-line flags into the Config.
// This function should be called after NewServerConfig() to override values with flags.
func (cfg *Config) ParseFlags() {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	addr := new(NetAddress)

	var (
		dbType     string
		dbHost     string
		dbPort     int
		dbName     string
		dbUser     string
		dbPassword string
		dbSSLMode  string
		jwtSecret  string
		jwtExpiry  time.Duration
		logLevel   string
	)

	fs.Var(addr, "a", "Net address host:port")
	fs.StringVar(&dbType, "db-type", "", "Database type (postgres, memory)")
	fs.StringVar(&dbHost, "db-host", "", "Database host")
	fs.IntVar(&dbPort, "db-port", 0, "Database port")
	fs.StringVar(&dbName, "db-name", "", "Database name")
	fs.StringVar(&dbUser, "db-user", "", "Database user")
	fs.StringVar(&dbPassword, "db-password", "", "Database password")
	fs.StringVar(&dbSSLMode, "db-sslmode", "", "Database SSL mode")
	fs.StringVar(&jwtSecret, "jwt-secret", "", "JWT secret key")
	fs.DurationVar(&jwtExpiry, "jwt-expiry", 0, "JWT token expiry")
	fs.StringVar(&logLevel, "log-level", "", "Log level (debug, info, warn, error)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return
	}

	if addr.Port != 0 {
		cfg.Server.Host = addr.Host
		cfg.Server.Port = addr.Port
	}

	if dbType != "" {
		cfg.Database.Type = dbType
	}

	if dbHost != "" {
		cfg.Database.Host = dbHost
	}

	if dbPort > 0 {
		cfg.Database.Port = dbPort
	}

	if dbName != "" {
		cfg.Database.Name = dbName
	}

	if dbUser != "" {
		cfg.Database.User = dbUser
	}

	if dbPassword != "" {
		cfg.Database.Password = dbPassword
	}

	if dbSSLMode != "" {
		cfg.Database.SSLMode = dbSSLMode
	}

	if jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}

	if jwtExpiry > 0 {
		cfg.JWT.TokenExpiry = jwtExpiry
	}

	if logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}
}

// GetDSN returns database connection string.
func (cfg *Config) GetDSN() string {
	if cfg.Database.Type == "postgres" {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
			cfg.Database.SSLMode,
		)
	}
	return ""
}

// GetServerAddr returns server address.
func (cfg *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

// Load is a compatibility function for backward compatibility.
func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		// Fallback to basic config if parsing fails
		return &Config{
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
		}
	}

	// Parse command line flags to override values
	cfg.ParseFlags()

	return cfg
}
