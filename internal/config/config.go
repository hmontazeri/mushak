package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig represents the mushak.yaml configuration
type AppConfig struct {
	InternalPort        int      `yaml:"internal_port"`
	HealthPath          string   `yaml:"health_path"`
	HealthTimeout       int      `yaml:"health_timeout"`
	ServiceName         string   `yaml:"service_name"`
	PersistentServices  []string `yaml:"persistent_services"`
	CacheLimit          string   `yaml:"cache_limit"` // e.g. "10GB" or "24h"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		InternalPort:  80,
		HealthPath:    "/",
		HealthTimeout: 30,
	}
}

// LoadConfig loads configuration from mushak.yaml
func LoadConfig(path string) (*AppConfig, error) {
	cfg := DefaultConfig()

	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

// SaveAppConfig saves application configuration to mushak.yaml
func SaveAppConfig(cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("mushak.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write mushak.yaml: %w", err)
	}

	return nil
}

// DeployConfig represents local deployment configuration
// Stored in .mushak/mushak.yaml
type DeployConfig struct {
	AppName    string `yaml:"app_name"`
	Host       string `yaml:"host"`
	User       string `yaml:"user"`
	Domain     string `yaml:"domain"`
	Branch     string `yaml:"branch"`
	RemoteName string `yaml:"remote_name"`
}

// SaveDeployConfig saves deployment configuration locally
func SaveDeployConfig(cfg *DeployConfig) error {
	// Create .mushak directory if it doesn't exist
	if err := os.MkdirAll(".mushak", 0755); err != nil {
		return fmt.Errorf("failed to create .mushak directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(".mushak/mushak.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadDeployConfig loads deployment configuration from .mushak/mushak.yaml
func LoadDeployConfig() (*DeployConfig, error) {
	data, err := os.ReadFile(".mushak/mushak.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read deploy config: %w", err)
	}

	var cfg DeployConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse deploy config: %w", err)
	}

	return &cfg, nil
}
