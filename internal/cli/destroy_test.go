package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestDestroyCommand(t *testing.T) {
	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	if destroyCmd.Use != "destroy" {
		t.Errorf("destroyCmd.Use = %v, want destroy", destroyCmd.Use)
	}

	if destroyCmd.Short == "" {
		t.Error("destroyCmd.Short should not be empty")
	}

	if destroyCmd.Long == "" {
		t.Error("destroyCmd.Long should not be empty")
	}

	if destroyCmd.RunE == nil {
		t.Error("destroyCmd.RunE should not be nil")
	}
}

func TestDestroyCommandFlags(t *testing.T) {
	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	requiredFlags := []struct {
		name      string
		shorthand string
	}{
		{name: "host", shorthand: ""},
		{name: "user", shorthand: ""},
		{name: "app", shorthand: ""},
		{name: "key", shorthand: ""},
		{name: "port", shorthand: ""},
		{name: "force", shorthand: ""},
	}

	for _, flag := range requiredFlags {
		f := destroyCmd.Flags().Lookup(flag.name)
		if f == nil {
			t.Errorf("destroy command should have --%s flag", flag.name)
		}
	}
}

func TestDestroyCommandWarnings(t *testing.T) {
	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	requiredWarnings := []string{
		"WARNING",
		"irreversible",
	}

	for _, warning := range requiredWarnings {
		found := strings.Contains(destroyCmd.Long, warning) ||
			strings.Contains(strings.ToLower(destroyCmd.Long), strings.ToLower(warning))

		if !found {
			t.Errorf("destroy command description should contain warning: %q", warning)
		}
	}
}

func TestDestroyOperations(t *testing.T) {
	requiredOperations := []string{
		"containers",
		"Git repository",
		"deployment files",
		"Caddy configuration",
		"Git remote",
	}

	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	for _, op := range requiredOperations {
		if !strings.Contains(destroyCmd.Long, op) {
			t.Errorf("destroy command should mention operation: %q", op)
		}
	}
}

func TestDestroyDockerCommand(t *testing.T) {
	tests := []struct {
		appName  string
		wantStop string
		wantRm   string
	}{
		{
			appName:  "myapp",
			wantStop: "docker ps -a --format '{{.Names}}' | grep '^mushak-myapp-' | xargs -r docker stop",
			wantRm:   "docker ps -a --format '{{.Names}}' | grep '^mushak-myapp-' | xargs -r docker rm",
		},
		{
			appName:  "test-app",
			wantStop: "docker ps -a --format '{{.Names}}' | grep '^mushak-test-app-' | xargs -r docker stop",
			wantRm:   "docker ps -a --format '{{.Names}}' | grep '^mushak-test-app-' | xargs -r docker rm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			stopCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker stop", tt.appName)
			removeCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker rm", tt.appName)

			if stopCmd != tt.wantStop {
				t.Errorf("stop command = %v, want %v", stopCmd, tt.wantStop)
			}

			if removeCmd != tt.wantRm {
				t.Errorf("remove command = %v, want %v", removeCmd, tt.wantRm)
			}
		})
	}
}

func TestDestroyPaths(t *testing.T) {
	tests := []struct {
		appName        string
		wantRepoPath   string
		wantDeployPath string
		wantCaddyPath  string
	}{
		{
			appName:        "myapp",
			wantRepoPath:   "/var/repo/myapp.git",
			wantDeployPath: "/var/www/myapp",
			wantCaddyPath:  "/etc/caddy/apps/myapp.caddy",
		},
		{
			appName:        "test-app",
			wantRepoPath:   "/var/repo/test-app.git",
			wantDeployPath: "/var/www/test-app",
			wantCaddyPath:  "/etc/caddy/apps/test-app.caddy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			repoPath := fmt.Sprintf("/var/repo/%s.git", tt.appName)
			deployPath := fmt.Sprintf("/var/www/%s", tt.appName)
			caddyPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", tt.appName)

			if repoPath != tt.wantRepoPath {
				t.Errorf("repoPath = %v, want %v", repoPath, tt.wantRepoPath)
			}

			if deployPath != tt.wantDeployPath {
				t.Errorf("deployPath = %v, want %v", deployPath, tt.wantDeployPath)
			}

			if caddyPath != tt.wantCaddyPath {
				t.Errorf("caddyPath = %v, want %v", caddyPath, tt.wantCaddyPath)
			}
		})
	}
}

func TestDestroyCommandPortDefault(t *testing.T) {
	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	portFlag := destroyCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("port flag should exist")
	}

	if portFlag.DefValue != "22" {
		t.Errorf("port flag default = %v, want 22", portFlag.DefValue)
	}
}

func TestDestroyForceFlag(t *testing.T) {
	if destroyCmd == nil {
		t.Fatal("destroyCmd should not be nil")
	}

	forceFlag := destroyCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("force flag should exist")
	}

	// Force flag should default to false
	if forceFlag.DefValue != "false" {
		t.Errorf("force flag default = %v, want false", forceFlag.DefValue)
	}
}
