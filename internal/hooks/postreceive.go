package hooks

import (
	"fmt"
)

// GeneratePostReceiveHook generates the post-receive hook script
func GeneratePostReceiveHook(appName, domain, branch string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Configuration
APP_NAME="%s"
DOMAIN="%s"
DEPLOY_BRANCH="%s"
INTERNAL_PORT=80
HEALTH_PATH="/"
HEALTH_TIMEOUT=30

echo "========================================="
echo "Mushak Deployment Started"
echo "========================================="

while read oldrev newrev refname; do
    # Get the branch name
    BRANCH=$(git rev-parse --symbolic --abbrev-ref $refname)

    echo "Branch: $BRANCH"

    # Only deploy configured branch
    if [ "$BRANCH" != "$DEPLOY_BRANCH" ]; then
        echo "⚠ Skipping deployment for branch: $BRANCH (configured: $DEPLOY_BRANCH)"
        exit 0
    fi

    # Get short commit SHA
    SHA=$(git rev-parse --short $newrev)
    echo "Commit: $SHA"

    # Paths
    DEPLOY_DIR="/var/www/$APP_NAME/$SHA"
    PROJECT_NAME="mushak-$APP_NAME-$SHA"

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

    echo ""
    echo "→ Checking out code to $DEPLOY_DIR..."

    # Create deployment directory
    mkdir -p $DEPLOY_DIR

    # Checkout the code
    GIT_WORK_TREE=$DEPLOY_DIR git checkout -f $newrev

    cd $DEPLOY_DIR

    echo ""
    echo "→ Reading configuration..."

    # Read mushak.yaml if it exists
    if [ -f "mushak.yaml" ]; then
        echo "  Found mushak.yaml"

        # Simple YAML parsing (works for simple key: value pairs)
        if grep -q "internal_port:" mushak.yaml; then
            INTERNAL_PORT=$(grep "internal_port:" mushak.yaml | awk '{print $2}')
            echo "  Internal port: $INTERNAL_PORT"
        fi

        if grep -q "health_path:" mushak.yaml; then
            HEALTH_PATH=$(grep "health_path:" mushak.yaml | awk '{print $2}')
            echo "  Health path: $HEALTH_PATH"
        fi

        if grep -q "health_timeout:" mushak.yaml; then
            HEALTH_TIMEOUT=$(grep "health_timeout:" mushak.yaml | awk '{print $2}')
            echo "  Health timeout: $HEALTH_TIMEOUT"
        fi
    else
        echo "  Using defaults (internal_port=$INTERNAL_PORT, health_path=$HEALTH_PATH)"
    fi

    echo ""
    echo "→ Detecting build method..."

    # Detect docker-compose.yml or Dockerfile
    if [ -f "docker-compose.yml" ] || [ -f "docker-compose.yaml" ]; then
        echo "  Found docker-compose.yml"
        BUILD_METHOD="compose"

        COMPOSE_FILE="docker-compose.yml"
        [ -f "docker-compose.yaml" ] && COMPOSE_FILE="docker-compose.yaml"

        # Get the first service name
        SERVICE_NAME=$(grep -E "^[a-zA-Z0-9_-]+:" $COMPOSE_FILE | head -1 | sed 's/://g' | tr -d ' ')
        echo "  Service name: $SERVICE_NAME"

        # Create override file with port mapping
        cat > docker-compose.override.yml <<EOF
version: '3'
services:
  $SERVICE_NAME:
    ports:
      - "$HOST_PORT:$INTERNAL_PORT"
EOF

        echo "  Created docker-compose.override.yml"

    elif [ -f "Dockerfile" ]; then
        echo "  Found Dockerfile"
        BUILD_METHOD="dockerfile"
    else
        echo "ERROR: No Dockerfile or docker-compose.yml found" >&2
        exit 1
    fi

    echo ""
    echo "→ Building and starting containers..."

    if [ "$BUILD_METHOD" = "compose" ]; then
        # Docker Compose
        docker compose -p $PROJECT_NAME up -d --build
        CONTAINER_NAME="${PROJECT_NAME}-${SERVICE_NAME}-1"
    else
        # Dockerfile
        docker build -t $PROJECT_NAME .
        docker run -d --name $PROJECT_NAME -p $HOST_PORT:$INTERNAL_PORT $PROJECT_NAME
        CONTAINER_NAME=$PROJECT_NAME
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
            echo "Rolling back..."

            # Rollback: stop and remove the new container
            if [ "$BUILD_METHOD" = "compose" ]; then
                docker compose -p $PROJECT_NAME down
            else
                docker stop $CONTAINER_NAME || true
                docker rm $CONTAINER_NAME || true
            fi

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

    echo ""
    echo "→ Cleaning up old containers..."

    # Stop and remove old containers (exclude current SHA)
    if [ "$BUILD_METHOD" = "compose" ]; then
        # Find old compose projects for this app
        docker ps -a --format "{{.Names}}" | grep "^mushak-$APP_NAME-" | grep -v "$SHA" | while read container; do
            echo "  Stopping $container"
            # Extract project name from container name
            PROJECT=$(echo $container | sed 's/-[^-]*$//')
            docker compose -p $PROJECT down 2>/dev/null || true
        done
    else
        # Find old containers for this app
        docker ps -a --format "{{.Names}}" | grep "^mushak-$APP_NAME-" | grep -v "$SHA" | while read container; do
            echo "  Stopping $container"
            docker stop $container 2>/dev/null || true
            docker rm $container 2>/dev/null || true
        done
    fi

    # Cleanup old deployment directories (keep last 3)
    cd /var/www/$APP_NAME
    ls -t | tail -n +4 | xargs -r rm -rf

    echo ""
    echo "========================================="
    echo "✓ Deployment Successful!"
    echo "========================================="
    echo "App: $APP_NAME"
    echo "SHA: $SHA"
    echo "Port: $HOST_PORT"
    echo "URL: https://$DOMAIN"
    echo "========================================="
done
`, appName, domain, branch)
}
