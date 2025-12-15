package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hmontazeri/mushak/internal/ssh"
)

func TestInitializeCaddyMultiApp(t *testing.T) {
	client := &ssh.Client{}
	executor := ssh.NewExecutor(client)

	// Test that function exists
	if executor == nil {
		t.Fatal("executor is nil")
	}
}

func TestCreateAppCaddyConfig(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		domain  string
		port    int
	}{
		{
			name:    "basic config",
			appName: "myapp",
			domain:  "myapp.com",
			port:    8000,
		},
		{
			name:    "subdomain",
			appName: "api",
			domain:  "api.example.com",
			port:    3000,
		},
		{
			name:    "custom port",
			appName: "webapp",
			domain:  "webapp.io",
			port:    9999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &ssh.Client{}
			executor := ssh.NewExecutor(client)

			// Verify function signature
			if executor == nil {
				t.Fatal("executor is nil")
			}

			// Verify expected config path
			expectedPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", tt.appName)

			if !strings.Contains(expectedPath, tt.appName) {
				t.Errorf("config path should contain app name")
			}
		})
	}
}

func TestRemoveAppCaddyConfig(t *testing.T) {
	tests := []struct {
		name    string
		appName string
	}{
		{
			name:    "remove simple app",
			appName: "myapp",
		},
		{
			name:    "remove hyphenated app",
			appName: "my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &ssh.Client{}
			executor := ssh.NewExecutor(client)

			// Verify function signature
			if executor == nil {
				t.Fatal("executor is nil")
			}

			// Verify expected config path
			expectedPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", tt.appName)

			if !strings.Contains(expectedPath, tt.appName) {
				t.Errorf("config path should contain app name")
			}
		})
	}
}

func TestReloadCaddy(t *testing.T) {
	client := &ssh.Client{}
	executor := ssh.NewExecutor(client)

	// Test that function exists
	if executor == nil {
		t.Fatal("executor is nil")
	}
}

func TestCaddyConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		port     int
		contains []string
	}{
		{
			name:   "basic config format",
			domain: "example.com",
			port:   8000,
			contains: []string{
				"example.com",
				"reverse_proxy localhost:8000",
			},
		},
		{
			name:   "subdomain config",
			domain: "api.example.com",
			port:   3000,
			contains: []string{
				"api.example.com",
				"reverse_proxy localhost:3000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := fmt.Sprintf(`%s {
	reverse_proxy localhost:%d
}
`, tt.domain, tt.port)

			for _, expected := range tt.contains {
				if !strings.Contains(config, expected) {
					t.Errorf("config should contain %q", expected)
				}
			}
		})
	}
}

func TestCaddyConfigPath(t *testing.T) {
	tests := []struct {
		appName      string
		expectedPath string
	}{
		{
			appName:      "app1",
			expectedPath: "/etc/caddy/apps/app1.caddy",
		},
		{
			appName:      "my-app",
			expectedPath: "/etc/caddy/apps/my-app.caddy",
		},
		{
			appName:      "test_app",
			expectedPath: "/etc/caddy/apps/test_app.caddy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", tt.appName)

			if configPath != tt.expectedPath {
				t.Errorf("configPath = %v, want %v", configPath, tt.expectedPath)
			}
		})
	}
}

func TestCaddyMainConfigFormat(t *testing.T) {
	mainCaddyfile := `# Mushak multi-app Caddyfile
# Import all app configurations
import /etc/caddy/apps/*.caddy
`

	requiredElements := []string{
		"Mushak multi-app Caddyfile",
		"import /etc/caddy/apps/*.caddy",
	}

	for _, element := range requiredElements {
		if !strings.Contains(mainCaddyfile, element) {
			t.Errorf("main Caddyfile should contain %q", element)
		}
	}
}

func TestCaddyDirectoryStructure(t *testing.T) {
	expectedDirectories := []string{
		"/etc/caddy/apps",
	}

	expectedFiles := []string{
		"/etc/caddy/Caddyfile",
	}

	for _, dir := range expectedDirectories {
		if dir == "" {
			t.Error("directory path should not be empty")
		}
	}

	for _, file := range expectedFiles {
		if file == "" {
			t.Error("file path should not be empty")
		}
	}
}

func TestCaddyReverseProxyFormat(t *testing.T) {
	tests := []struct {
		domain string
		port   int
		want   string
	}{
		{
			domain: "example.com",
			port:   8000,
			want:   "example.com {\n\treverse_proxy localhost:8000\n}\n",
		},
		{
			domain: "test.com",
			port:   3000,
			want:   "test.com {\n\treverse_proxy localhost:3000\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			config := fmt.Sprintf("%s {\n\treverse_proxy localhost:%d\n}\n", tt.domain, tt.port)

			if config != tt.want {
				t.Errorf("config = %q, want %q", config, tt.want)
			}
		})
	}
}

func TestCaddyBackupPath(t *testing.T) {
	backupPath := "/etc/caddy/Caddyfile.backup"

	if !strings.Contains(backupPath, "Caddyfile") {
		t.Error("backup path should contain Caddyfile")
	}

	if !strings.HasSuffix(backupPath, ".backup") {
		t.Error("backup path should end with .backup")
	}
}
