// Package config handles application configuration loading and validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"telegram-cf-bot/internal/constants"
	apperrors "telegram-cf-bot/internal/errors"
)

// Config holds all application configuration.
type Config struct {
	Telegram        TelegramConfig   `yaml:"telegram"`
	Cloudflare      CloudflareConfig `yaml:"cloudflare"`
	AuthorizedUsers []int64          `yaml:"authorized_users"`
	AdminID         int64            `yaml:"admin_id"`
	Logging         LoggingConfig    `yaml:"logging"`
	configPath      string           `yaml:"-"`
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
}

// CloudflareConfig holds Cloudflare API configuration.
type CloudflareConfig struct {
	AccountID string `yaml:"account_id"`
	APIToken  string `yaml:"api_token"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level    string `yaml:"level"`
	ToFile   bool   `yaml:"to_file"`
	FilePath string `yaml:"file_path"`
}

// Load loads configuration from file with validation.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if configPath == "" {
		configPath = findConfigFile()
	}

	if configPath == "" {
		return nil, apperrors.New(apperrors.ErrInvalidConfig, "configuration file not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidConfig, "failed to read config file", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidConfig, "failed to parse config file", err)
	}

	cfg.configPath = configPath

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = constants.DefaultLogLevel
	}
	if cfg.Logging.FilePath == "" {
		cfg.Logging.FilePath = constants.DefaultLogFilePath
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return apperrors.New(apperrors.ErrInvalidConfig, "telegram.bot_token is required")
	}

	if c.Cloudflare.AccountID == "" {
		return apperrors.New(apperrors.ErrInvalidConfig, "cloudflare.account_id is required")
	}

	if c.Cloudflare.APIToken == "" {
		return apperrors.New(apperrors.ErrInvalidConfig, "cloudflare.api_token is required")
	}

	return nil
}

// Save persists the configuration to disk.
func (c *Config) Save() error {
	if c.configPath == "" {
		return apperrors.New(apperrors.ErrInvalidConfig, "config path not set")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidConfig, "failed to marshal config", err)
	}

	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidConfig, "failed to write config file", err)
	}

	return nil
}

// GetConfigPath returns the configuration file path.
func (c *Config) GetConfigPath() string {
	return c.configPath
}

// SetConfigPath sets the configuration file path.
func (c *Config) SetConfigPath(path string) {
	c.configPath = path
}

// IsAuthorized checks if a user ID is in the authorized list or is admin.
func (c *Config) IsAuthorized(userID int64) bool {
	if userID == c.AdminID {
		return true
	}

	for _, id := range c.AuthorizedUsers {
		if id == userID {
			return true
		}
	}

	return false
}

// IsAdmin checks if a user ID is the admin.
func (c *Config) IsAdmin(userID int64) bool {
	return userID == c.AdminID
}

// AddAuthorizedUser adds a user to the authorized list.
func (c *Config) AddAuthorizedUser(userID int64) error {
	if c.IsAuthorized(userID) {
		return apperrors.New(apperrors.ErrUserAlreadyExists, fmt.Sprintf("user %d is already authorized", userID))
	}

	c.AuthorizedUsers = append(c.AuthorizedUsers, userID)
	return c.Save()
}

// RemoveAuthorizedUser removes a user from the authorized list.
func (c *Config) RemoveAuthorizedUser(userID int64) error {
	if userID == c.AdminID {
		return apperrors.New(apperrors.ErrInvalidConfig, "cannot remove admin user")
	}

	found := false
	for i, id := range c.AuthorizedUsers {
		if id == userID {
			c.AuthorizedUsers = append(c.AuthorizedUsers[:i], c.AuthorizedUsers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return apperrors.New(apperrors.ErrUserNotFound, fmt.Sprintf("user %d is not in authorized list", userID))
	}

	return c.Save()
}

// findConfigFile searches for config.yaml in common locations.
func findConfigFile() string {
	paths := []string{
		"config.yaml",
		"config.yml",
	}

	// Add executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append(paths,
			filepath.Join(execDir, "config.yaml"),
			filepath.Join(execDir, "config.yml"),
		)
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
