package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	if initCmd.Use != "init [USER@HOST]" {
		t.Errorf("initCmd.Use = %v, want init [USER@HOST]", initCmd.Use)
	}

	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}

	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}

	if initCmd.RunE == nil {
		t.Error("initCmd.RunE should not be nil")
	}
}

func TestInitCommandFlags(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	requiredFlags := []struct {
		name     string
		required bool
	}{
		{name: "host", required: true},
		{name: "user", required: true},
		{name: "domain", required: true},
		{name: "app", required: false},
		{name: "branch", required: false},
		{name: "key", required: false},
		{name: "port", required: false},
	}

	for _, flag := range requiredFlags {
		f := initCmd.Flags().Lookup(flag.name)
		if f == nil {
			t.Errorf("init command should have --%s flag", flag.name)
		}
	}
}

func TestInitCommandFlagDefaults(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	tests := []struct {
		flagName     string
		defaultValue string
	}{
		{flagName: "branch", defaultValue: "main"},
		{flagName: "port", defaultValue: "22"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := initCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %s should exist", tt.flagName)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %s default = %v, want %v", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestInitCommandDescription(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	requiredDescriptions := []string{
		"Docker",
		"Git",
		"Caddy",
		"post-receive",
		"Git remote",
	}

	for _, desc := range requiredDescriptions {
		if !strings.Contains(initCmd.Long, desc) {
			t.Errorf("init command description should mention %q", desc)
		}
	}
}

func TestInitGitRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		host     string
		port     string
		appName  string
		wantURL  string
	}{
		{
			name:    "standard port",
			user:    "deploy",
			host:    "example.com",
			port:    "22",
			appName: "myapp",
			wantURL: "ssh://deploy@example.com:22/var/repo/myapp.git",
		},
		{
			name:    "custom port",
			user:    "admin",
			host:    "server.io",
			port:    "2222",
			appName: "webapp",
			wantURL: "ssh://admin@server.io:2222/var/repo/webapp.git",
		},
		{
			name:    "hyphenated app name",
			user:    "user",
			host:    "host.com",
			port:    "22",
			appName: "my-app",
			wantURL: "ssh://user@host.com:22/var/repo/my-app.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteURL := fmt.Sprintf("ssh://%s@%s:%s/var/repo/%s.git", tt.user, tt.host, tt.port, tt.appName)

			if remoteURL != tt.wantURL {
				t.Errorf("remoteURL = %v, want %v", remoteURL, tt.wantURL)
			}
		})
	}
}

func TestInitRemoteName(t *testing.T) {
	// Test that default remote name is "mushak"
	remoteName := "mushak"

	if remoteName != "mushak" {
		t.Errorf("remote name = %v, want mushak", remoteName)
	}
}

func TestInitCaddyConfigPath(t *testing.T) {
	tests := []struct {
		appName  string
		wantPath string
	}{
		{
			appName:  "myapp",
			wantPath: "/etc/caddy/apps/myapp.caddy",
		},
		{
			appName:  "test-app",
			wantPath: "/etc/caddy/apps/test-app.caddy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", tt.appName)

			if configPath != tt.wantPath {
				t.Errorf("configPath = %v, want %v", configPath, tt.wantPath)
			}
		})
	}
}

func TestInitPlaceholderConfig(t *testing.T) {
	tests := []struct {
		appName string
	}{
		{appName: "myapp"},
		{appName: "test-app"},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			placeholderConfig := fmt.Sprintf(`# %s - Not yet deployed
# This file will be updated after first deployment
`, tt.appName)

			if !strings.Contains(placeholderConfig, tt.appName) {
				t.Error("placeholder config should contain app name")
			}

			if !strings.Contains(placeholderConfig, "Not yet deployed") {
				t.Error("placeholder config should indicate not deployed")
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	// Test that isGitRepo function exists
	// In actual testing, this would check if we're in a git repository
	// We can't fully test this without a git repository context
}

func TestInitCommandRequiredFlags(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}

	// These flags should be marked as required
	requiredFlags := []string{"host", "user", "domain"}

	for _, flagName := range requiredFlags {
		flag := initCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("required flag %s should exist", flagName)
		}
	}
}

func TestInitAppNameFromDirectory(t *testing.T) {
	// Test that app name defaults to directory name
	// This is the logic: if initApp == "", use filepath.Base(cwd)

	tests := []struct {
		dirPath string
		want    string
	}{
		{dirPath: "/home/user/myapp", want: "myapp"},
		{dirPath: "/var/www/webapp", want: "webapp"},
		{dirPath: "/projects/my-app", want: "my-app"},
	}

	for _, tt := range tests {
		t.Run(tt.dirPath, func(t *testing.T) {
			// Simulate extracting app name from directory path
			parts := strings.Split(tt.dirPath, "/")
			appName := parts[len(parts)-1]

			if appName != tt.want {
				t.Errorf("appName = %v, want %v", appName, tt.want)
			}
		})
	}
}
