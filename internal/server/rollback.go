package server

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
)

// DeploymentVersion represents a deployed version available for rollback
type DeploymentVersion struct {
	SHA        string
	Timestamp  string
	Port       string
	Method     string
	HasImage   bool
	HasDir     bool
	IsCurrent  bool
}

// ListVersions lists available versions for rollback
func ListVersions(executor *ssh.Executor, appName string) ([]DeploymentVersion, error) {
	// Get current deployment SHA
	currentSHA, err := getCurrentSHA(executor, appName)
	if err != nil {
		currentSHA = ""
	}

	// Read deployment manifest
	manifestCmd := fmt.Sprintf("cat /var/www/%s/.deployments 2>/dev/null || echo ''", appName)
	manifest, _ := executor.Run(manifestCmd)

	// Get available tagged images
	imageCmd := fmt.Sprintf("docker images mushak-%s --format '{{.Tag}}' 2>/dev/null | grep -v latest || echo ''", appName)
	imagesOutput, _ := executor.Run(imageCmd)
	availableImages := make(map[string]bool)
	for _, tag := range strings.Split(strings.TrimSpace(imagesOutput), "\n") {
		if tag != "" {
			availableImages[tag] = true
		}
	}

	// Get available deployment directories
	dirCmd := fmt.Sprintf("ls -d /var/www/%s/*/ 2>/dev/null | xargs -I {} basename {} | grep -v current || echo ''", appName)
	dirsOutput, _ := executor.Run(dirCmd)
	availableDirs := make(map[string]bool)
	for _, dir := range strings.Split(strings.TrimSpace(dirsOutput), "\n") {
		if dir != "" && dir != "current" {
			availableDirs[dir] = true
		}
	}

	// Parse manifest and build version list
	var versions []DeploymentVersion
	seenSHAs := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(manifest))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		sha := parts[0]
		if seenSHAs[sha] {
			continue // Skip duplicates, keep latest entry
		}
		seenSHAs[sha] = true

		version := DeploymentVersion{
			SHA:       sha,
			Timestamp: parts[1],
			HasImage:  availableImages[sha],
			HasDir:    availableDirs[sha],
			IsCurrent: sha == currentSHA,
		}

		if len(parts) >= 3 {
			version.Port = parts[2]
		}
		if len(parts) >= 4 {
			version.Method = parts[3]
		}

		// Only include versions that have an image (can be rolled back to)
		if version.HasImage {
			versions = append(versions, version)
		}
	}

	// Reverse to show newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return versions, nil
}

// getCurrentSHA gets the SHA of the currently deployed version
func getCurrentSHA(executor *ssh.Executor, appName string) (string, error) {
	// Try to get from running container
	cmd := fmt.Sprintf("docker ps --filter 'name=mushak-%s-' --format '{{.Names}}' | head -1 | sed 's/mushak-%s-//' | cut -d'-' -f1", appName, appName)
	sha, err := executor.Run(cmd)
	if err == nil && strings.TrimSpace(sha) != "" {
		return strings.TrimSpace(sha), nil
	}

	// Fallback: check the current symlink
	cmd = fmt.Sprintf("readlink /var/www/%s/current 2>/dev/null | xargs basename", appName)
	sha, err = executor.Run(cmd)
	if err == nil && strings.TrimSpace(sha) != "" {
		return strings.TrimSpace(sha), nil
	}

	return "", fmt.Errorf("could not determine current deployment")
}

// ExecuteRollback performs a rollback to the specified SHA
func ExecuteRollback(executor *ssh.Executor, cfg *config.DeployConfig, targetSHA string) error {
	appName := cfg.AppName
	domain := cfg.Domain

	ui.PrintInfo(fmt.Sprintf("Rolling back to version %s...", targetSHA))

	// Verify the target image exists
	checkCmd := fmt.Sprintf("docker images mushak-%s:%s -q", appName, targetSHA)
	imageID, err := executor.Run(checkCmd)
	if err != nil || strings.TrimSpace(imageID) == "" {
		return fmt.Errorf("image not found for SHA %s. Cannot rollback", targetSHA)
	}

	// Check if deployment directory exists
	deployDir := fmt.Sprintf("/var/www/%s/%s", appName, targetSHA)
	dirCheckCmd := fmt.Sprintf("test -d %s && echo 'exists'", deployDir)
	dirExists, _ := executor.Run(dirCheckCmd)

	if strings.TrimSpace(dirExists) != "exists" {
		return fmt.Errorf("deployment directory not found for SHA %s. Cannot rollback", targetSHA)
	}

	// Execute rollback script on server
	rollbackScript := generateRollbackScript(appName, domain, targetSHA)

	fmt.Println("----------------------------------------")
	if err := executor.StreamRun(rollbackScript, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	fmt.Println("----------------------------------------")

	return nil
}

// generateRollbackScript generates a bash script to perform the rollback
func generateRollbackScript(appName, domain, targetSHA string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

APP_NAME="%s"
DOMAIN="%s"
TARGET_SHA="%s"
DEPLOY_DIR="/var/www/$APP_NAME/$TARGET_SHA"
CURRENT_LINK="/var/www/$APP_NAME/current"
PROJECT_NAME="mushak-$APP_NAME-$TARGET_SHA"
IMAGE_REPO="mushak-$APP_NAME"
NETWORK_NAME="mushak-${APP_NAME}-net"

echo "========================================="
echo "Mushak Rollback Started"
echo "========================================="
echo "App: $APP_NAME"
echo "Target: $TARGET_SHA"

echo ""
echo "→ Finding available port..."

# Find a free port (8000-9000)
find_free_port() {
    for port in {8000..9000}; do
        if ! ss -tuln | grep -q ":$port "; then
            echo $port
            return 0
        fi
    done
    echo "ERROR: No free ports available in range 8000-9000" >&2
    exit 1
}

HOST_PORT=$(find_free_port)
echo "  Using port: $HOST_PORT"

cd "$DEPLOY_DIR"

# Ensure network exists
docker network create $NETWORK_NAME 2>/dev/null || true

echo ""
echo "→ Detecting deployment method..."

# Detect if this was a compose or dockerfile deployment
if [ -f "docker-compose.yml" ] || [ -f "docker-compose.yaml" ]; then
    BUILD_METHOD="compose"
    COMPOSE_FILE="docker-compose.yml"
    [ -f "docker-compose.yaml" ] && COMPOSE_FILE="docker-compose.yaml"
    
    # Get the web service name
    WEB_SERVICE=$(grep -E "^  [a-zA-Z0-9_-]*web[a-zA-Z0-9_-]*:" $COMPOSE_FILE | head -1 | sed 's/://g' | tr -d ' ')
    if [ -z "$WEB_SERVICE" ]; then
        WEB_SERVICE=$(grep -E "^  [a-zA-Z0-9_-]+:" $COMPOSE_FILE | head -1 | sed 's/://g' | tr -d ' ')
    fi
    SERVICE_NAME=$WEB_SERVICE
    CONTAINER_NAME="${PROJECT_NAME}-${SERVICE_NAME}"
    echo "  Method: docker-compose"
    echo "  Service: $SERVICE_NAME"
else
    BUILD_METHOD="dockerfile"
    CONTAINER_NAME="$PROJECT_NAME"
    echo "  Method: Dockerfile"
fi

# Read internal port from mushak.yaml if exists
INTERNAL_PORT=80
HEALTH_PATH="/"
HEALTH_TIMEOUT=30

if [ -f "mushak.yaml" ]; then
    if grep -q "internal_port:" mushak.yaml; then
        INTERNAL_PORT=$(grep "internal_port:" mushak.yaml | awk '{print $2}')
    fi
    if grep -q "health_path:" mushak.yaml; then
        HEALTH_PATH=$(grep "health_path:" mushak.yaml | awk '{print $2}')
    fi
    if grep -q "health_timeout:" mushak.yaml; then
        HEALTH_TIMEOUT=$(grep "health_timeout:" mushak.yaml | awk '{print $2}')
    fi
fi

echo ""
echo "→ Starting container from cached image..."

# Stop any existing container with the same name
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

if [ "$BUILD_METHOD" = "compose" ]; then
    # For compose, we need to recreate the override file with the new port
    cat > docker-compose.override.yml <<EOF
version: '3'
networks:
  default:
    external: true
    name: $NETWORK_NAME
services:
  $SERVICE_NAME:
    image: ${IMAGE_REPO}:${TARGET_SHA}
    container_name: ${CONTAINER_NAME}
    ports:
      - "$HOST_PORT:$INTERNAL_PORT"
EOF

    # Start the service using the pre-built image
    docker compose -p $PROJECT_NAME up -d --no-build $SERVICE_NAME
else
    # For Dockerfile deployments, run the tagged image
    ENV_OPTS=""
    if [ -f ".env" ]; then
        ENV_OPTS="--env-file .env"
    fi
    
    docker run -d --name "$CONTAINER_NAME" --network "$NETWORK_NAME" $ENV_OPTS -p $HOST_PORT:$INTERNAL_PORT "${IMAGE_REPO}:${TARGET_SHA}"
fi

echo "  Container started: $CONTAINER_NAME"

echo ""
echo "→ Waiting for service to be healthy..."

# Health check with retry
RETRY_COUNT=0
until curl -sf http://localhost:$HOST_PORT$HEALTH_PATH > /dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    
    if [ $RETRY_COUNT -ge $HEALTH_TIMEOUT ]; then
        echo ""
        echo "ERROR: Health check failed after $HEALTH_TIMEOUT seconds" >&2
        echo "Aborting rollback..."
        
        # Cleanup failed rollback container
        docker stop "$CONTAINER_NAME" 2>/dev/null || true
        docker rm "$CONTAINER_NAME" 2>/dev/null || true
        
        exit 1
    fi
    
    echo -n "."
    sleep 1
done

echo ""
echo "  Service is healthy!"

echo ""
echo "→ Updating Caddy configuration..."

# Update Caddy config
sudo tee /etc/caddy/apps/$APP_NAME.caddy > /dev/null <<EOF
$DOMAIN {
	reverse_proxy localhost:$HOST_PORT
}
EOF

# Reload Caddy
sudo systemctl reload caddy

echo "  Caddy updated and reloaded"

# Update current symlink
ln -snf "$DEPLOY_DIR" "$CURRENT_LINK"

echo ""
echo "→ Stopping old containers..."

# Stop old containers (exclude the rollback target)
docker ps -a --format "{{.Names}}" | grep "^mushak-$APP_NAME-" | grep -v "$TARGET_SHA" | grep -v "_" | while read container; do
    echo "  Stopping $container"
    docker stop "$container" 2>/dev/null || true
    docker rm "$container" 2>/dev/null || true
done

# Record rollback in deployment manifest
DEPLOYMENTS_FILE="/var/www/$APP_NAME/.deployments"
echo "${TARGET_SHA} $(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ) ${HOST_PORT} rollback" >> "$DEPLOYMENTS_FILE"

echo ""
echo "========================================="
echo "✓ Rollback Successful!"
echo "========================================="
echo "App: $APP_NAME"
echo "SHA: $TARGET_SHA"
echo "Port: $HOST_PORT"
echo "URL: https://$DOMAIN"
echo "========================================="
`, appName, domain, targetSHA)
}

