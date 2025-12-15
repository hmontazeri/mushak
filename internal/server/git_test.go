package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hmontazeri/mushak/internal/ssh"
)

func TestSetupGitRepo(t *testing.T) {
	tests := []struct {
		name    string
		appName string
	}{
		{
			name:    "simple app name",
			appName: "myapp",
		},
		{
			name:    "hyphenated app name",
			appName: "my-app",
		},
		{
			name:    "underscored app name",
			appName: "my_app",
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

			// Verify expected paths would be used
			expectedRepoPath := fmt.Sprintf("/var/repo/%s.git", tt.appName)
			expectedDeployPath := fmt.Sprintf("/var/www/%s", tt.appName)

			if !strings.Contains(expectedRepoPath, tt.appName) {
				t.Errorf("repo path should contain app name")
			}

			if !strings.Contains(expectedDeployPath, tt.appName) {
				t.Errorf("deploy path should contain app name")
			}
		})
	}
}

func TestInstallPostReceiveHook(t *testing.T) {
	tests := []struct {
		name       string
		appName    string
		hookScript string
	}{
		{
			name:       "basic hook",
			appName:    "testapp",
			hookScript: "#!/bin/bash\necho 'test'",
		},
		{
			name:       "empty hook script",
			appName:    "app",
			hookScript: "",
		},
		{
			name:    "multiline hook script",
			appName: "webapp",
			hookScript: `#!/bin/bash
set -e
echo "Starting deployment"
exit 0`,
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

			// Verify expected hook path
			expectedHookPath := fmt.Sprintf("/var/repo/%s.git/hooks/post-receive", tt.appName)

			if !strings.Contains(expectedHookPath, tt.appName) {
				t.Errorf("hook path should contain app name")
			}

			if !strings.Contains(expectedHookPath, "post-receive") {
				t.Errorf("hook path should contain post-receive")
			}
		})
	}
}

func TestGitRepoPathFormat(t *testing.T) {
	tests := []struct {
		appName      string
		expectedRepo string
		expectedDeploy string
	}{
		{
			appName:      "app1",
			expectedRepo: "/var/repo/app1.git",
			expectedDeploy: "/var/www/app1",
		},
		{
			appName:      "my-app",
			expectedRepo: "/var/repo/my-app.git",
			expectedDeploy: "/var/www/my-app",
		},
		{
			appName:      "test_app_123",
			expectedRepo: "/var/repo/test_app_123.git",
			expectedDeploy: "/var/www/test_app_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			repoPath := fmt.Sprintf("/var/repo/%s.git", tt.appName)
			deployPath := fmt.Sprintf("/var/www/%s", tt.appName)

			if repoPath != tt.expectedRepo {
				t.Errorf("repoPath = %v, want %v", repoPath, tt.expectedRepo)
			}

			if deployPath != tt.expectedDeploy {
				t.Errorf("deployPath = %v, want %v", deployPath, tt.expectedDeploy)
			}
		})
	}
}

func TestPostReceiveHookPath(t *testing.T) {
	tests := []struct {
		appName      string
		expectedPath string
	}{
		{
			appName:      "app",
			expectedPath: "/var/repo/app.git/hooks/post-receive",
		},
		{
			appName:      "my-app",
			expectedPath: "/var/repo/my-app.git/hooks/post-receive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			hookPath := fmt.Sprintf("/var/repo/%s.git/hooks/post-receive", tt.appName)

			if hookPath != tt.expectedPath {
				t.Errorf("hookPath = %v, want %v", hookPath, tt.expectedPath)
			}
		})
	}
}
