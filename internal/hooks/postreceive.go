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

    # Copy environment file (try .env.prod first, then .env)
    if [ -f "/var/www/$APP_NAME/.env.prod" ]; then
        echo "→ Loading environment variables from .env.prod..."
        cp "/var/www/$APP_NAME/.env.prod" .env.prod
        # Also copy to .env for compatibility
        cp "/var/www/$APP_NAME/.env.prod" .env
    elif [ -f "/var/www/$APP_NAME/.env" ]; then
        echo "→ Loading environment variables from .env..."
        cp "/var/www/$APP_NAME/.env" .env
    else
        echo "⚠ No .env.prod or .env file found. Use 'mushak env set' to configure environment variables."
    fi

    echo ""
    echo "→ Reading configuration..."

    # Read mushak.yaml if it exists
    CUSTOM_PERSISTENT_SERVICES=""
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

        # Read persistent_services array (simple parsing)
        if grep -q "persistent_services:" mushak.yaml; then
            CUSTOM_PERSISTENT_SERVICES=$(grep -A 10 "persistent_services:" mushak.yaml | grep "^  - " | sed 's/^  - //' | tr '\n' ' ')
            echo "  Persistent services: $CUSTOM_PERSISTENT_SERVICES"
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

        # Get the service name - prefer one with 'web' in the name, otherwise use first
        WEB_SERVICE=$(grep -E "^  [a-zA-Z0-9_-]*web[a-zA-Z0-9_-]*:" $COMPOSE_FILE | head -1 | sed 's/://g' | tr -d ' ')

        if [ -n "$WEB_SERVICE" ]; then
            SERVICE_NAME=$WEB_SERVICE
            echo "  Service name: $SERVICE_NAME (detected web service)"
        else
            SERVICE_NAME=$(grep -E "^  [a-zA-Z0-9_-]+:" $COMPOSE_FILE | head -1 | sed 's/://g' | tr -d ' ')
            echo "  Service name: $SERVICE_NAME (using first service - no 'web' service found)"
        fi

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
        # Detect infrastructure services (databases, caches, etc.) that should persist
        # These services won't be restarted on redeployments
        INFRASTRUCTURE_IMAGES="postgres|mysql|mariadb|mongodb|mongo|redis|memcached|rabbitmq|elasticsearch|timescale"

        # Get list of all services
        ALL_SERVICES=$(docker compose -p $PROJECT_NAME config --services 2>/dev/null || grep -E "^  [a-zA-Z0-9_-]+:" $COMPOSE_FILE | sed 's/://g' | tr -d ' ')

        # Identify infrastructure vs application services
        INFRA_SERVICES=""
        APP_SERVICES=""

        for service in $ALL_SERVICES; do
            # Check if service is in custom persistent list
            IS_CUSTOM_PERSISTENT=0
            for custom_svc in $CUSTOM_PERSISTENT_SERVICES; do
                if [ "$service" = "$custom_svc" ]; then
                    IS_CUSTOM_PERSISTENT=1
                    break
                fi
            done

            if [ $IS_CUSTOM_PERSISTENT -eq 1 ]; then
                INFRA_SERVICES="$INFRA_SERVICES $service"
            else
                # Check if service uses an infrastructure image
                IMAGE=$(grep -A 5 "^  ${service}:" $COMPOSE_FILE | grep "image:" | head -1 | awk '{print $2}' | cut -d: -f1)

                if echo "$IMAGE" | grep -qE "$INFRASTRUCTURE_IMAGES"; then
                    INFRA_SERVICES="$INFRA_SERVICES $service"
                else
                    APP_SERVICES="$APP_SERVICES $service"
                fi
            fi
        done

        echo "  Infrastructure services: ${INFRA_SERVICES:-none}"
        echo "  Application services: ${APP_SERVICES:-all}"

        # Check if docker-compose has custom container names
        HAS_CUSTOM_NAMES=$(grep -E "^\s*container_name:" docker-compose.yml docker-compose.yaml 2>/dev/null || true)

        if [ -n "$HAS_CUSTOM_NAMES" ]; then
            echo "  ⚠ Warning: Custom container_name detected."
            echo "  Stopping previous application containers..."

            # Only stop application services, keep infrastructure running
            docker ps -a --format "{{.Names}}" | grep -E "^(mushak-)?${APP_NAME}[_-]" | while read container; do
                # Check if this container is infrastructure (by checking if name contains infra service names)
                IS_INFRA=0
                for infra_svc in $INFRA_SERVICES; do
                    if echo "$container" | grep -q "$infra_svc"; then
                        IS_INFRA=1
                        break
                    fi
                done

                if [ $IS_INFRA -eq 0 ]; then
                    echo "    Stopping $container"
                    docker stop $container 2>/dev/null || true
                    docker rm $container 2>/dev/null || true
                else
                    echo "    Keeping $container (infrastructure)"
                fi
            done
        fi

        # Start infrastructure services first if not running
        if [ -n "$INFRA_SERVICES" ]; then
            echo "  Ensuring infrastructure services are running..."
            for infra_svc in $INFRA_SERVICES; do
                docker compose -p $PROJECT_NAME up -d --no-build $infra_svc 2>/dev/null || true
            done
        fi

        # Build and deploy application services only
        if [ -n "$APP_SERVICES" ]; then
            echo "  Building and deploying application services..."
            docker compose -p $PROJECT_NAME up -d --build $APP_SERVICES
        else
            # Fallback: deploy everything if we couldn't categorize
            docker compose -p $PROJECT_NAME up -d --build
        fi

        CONTAINER_NAME="${PROJECT_NAME}-${SERVICE_NAME}-1"
    else
        # Dockerfile
        docker build -t $PROJECT_NAME .

        ENV_OPTS=""
        if [ -f ".env" ]; then
            ENV_OPTS="--env-file .env"
        fi

        docker run -d --name $PROJECT_NAME $ENV_OPTS -p $HOST_PORT:$INTERNAL_PORT $PROJECT_NAME
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
    # IMPORTANT: We use 'docker compose down' WITHOUT -v flag to preserve volumes
    # This ensures database data and other persistent storage is NOT deleted
    if [ "$BUILD_METHOD" = "compose" ]; then
        # Find old compose projects for this app
        docker ps -a --format "{{.Names}}" | grep "^mushak-$APP_NAME-" | grep -v "$SHA" | while read container; do
            echo "  Stopping $container"
            # Extract project name from container name
            PROJECT=$(echo $container | sed 's/-[^-]*$//')
            # Note: No -v flag = volumes are preserved
            docker compose -p $PROJECT down 2>/dev/null || true
        done
        echo "  (Volumes preserved)"
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
