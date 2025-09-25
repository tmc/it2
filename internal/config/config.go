package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Connection settings
	URL     string        `yaml:"url" env:"IT2_URL"`
	Timeout time.Duration `yaml:"timeout" env:"IT2_TIMEOUT"`

	// Output settings
	Format   string `yaml:"format" env:"IT2_FORMAT"`
	Color    bool   `yaml:"color" env:"IT2_COLOR"`
	Template string `yaml:"template" env:"IT2_TEMPLATE"`

	// Retry settings
	RetryAttempts int           `yaml:"retry_attempts" env:"IT2_RETRY_ATTEMPTS"`
	RetryDelay    time.Duration `yaml:"retry_delay" env:"IT2_RETRY_DELAY"`

	// Logging settings
	LogLevel string `yaml:"log_level" env:"IT2_LOG_LEVEL"`
	Verbose  bool   `yaml:"verbose" env:"IT2_VERBOSE"`
	Debug    bool   `yaml:"debug" env:"IT2_DEBUG"`

	// Progress indicators
	ShowProgress bool `yaml:"show_progress" env:"IT2_SHOW_PROGRESS"`

	// Authentication
	Cookie string `yaml:"cookie" env:"ITERM2_COOKIE"`
	Key    string `yaml:"key" env:"ITERM2_KEY"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		URL:           "ws://localhost:1912",
		Timeout:       5 * time.Second,
		Format:        "text",
		Color:         true,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
		LogLevel:      "info",
		Verbose:       false,
		Debug:         false,
		ShowProgress:  true,
	}
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from config file
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if err := cfg.loadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

func (c *Config) loadFromEnv() {
	if url := os.Getenv("IT2_URL"); url != "" {
		c.URL = url
	}
	if timeout := os.Getenv("IT2_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Timeout = d
		}
	}
	if format := os.Getenv("IT2_FORMAT"); format != "" {
		c.Format = format
	}
	if color := os.Getenv("IT2_COLOR"); color != "" {
		c.Color = color == "true" || color == "1"
	}
	if template := os.Getenv("IT2_TEMPLATE"); template != "" {
		c.Template = template
	}
	if retryAttempts := os.Getenv("IT2_RETRY_ATTEMPTS"); retryAttempts != "" {
		if attempts, err := time.ParseDuration(retryAttempts); err == nil {
			c.RetryAttempts = int(attempts)
		}
	}
	if retryDelay := os.Getenv("IT2_RETRY_DELAY"); retryDelay != "" {
		if d, err := time.ParseDuration(retryDelay); err == nil {
			c.RetryDelay = d
		}
	}
	if logLevel := os.Getenv("IT2_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
	if verbose := os.Getenv("IT2_VERBOSE"); verbose != "" {
		c.Verbose = verbose == "true" || verbose == "1"
	}
	if debug := os.Getenv("IT2_DEBUG"); debug != "" {
		c.Debug = debug == "true" || debug == "1"
	}
	if showProgress := os.Getenv("IT2_SHOW_PROGRESS"); showProgress != "" {
		c.ShowProgress = showProgress == "true" || showProgress == "1"
	}
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		c.Cookie = cookie
	}
	if key := os.Getenv("ITERM2_KEY"); key != "" {
		c.Key = key
	}
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".it2", "config.yaml")
	}
	return filepath.Join(homeDir, ".it2", "config.yaml")
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".it2"
	}
	return filepath.Join(homeDir, ".it2")
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	return getConfigPath()
}
