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
2.  Run init with a different app name and domain.
    ```bash
    mushak init \
      --host your-server-ip \
      --user user \
      --domain api.myapp.com \
      --app my-api
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
mushak init --branch production ...
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
