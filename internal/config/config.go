package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	MemosURL   string `mapstructure:"memos_url"`
	MemosToken string `mapstructure:"memos_token"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file locations
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default to ~/.flowgres2memos/config.yaml
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		configDir := filepath.Join(home, ".flowgres2memos")
		v.AddConfigPath(configDir)
		v.AddConfigPath(".") // Also check current directory
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read environment variables
	v.SetEnvPrefix("FLOWGRES2MEMOS")
	v.AutomaticEnv()

	// Bind environment variables to config keys
	v.BindEnv("memos_url", "MEMOS_URL")
	v.BindEnv("memos_token", "MEMOS_TOKEN")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error if env vars are set
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if all required configuration fields are set
func (c *Config) Validate() error {
	if c.MemosURL == "" {
		return fmt.Errorf("memos_url is required (set via config file or MEMOS_URL env var)")
	}
	if c.MemosToken == "" {
		return fmt.Errorf("memos_token is required (set via config file or MEMOS_TOKEN env var)")
	}
	return nil
}

// GetConfigDir returns the default config directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".flowgres2memos"), nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}
