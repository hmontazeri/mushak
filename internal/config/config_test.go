package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.InternalPort != 80 {
		t.Errorf("DefaultConfig().InternalPort = %v, want 80", cfg.InternalPort)
	}

	if cfg.HealthPath != "/" {
		t.Errorf("DefaultConfig().HealthPath = %v, want /", cfg.HealthPath)
	}

	if cfg.HealthTimeout != 30 {
		t.Errorf("DefaultConfig().HealthTimeout = %v, want 30", cfg.HealthTimeout)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/mushak.yaml")
	if err != nil {
		t.Errorf("LoadConfig() with non-existent file error = %v, want nil", err)
		return
	}

	// Should return default config
	if cfg.InternalPort != 80 {
		t.Errorf("LoadConfig().InternalPort = %v, want 80", cfg.InternalPort)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mushak.yaml")

	configContent := `internal_port: 3000
health_path: /health
health_timeout: 60
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.InternalPort != 3000 {
		t.Errorf("LoadConfig().InternalPort = %v, want 3000", cfg.InternalPort)
	}

	if cfg.HealthPath != "/health" {
		t.Errorf("LoadConfig().HealthPath = %v, want /health", cfg.HealthPath)
	}

	if cfg.HealthTimeout != 60 {
		t.Errorf("LoadConfig().HealthTimeout = %v, want 60", cfg.HealthTimeout)
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	// Create a temporary config file with only some fields
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mushak.yaml")

	configContent := `internal_port: 8080
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.InternalPort != 8080 {
		t.Errorf("LoadConfig().InternalPort = %v, want 8080", cfg.InternalPort)
	}

	// Should use defaults for unspecified fields
	if cfg.HealthPath != "/" {
		t.Errorf("LoadConfig().HealthPath = %v, want / (default)", cfg.HealthPath)
	}

	if cfg.HealthTimeout != 30 {
		t.Errorf("LoadConfig().HealthTimeout = %v, want 30 (default)", cfg.HealthTimeout)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mushak.yaml")

	configContent := `internal_port: invalid
health_path: [unclosed
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() with invalid YAML should return error")
	}
}

func TestSaveDeployConfig(t *testing.T) {
	// Change to temp directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	cfg := &DeployConfig{
		AppName:    "testapp",
		Host:       "example.com",
		User:       "testuser",
		Domain:     "test.example.com",
		Branch:     "main",
		RemoteName: "mushak",
	}

	err := SaveDeployConfig(cfg)
	if err != nil {
		t.Fatalf("SaveDeployConfig() error = %v", err)
	}

	// Check that directory was created
	if _, err := os.Stat(".mushak"); os.IsNotExist(err) {
		t.Error(".mushak directory was not created")
	}

	// Check that config file was created
	if _, err := os.Stat(".mushak/mushak.yaml"); os.IsNotExist(err) {
		t.Error(".mushak/mushak.yaml was not created")
	}

	// Read and verify config
	data, err := os.ReadFile(".mushak/mushak.yaml")
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var loadedCfg DeployConfig
	err = yaml.Unmarshal(data, &loadedCfg)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if loadedCfg.AppName != cfg.AppName {
		t.Errorf("Saved AppName = %v, want %v", loadedCfg.AppName, cfg.AppName)
	}

	if loadedCfg.Host != cfg.Host {
		t.Errorf("Saved Host = %v, want %v", loadedCfg.Host, cfg.Host)
	}

	if loadedCfg.User != cfg.User {
		t.Errorf("Saved User = %v, want %v", loadedCfg.User, cfg.User)
	}

	if loadedCfg.Domain != cfg.Domain {
		t.Errorf("Saved Domain = %v, want %v", loadedCfg.Domain, cfg.Domain)
	}

	if loadedCfg.Branch != cfg.Branch {
		t.Errorf("Saved Branch = %v, want %v", loadedCfg.Branch, cfg.Branch)
	}

	if loadedCfg.RemoteName != cfg.RemoteName {
		t.Errorf("Saved RemoteName = %v, want %v", loadedCfg.RemoteName, cfg.RemoteName)
	}
}

func TestLoadDeployConfig(t *testing.T) {
	// Change to temp directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create config directory and file
	os.MkdirAll(".mushak", 0755)

	configContent := `app_name: myapp
host: server.example.com
user: deployuser
domain: myapp.example.com
branch: production
remote_name: origin
`

	err := os.WriteFile(".mushak/mushak.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadDeployConfig()
	if err != nil {
		t.Fatalf("LoadDeployConfig() error = %v", err)
	}

	if cfg.AppName != "myapp" {
		t.Errorf("LoadDeployConfig().AppName = %v, want myapp", cfg.AppName)
	}

	if cfg.Host != "server.example.com" {
		t.Errorf("LoadDeployConfig().Host = %v, want server.example.com", cfg.Host)
	}

	if cfg.User != "deployuser" {
		t.Errorf("LoadDeployConfig().User = %v, want deployuser", cfg.User)
	}

	if cfg.Domain != "myapp.example.com" {
		t.Errorf("LoadDeployConfig().Domain = %v, want myapp.example.com", cfg.Domain)
	}

	if cfg.Branch != "production" {
		t.Errorf("LoadDeployConfig().Branch = %v, want production", cfg.Branch)
	}

	if cfg.RemoteName != "origin" {
		t.Errorf("LoadDeployConfig().RemoteName = %v, want origin", cfg.RemoteName)
	}
}

func TestLoadDeployConfig_NonExistent(t *testing.T) {
	// Change to temp directory with no config
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	_, err := LoadDeployConfig()
	if err == nil {
		t.Error("LoadDeployConfig() should return error when config doesn't exist")
	}
}

func TestSaveAndLoadDeployConfig_RoundTrip(t *testing.T) {
	// Change to temp directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	originalCfg := &DeployConfig{
		AppName:    "roundtrip",
		Host:       "roundtrip.example.com",
		User:       "rtuser",
		Domain:     "rt.example.com",
		Branch:     "master",
		RemoteName: "upstream",
	}

	// Save
	err := SaveDeployConfig(originalCfg)
	if err != nil {
		t.Fatalf("SaveDeployConfig() error = %v", err)
	}

	// Load
	loadedCfg, err := LoadDeployConfig()
	if err != nil {
		t.Fatalf("LoadDeployConfig() error = %v", err)
	}

	// Compare
	if loadedCfg.AppName != originalCfg.AppName {
		t.Errorf("AppName = %v, want %v", loadedCfg.AppName, originalCfg.AppName)
	}

	if loadedCfg.Host != originalCfg.Host {
		t.Errorf("Host = %v, want %v", loadedCfg.Host, originalCfg.Host)
	}

	if loadedCfg.User != originalCfg.User {
		t.Errorf("User = %v, want %v", loadedCfg.User, originalCfg.User)
	}

	if loadedCfg.Domain != originalCfg.Domain {
		t.Errorf("Domain = %v, want %v", loadedCfg.Domain, originalCfg.Domain)
	}

	if loadedCfg.Branch != originalCfg.Branch {
		t.Errorf("Branch = %v, want %v", loadedCfg.Branch, originalCfg.Branch)
	}

	if loadedCfg.RemoteName != originalCfg.RemoteName {
		t.Errorf("RemoteName = %v, want %v", loadedCfg.RemoteName, originalCfg.RemoteName)
	}
}
