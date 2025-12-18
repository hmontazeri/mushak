# Nitty Gritty Details

This section dives into the internals of Mushak. You don't need to know this to use the tool, but it helps if you want to understand the magic.

## The Deployment Lifecycle

When you run `mushak deploy`, the following sequence occurs:

1.  **Git Push**: Your code is pushed over SSH to a bare Git repository on the server at `/var/repo/<app>.git`.
2.  **Post-Receive Hook**: The git hook triggers the Mushak deployment script.
3.  **Checkout**: Mushak checks out your code into a commit-based directory in `/var/www/<app>/<commit_sha>`.
4.  **Environment Variables**:
    *   Mushak copies `.env.prod` (or `.env` as fallback) from `/var/www/<app>/.env.prod` to the deployment directory.
    *   All services in docker-compose can access these variables.
5.  **Build**:
    *   For `docker-compose.yml`: Detects the web service (services with "web" in the name), creates a `docker-compose.override.yml` file with:
        *   Port mapping for the web service (random port 8000-9000)
        *   Container name overrides for ALL services (enables zero-downtime deployments)
        *   Application services get versioned names: `mushak-<app>-<sha>-<service>`
        *   Infrastructure services get static names: `<app>_<service>`
    *   Infrastructure services (databases, caches) are identified automatically by image name or via `persistent_services` in mushak.yaml.
    *   For `Dockerfile`: Runs `docker build`.
6.  **Run**:
    *   It finds a free port between 8000 and 9000.
    *   **Infrastructure services** (postgres, redis, etc.) are ensured to be running but NOT restarted.
    *   **Application services** (web, workers) are rebuilt and redeployed.
    *   Only the web service gets the dynamic port mapping for external access.
    *   This smart restart prevents database restarts and maintains connections.
7.  **Health Check**:
    *   Mushak polls `http://localhost:<random_port>/<health_path>` repeatedly.
    *   It waits up to `health_timeout` seconds (default: 30s).
8.  **Switch Traffic**:
    *   Once healthy, Mushak updates the Caddy configuration file.
    *   Caddy reloads and instantly points the domain to the new port. This atomic switch ensures zero downtime.
9.  **Cleanup & Image Management**:
    *   Mushak stops the old container(s) and removes old deployment directories (keeps last 3).
    *   **Image Tagging**: Each deployment's image is tagged as `mushak-<app>:<sha>` for rollback support.
    *   **Image Cleanup**: Old images are automatically pruned, keeping the last 3 versions for rollback.
    *   **Dangling Images**: Build cache and dangling images are pruned after each deployment.
    *   **IMPORTANT**: Docker volumes are NEVER removed. Mushak uses `docker compose down` without the `-v` flag, preserving all database data, uploads, and persistent storage.

## Rollback

Mushak supports instant rollbacks to previous deployments using cached Docker images:

1.  **No Rebuild Required**: Rollbacks use pre-built images, making them nearly instant.
2.  **Health Check**: The rollback target is health-checked before switching traffic.
3.  **Zero Downtime**: Traffic is switched atomically via Caddy configuration.
4.  **Retention Policy**: Last 3 images are kept for rollback support.

Use `mushak rollback` to list available versions or `mushak rollback <sha>` to rollback to a specific version.

## Directory Structure

On your server, Mushak organizes files as follows:

```
/
├── var/
│   ├── repo/
│   │   └── myapp.git/       # Bare git repository
│   └── www/
│       └── myapp/
│           ├── .env.prod    # Environment variables (managed by mushak env)
│           ├── .deployments # Deployment history (for rollback)
│           ├── current/     # Symlink to current deployment
│           ├── abc123d/     # Deployment by commit SHA
│           │   ├── .env.prod  # Copied from parent directory
│           │   └── ...      # Your application code
│           └── def456e/     # Previous deployment (kept for rollback)
└── etc/
    └── caddy/
        ├── Caddyfile        # Master Caddyfile
        └── apps/
            └── myapp.caddy  # App-specific config
```

## Docker Images

Mushak manages Docker images with a naming convention:

- `mushak-<app>:<sha>` - Tagged images for each deployment (kept for rollback)
- `mushak-<app>:latest` - Always points to the current deployment
- `mushak-<app>-<sha>-<service>` - Container names for compose services

## Security Models

- **SSH**: All transport happens over standard SSH. Mushak relies on your existing SSH key configuration.
- **Isolation**: Each app runs in its own Docker container/network. They cannot access each other unless explicitly linked (or via public internet).
- **HTTPS**: Caddy automatically manages Let's Encrypt certificates. You get an A+ SSL rating out of the box.
