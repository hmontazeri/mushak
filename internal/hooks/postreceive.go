package hooks

import (
	"fmt"
)

// GeneratePostReceiveHook generates the post-receive hook script
func GeneratePostReceiveHook(appName, domain, branch string, noCache bool) string {
	buildOpts := ""
	if noCache {
		buildOpts = "--no-cache"
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

# Configuration
APP_NAME="%s"
DOMAIN="%s"
DEPLOY_BRANCH="%s"
BUILD_OPTS="%s"
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
    CURRENT_LINK="/var/www/$APP_NAME/current"
    PROJECT_NAME="mushak-$APP_NAME-$SHA"

    # Function to sanitize docker-compose.yml (remove hardcoded ports)
    sanitize_docker_compose() {
        local file=$1
        if [ -f "$file" ]; then
            # Check if ports are defined
            if grep -q "^[[:space:]]*ports:" "$file"; then
                echo "→ WARN: Removing hardcoded 'ports' from $file to prevent conflicts..."
                cp "$file" "${file}.bak"
                
                # awk script to comment out ports section
                awk '
                /^[[:space:]]*ports:/ {
                    # Capture indentation length
                    match($0, /^[[:space:]]*/)
                    current_indent = RLENGTH
                    in_ports = 1
                    print "# " $0 " (commented by Mushak)"
                    next
                }
                in_ports {
                    # Pass through empty lines
                    if ($0 ~ /^[[:space:]]*$/) {
                        print
                        next
                    }
                    
                    # Check indentation
                    match($0, /^[[:space:]]*/)
                    this_indent = RLENGTH
                    
                    if (this_indent > current_indent) {
                        print "# " $0
                        next
                    } else {
                        in_ports = 0
                    }
                }
                { print }
                ' "${file}.bak" > "$file"
            fi
        fi
    }

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

    # Update stable symlink (used for infrastructure services consistency)
    ln -snf "$DEPLOY_DIR" "$CURRENT_LINK"

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

    # Sanitize docker-compose files to remove hardcoded ports
    sanitize_docker_compose "docker-compose.yml"
    sanitize_docker_compose "docker-compose.yaml"

    echo ""
    echo "→ Reading configuration..."

    # Read mushak.yaml if it exists
    CUSTOM_PERSISTENT_SERVICES=""
    CACHE_LIMIT="24h"
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

        if grep -q "cache_limit:" mushak.yaml; then
            CACHE_LIMIT=$(grep "cache_limit:" mushak.yaml | awk '{print $2}')
            echo "  Cache limit: $CACHE_LIMIT"
        fi

        # Read persistent_services array (simple parsing)
        if grep -q "persistent_services:" mushak.yaml; then
            CUSTOM_PERSISTENT_SERVICES=$(grep -A 10 "persistent_services:" mushak.yaml | grep "^  - " | sed 's/^  - //' | tr '\n' ' ')
            echo "  Persistent services: $CUSTOM_PERSISTENT_SERVICES"
        fi
    else
        echo "  Using defaults (internal_port=$INTERNAL_PORT, health_path=$HEALTH_PATH, cache_limit=$CACHE_LIMIT)"
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

        # Create override file with port mapping and container name overrides
        # Override container_name for ALL services to enable zero-downtime deployments
        # Infrastructure services get static names, app services get versioned names

        CONTAINER_NAME="${PROJECT_NAME}-${SERVICE_NAME}-1"
        INFRA_PROJECT_NAME="mushak-${APP_NAME}-infra"
        NETWORK_NAME="mushak-${APP_NAME}-net"

        # Create shared network if it doesn't exist
        docker network create $NETWORK_NAME 2>/dev/null || true

        # Create override file with port mapping, container name overrides, and network configuration
        # Override container_name for ALL services to enable zero-downtime deployments
        # Infrastructure services get static names, app services get versioned names

        cat > docker-compose.override.yml <<EOF
version: '3'
networks:
  default:
    external: true
    name: $NETWORK_NAME
services:
EOF

        # Generate configuration for all application services
        for app_svc in $APP_SERVICES; do
            cat >> docker-compose.override.yml <<EOF
  $app_svc:
    container_name: ${PROJECT_NAME}-${app_svc}
EOF

            # Add port mapping for the main web service
            if [ "$app_svc" = "$SERVICE_NAME" ]; then
                cat >> docker-compose.override.yml <<EOF
    ports:
      - "$HOST_PORT:$INTERNAL_PORT"
EOF
            fi

            # Configure external_links for application services to reach infrastructure services
            # This allows app services to find infra services by their service name (hostname aliasing)
            if [ -n "$INFRA_SERVICES" ]; then
                echo "    external_links:" >> docker-compose.override.yml
                for infra_svc in $INFRA_SERVICES; do
                    # Map static container name to service name
                    # e.g. bareagent_postgres:postgres
                    echo "      - ${APP_NAME}_${infra_svc}:${infra_svc}" >> docker-compose.override.yml
                done
            fi
        done

        # Keep infrastructure services with static names (app-specific, not versioned)
        for infra_svc in $INFRA_SERVICES; do
            cat >> docker-compose.override.yml <<EOF
  $infra_svc:
    container_name: ${APP_NAME}_${infra_svc}
EOF
        done

        echo "  Created docker-compose.override.yml"
        echo "    - Overriding container names for zero-downtime deployments"
        echo "    - Configuring shared network: $NETWORK_NAME"
        if [ -n "$INFRA_SERVICES" ]; then
            echo "    - Configuring external links for infrastructure services"
        fi

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
        echo "  Infrastructure services: ${INFRA_SERVICES:-none}"
        echo "  Application services: ${APP_SERVICES:-all}"

        # Note: Mushak automatically overrides container_name in docker-compose.override.yml
        # This enables zero-downtime deployments even if you have custom names

        # Start infrastructure services first if not running
        # We use a SEPARATE persistent project for infrastructure to avoid conflicts with versioned app deployments
        if [ -n "$INFRA_SERVICES" ]; then
            echo "  Ensuring infrastructure services are running..."

            # We need a separate override file for infra to use the correct network but NOT the versioned container names
            # Actually, we can reuse the generated override file because it has the static names for infra services!
            # But we need to be careful not to start app services here.
            
            # Deploy infra services using the INFRA project name
            # This ensures they persist across deployments and don't get recreated unless changed
            # We use --project-directory with a stable symlink to prevent recreation when the SHA directory changes
            for infra_svc in $INFRA_SERVICES; do
                 docker compose --project-directory "$CURRENT_LINK" -p $INFRA_PROJECT_NAME -f $COMPOSE_FILE -f docker-compose.override.yml up -d --remove-orphans $infra_svc
            done
        fi

        # Build and deploy application services only
        # We use the SHA-versioned project name for zero-downtime updates
        if [ -n "$APP_SERVICES" ]; then
            echo "  Building and deploying application services..."
            # Use --no-deps to prevent Docker from trying to interact with the infra services in THIS project scope
            # (since they are now managed by the infra project)
            # Explicitly build first if build opts are present (e.g. --no-cache)
            # 'up --build' doesn't support --no-cache so we handle it separately
            if [[ "$BUILD_OPTS" == *"--no-cache"* ]]; then
                docker compose -p $PROJECT_NAME build $BUILD_OPTS $APP_SERVICES
                docker compose -p $PROJECT_NAME up -d --no-deps $APP_SERVICES
            else
                docker compose -p $PROJECT_NAME up -d --build $BUILD_OPTS --no-deps $APP_SERVICES
            fi
        else
            # Fallback: deploy everything if we couldn't categorize
            if [[ "$BUILD_OPTS" == *"--no-cache"* ]]; then
                docker compose -p $PROJECT_NAME build $BUILD_OPTS
                docker compose -p $PROJECT_NAME up -d
            else
                docker compose -p $PROJECT_NAME up -d --build $BUILD_OPTS
            fi
        fi
    else
        # Dockerfile
        docker build $BUILD_OPTS -t $PROJECT_NAME .

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
    ls -t -d */ 2>/dev/null | grep -v "current" | tail -n +4 | xargs -r rm -rf

    echo ""
    echo "→ Tagging images for rollback..."

    # Image tagging and cleanup for rollback support
    # We tag images with the app name and SHA for easy identification
    IMAGE_REPO="mushak-${APP_NAME}"
    DEPLOYMENTS_FILE="/var/www/$APP_NAME/.deployments"
    KEEP_IMAGES=3

    if [ "$BUILD_METHOD" = "compose" ]; then
        # For compose, tag the built images for the main web service
        # Get the image ID of the service we just deployed
        BUILT_IMAGE=$(docker compose -p $PROJECT_NAME images -q $SERVICE_NAME 2>/dev/null | head -1)
        if [ -n "$BUILT_IMAGE" ]; then
            docker tag "$BUILT_IMAGE" "${IMAGE_REPO}:${SHA}" 2>/dev/null || true
            docker tag "$BUILT_IMAGE" "${IMAGE_REPO}:latest" 2>/dev/null || true
            echo "  Tagged image: ${IMAGE_REPO}:${SHA}"
        fi
    else
        # For Dockerfile, tag the built image
        docker tag "$PROJECT_NAME" "${IMAGE_REPO}:${SHA}" 2>/dev/null || true
        docker tag "$PROJECT_NAME" "${IMAGE_REPO}:latest" 2>/dev/null || true
        echo "  Tagged image: ${IMAGE_REPO}:${SHA}"
    fi

    # Record deployment to manifest file (for rollback listing)
    # Format: SHA TIMESTAMP PORT BUILD_METHOD
    echo "${SHA} $(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ) ${HOST_PORT} ${BUILD_METHOD}" >> "$DEPLOYMENTS_FILE"
    echo "  Recorded deployment to manifest"

    echo ""
    echo "→ Cleaning up old images (keeping last $KEEP_IMAGES)..."

    # Get all tagged versions for this app, sorted by creation time (newest first)
    # Skip 'latest' and keep the most recent N versions
    OLD_TAGS=$(docker images "${IMAGE_REPO}" --format "{{.Tag}} {{.CreatedAt}}" 2>/dev/null | \
        grep -v "latest" | \
        sort -k2 -r | \
        tail -n +$((KEEP_IMAGES + 1)) | \
        awk '{print $1}')

    for tag in $OLD_TAGS; do
        echo "  Removing old image: ${IMAGE_REPO}:${tag}"
        docker rmi "${IMAGE_REPO}:${tag}" 2>/dev/null || true
    done

    # Also cleanup old project-specific images that are no longer tagged
    # These are images like mushak-myapp-abc123f that we can safely remove
    docker images --format "{{.Repository}}:{{.Tag}}" | grep "^mushak-${APP_NAME}-" | grep -v "$SHA" | while read old_image; do
        echo "  Removing old build image: $old_image"
        docker rmi "$old_image" 2>/dev/null || true
    done

    # Prune dangling images and build cache to free up space
    docker image prune -f > /dev/null 2>&1 || true
    
    # Use CACHE_LIMIT for builder prune. 
    # If it looks like a duration (ends with h, m, s), use until filter.
    # Otherwise, it's just a raw filter or we fallback to 24h.
    if [[ "$CACHE_LIMIT" =~ [0-9]+[hms]$ ]]; then
        docker builder prune -f --filter "until=$CACHE_LIMIT" > /dev/null 2>&1 || true
    else
        docker builder prune -f --filter "$CACHE_LIMIT" > /dev/null 2>&1 || true
    fi
    echo "  Pruned dangling images and build cache (limit: $CACHE_LIMIT)"

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
`, appName, domain, branch, buildOpts)
}
