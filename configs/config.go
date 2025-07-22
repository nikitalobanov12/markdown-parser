package configs

import (
	"encoding/json"
	"os"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig    `json:"server"`
	Parser    ParserConfig    `json:"parser"`
	WebSocket WebSocketConfig `json:"websocket"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string   `json:"port"`
	Host         string   `json:"host"`
	AllowOrigins []string `json:"allow_origins"`
}

// ParserConfig holds parser configuration
type ParserConfig struct {
	MaxContentSize int64 `json:"max_content_size"`
	EnableGFM      bool  `json:"enable_gfm"`
	EnableTables   bool  `json:"enable_tables"`
	EnableAutolink bool  `json:"enable_autolink"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	MaxConnections    int   `json:"max_connections"`
	MaxMessageSize    int64 `json:"max_message_size"`
	PingPeriodSeconds int   `json:"ping_period_seconds"`
	PongWaitSeconds   int   `json:"pong_wait_seconds"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
			AllowOrigins: []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://127.0.0.1:3000",
			},
		},
		Parser: ParserConfig{
			MaxContentSize: 1024 * 1024, // 1MB
			EnableGFM:      true,
			EnableTables:   true,
			EnableAutolink: true,
		},
		WebSocket: WebSocketConfig{
			MaxConnections:    1000,
			MaxMessageSize:    512 * 1024, // 512KB
			PingPeriodSeconds: 54,
			PongWaitSeconds:   60,
		},
	}
}

// LoadConfig loads configuration from a file or returns default
func LoadConfig(filepath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Fill in any missing values with defaults
	defaultConfig := DefaultConfig()
	if config.Server.Port == "" {
		config.Server.Port = defaultConfig.Server.Port
	}
	if config.Server.Host == "" {
		config.Server.Host = defaultConfig.Server.Host
	}
	if len(config.Server.AllowOrigins) == 0 {
		config.Server.AllowOrigins = defaultConfig.Server.AllowOrigins
	}

	return &config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(filepath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}