package hooks

import (
	"strings"
	"testing"
)

func TestGeneratePostReceiveHook(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		domain   string
		branch   string
		contains []string
	}{
		{
			name:    "basic hook generation",
			appName: "myapp",
			domain:  "myapp.example.com",
			branch:  "main",
			contains: []string{
				"#!/bin/bash",
				"APP_NAME=\"myapp\"",
				"DOMAIN=\"myapp.example.com\"",
				"DEPLOY_BRANCH=\"main\"",
				"Mushak Deployment Started",
			},
		},
		{
			name:    "production branch",
			appName: "webapp",
			domain:  "webapp.com",
			branch:  "production",
			contains: []string{
				"#!/bin/bash",
				"APP_NAME=\"webapp\"",
				"DOMAIN=\"webapp.com\"",
				"DEPLOY_BRANCH=\"production\"",
			},
		},
		{
			name:    "custom port configuration",
			appName: "api",
			domain:  "api.example.com",
			branch:  "develop",
			contains: []string{
				"INTERNAL_PORT=80",
				"HEALTH_PATH=\"/\"",
				"HEALTH_TIMEOUT=30",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := GeneratePostReceiveHook(tt.appName, tt.domain, tt.branch)

			if script == "" {
				t.Error("GeneratePostReceiveHook() returned empty script")
				return
			}

			// Check that script contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(script, expected) {
					t.Errorf("Generated script doesn't contain expected string: %q", expected)
				}
			}
		})
	}
}

func TestGeneratePostReceiveHook_ScriptStructure(t *testing.T) {
	appName := "testapp"
	domain := "test.example.com"
	branch := "main"

	script := GeneratePostReceiveHook(appName, domain, branch)

	// Test that script has proper bash structure
	requiredElements := []string{
		"#!/bin/bash",
		"set -e",
		"while read oldrev newrev refname",
		"done",
	}

	for _, element := range requiredElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing required bash element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_DeploymentSteps(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	// Check that deployment steps are included
	deploymentSteps := []string{
		"Finding available port",
		"Checking out code",
		"Reading configuration",
		"Detecting build method",
		"Building and starting containers",
		"Waiting for service to be healthy",
		"Updating Caddy configuration",
		"Cleaning up old containers",
		"Deployment Successful",
	}

	for _, step := range deploymentSteps {
		if !strings.Contains(script, step) {
			t.Errorf("Script missing deployment step: %q", step)
		}
	}
}

func TestGeneratePostReceiveHook_DockerSupport(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	// Check for Docker Compose support
	dockerComposeElements := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"docker compose",
		"docker-compose.override.yml",
	}

	for _, element := range dockerComposeElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing Docker Compose element: %q", element)
		}
	}

	// Check for Dockerfile support
	dockerfileElements := []string{
		"Dockerfile",
		"docker build",
		"docker run",
	}

	for _, element := range dockerfileElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing Dockerfile element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_HealthCheck(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	healthCheckElements := []string{
		"Waiting for service to be healthy",
		"curl -sf",
		"Health check failed",
		"Service is healthy",
		"HEALTH_TIMEOUT",
		"RETRY_COUNT",
	}

	for _, element := range healthCheckElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing health check element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_Rollback(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	rollbackElements := []string{
		"Rolling back",
		"docker stop",
		"docker rm",
		"docker compose -p $PROJECT_NAME down",
	}

	for _, element := range rollbackElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing rollback element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_CaddyIntegration(t *testing.T) {
	script := GeneratePostReceiveHook("myapp", "myapp.com", "main")

	caddyElements := []string{
		"Updating Caddy configuration",
		"/etc/caddy/apps/$APP_NAME.caddy",
		"reverse_proxy localhost:$HOST_PORT",
		"systemctl reload caddy",
		"myapp.com",
	}

	for _, element := range caddyElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing Caddy element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_Cleanup(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	cleanupElements := []string{
		"Cleaning up old containers",
		"grep -v \"$SHA\"",
		"ls -t | tail -n +4",
	}

	for _, element := range cleanupElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing cleanup element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_NetworkAndInfra(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	expectedElements := []string{
		"mushak-${APP_NAME}-net",
		"docker network create",
		"external: true",
		"mushak-${APP_NAME}-infra",
		"--no-deps",
		"external_links:",
		"${APP_NAME}_${infra_svc}:${infra_svc}",
	}

	// This test assumes detailed knowledge of how internal variables are set in the script
	// Ideally we would mock the docker-compose.yml presence and content, but since we are just checking string generation:
	
	// We check for the presence of the logic blocks
	for _, element := range expectedElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing network/infra element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_BranchFiltering(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "production")

	branchElements := []string{
		"BRANCH=$(git rev-parse --symbolic --abbrev-ref $refname)",
		"if [ \"$BRANCH\" != \"$DEPLOY_BRANCH\" ]",
		"Skipping deployment for branch",
	}

	for _, element := range branchElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing branch filtering element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_ConfigurationParsing(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	configElements := []string{
		"mushak.yaml",
		"internal_port:",
		"health_path:",
		"health_timeout:",
		"Using defaults",
	}

	for _, element := range configElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing configuration parsing element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_PortManagement(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	portElements := []string{
		"Finding available port",
		"find_free_port()",
		"for port in {8000..9000}",
		"ss -tuln",
		"No free ports available",
		"HOST_PORT=",
	}

	for _, element := range portElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing port management element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_PathsAndDirectories(t *testing.T) {
	script := GeneratePostReceiveHook("testapp", "test.com", "main")

	pathElements := []string{
		"DEPLOY_DIR=\"/var/www/$APP_NAME",
		"PROJECT_NAME=\"mushak-$APP_NAME",
		"GIT_WORK_TREE=$DEPLOY_DIR git checkout",
		"mkdir -p $DEPLOY_DIR",
	}

	for _, element := range pathElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing path/directory element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_ErrorHandling(t *testing.T) {
	script := GeneratePostReceiveHook("app", "app.com", "main")

	errorElements := []string{
		"set -e",
		"ERROR:",
		"exit 1",
		">&2", // stderr redirect
	}

	for _, element := range errorElements {
		if !strings.Contains(script, element) {
			t.Errorf("Script missing error handling element: %q", element)
		}
	}
}

func TestGeneratePostReceiveHook_EmptyInputs(t *testing.T) {
	// Test that function handles empty inputs gracefully
	script := GeneratePostReceiveHook("", "", "")

	// Should still generate a script structure even with empty inputs
	if script == "" {
		t.Error("GeneratePostReceiveHook() with empty inputs returned empty script")
	}

	if !strings.Contains(script, "#!/bin/bash") {
		t.Error("Generated script doesn't have bash shebang")
	}
}

func TestGeneratePostReceiveHook_SpecialCharacters(t *testing.T) {
	// Test with special characters that might break bash
	tests := []struct {
		name    string
		appName string
		domain  string
		branch  string
	}{
		{
			name:    "hyphenated app name",
			appName: "my-app",
			domain:  "my-app.com",
			branch:  "main",
		},
		{
			name:    "underscored app name",
			appName: "my_app",
			domain:  "my-app.com",
			branch:  "feature-branch",
		},
		{
			name:    "subdomain",
			appName: "app",
			domain:  "app.staging.example.com",
			branch:  "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := GeneratePostReceiveHook(tt.appName, tt.domain, tt.branch)

			if script == "" {
				t.Error("GeneratePostReceiveHook() returned empty script")
			}

			// Verify the values are properly embedded
			if !strings.Contains(script, tt.appName) {
				t.Errorf("Script doesn't contain app name: %s", tt.appName)
			}

			if !strings.Contains(script, tt.domain) {
				t.Errorf("Script doesn't contain domain: %s", tt.domain)
			}

			if !strings.Contains(script, tt.branch) {
				t.Errorf("Script doesn't contain branch: %s", tt.branch)
			}
		})
	}
}
