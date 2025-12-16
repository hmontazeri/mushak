# Deployment Guide

Mushak is designed to make deployment as boring as possible.

## Standard Deployment Flow

The standard workflow is:

1.  Make code changes locally.
2.  Commit your changes to git.
    ```bash
    git add .
    git commit -m "New feature"
    ```
3.  Deploy.
    ```bash
    mushak deploy
    ```

> [!TIP]
> You don't need to push to GitHub/GitLab first. `mushak deploy` pushes directly to your private server.

## Deploying Multiple Apps

Mushak shines at hosting multiple isolated apps on a single affordable VPS. Every app is sandboxed.

To deploy a second app:

1.  Navigate to the second project folder.
2.  Run init.
    ```bash
    mushak init user@your-server-ip --domain api.myapp.com
    ```
3.  Deploy.
    ```bash
    mushak deploy
    ```

You now have two apps running on the same server:
- `myapp.com` -> `my-app-name` container
- `api.myapp.com` -> `my-api` container

Caddy automatically routes traffic based on the hostname.

## Branch Deployments

By default, Mushak deploys the current branch. You can enforce a specific branch (like `production`) during initialization:

```bash
mushak init user@your-server --domain myapp.com --branch production
```

Or you can use the `--branch` flag during deploy to deploy a specific local branch to the server:

```bash
mushak deploy --branch my-feature-branch
```

## Environment Variables

Mushak manages environment variables separately from your code (they're not committed to git).

### Uploading Environment Files

If you have a `.env.prod` file locally:

```bash
# Auto-upload during init (you'll be prompted)
mushak init ...

# Auto-upload during deploy (if missing on server)
mushak deploy

# Manually upload
mushak env push              # Auto-detects .env.prod, .env.production, or .env
mushak env push .env.prod    # Upload specific file
```

### Setting Individual Variables

```bash
# Set variables and trigger redeployment
mushak env set DATABASE_PASSWORD=secret API_KEY=abc123
```

### Syncing with Team

```bash
# Download environment file from server
mushak env pull

# Compare local vs server
mushak env diff
```

**Environment file priority:**
1. `.env.prod` on server
2. `.env` on server (fallback)
3. Creates `.env.prod` by default if neither exists

During deployment, Mushak copies the environment file to each release directory, making it available to all services in docker-compose.

## Data Persistence

### Volumes Are Always Preserved

**IMPORTANT**: Mushak **NEVER** removes Docker volumes. Your data is safe across all deployments.

When cleaning up old containers, Mushak uses `docker compose down` **without** the `-v` or `--volumes` flag. This means:

✅ **Database data persists** - Your postgres, mysql, mongodb data is never deleted
✅ **Uploaded files remain** - User uploads and storage volumes are safe
✅ **Named volumes survive** - Any volumes defined in docker-compose.yml persist
✅ **Bind mounts are safe** - Host-mounted directories remain intact

**Example:**
```yaml
volumes:
  postgres_data:      # Persists across deployments ✓
  user_uploads:       # Persists across deployments ✓
  app_cache:          # Persists across deployments ✓
```

You can redeploy 100 times - your data stays safe.

### Persistent Services

Infrastructure services (databases, caches) are also **not restarted** during redeployments, providing:
- Zero downtime for databases
- No connection drops
- Continuous data availability

See [Configuration - Persistent Services](/guide/configuration#persistent-services) for details.
