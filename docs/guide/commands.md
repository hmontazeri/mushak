# CLI Reference

## mushak init

Initializes a new application on the remote server.

```bash
mushak init USER@HOST
```

**Arguments:**
- `USER@HOST`: SSH connection string (e.g., `root@192.168.1.100`)

Mushak will interactively prompt you for:
- **Domain**: The domain name for your app
- **App name**: Defaults to current directory name

**Examples:**

```bash
# Interactive mode
mushak init root@1.2.3.4

# With flags
mushak init root@1.2.3.4 --domain myapp.com --app my-app
```

**Flags:**
- `--host`: The Server hostname or IP (overrides HOST from argument)
- `--user`: The SSH username (overrides USER from argument)
- `--domain`: The domain name for this app (skips prompt if provided)
- `--app`: The internal name for the app (skips prompt if provided)
- `--branch`: The git branch to track. Defaults to `main`.
- `--key`: Path to private SSH key. Defaults to `~/.ssh/id_rsa`.
- `--port`: SSH port. Defaults to `22`.

## mushak deploy

Deploys the current project state to the server.

```bash
mushak deploy [flags]
```

**Flags:**
- `--force`, `-f`: Force push to the git remote. Useful if history diverged.
- `--no-cache`: Force a rebuild of the application without using usage of Docker cache.
- `--branch`: Deploy a specific local branch instead of the current one.

## mushak env

Manage environment variables for your application.

### mushak env set

Sets environment variables on the remote server and restarts the application.
This command securely updates the environment file on the server and triggers a redeployment to ensure the new variables are applied.

**Environment file priority:**
- Mushak will first look for `.env.prod` on the server
- If `.env.prod` doesn't exist, it will look for `.env`
- If neither exists, it will create `.env.prod` by default
- During deployment, the environment file is copied to each release directory

```bash
mushak env set [KEY=VALUE]...
```

**Example:**

```bash
mushak env set DB_HOST=db.example.com API_KEY=secret123 DATABASE_PASSWORD=mysecret
```

### mushak env push

Uploads your local environment file to the server. Auto-detects `.env.prod`, `.env.production`, or `.env` in that order, or you can specify a file explicitly.

```bash
mushak env push [file] [flags]
```

**Flags:**
- `--deploy`, `-d`: Trigger a redeployment after uploading the file.

**Examples:**

```bash
# Auto-detect and upload
mushak env push

# Upload specific file
mushak env push .env.production

# Upload and immediately redeploy
mushak env push --deploy
```

### mushak env pull

Downloads the environment file from the server to your local `.env.prod` file. Useful for syncing environment variables across team members.

```bash
mushak env pull
```

**Example:**

```bash
mushak env pull
```

### mushak env diff

Compares your local and server environment files, showing which variables exist only locally, only on the server, or have different values.

```bash
mushak env diff
```

**Example:**

```bash
mushak env diff
# Output:
# + NEW_VAR (only in local)
# - OLD_VAR (only on server)
# ≠ API_KEY (values differ)
```

## mushak domain

Update the domain name for the deployed application. This command updates the local configuration, the running Caddy configuration on the server (immediate effect), and the deployment hook for future deployments.

The command will prompt for confirmation that you have updated your DNS records. You can skip this prompt with the `--force` flag.

```bash
mushak domain [NEW_DOMAIN] [flags]
```

**Example:**

```bash
mushak domain new-app.example.com
```

**Flags:**
- `--force`, `-f`: Skip DNS confirmation prompt.

## mushak destroy

Completely removes an application from the server. **Destructive action.**

```bash
mushak destroy [flags]
```

**Flags:**
- `--app`: Name of the app to destroy.
- `--force`: Skip confirmation prompt.

## mushak logs

View real-time logs from your deployed application.

```bash
mushak logs [flags]
```

**Flags:**
- `--tail`, `-n`: Number of lines to show (default "100").
- `--follow`, `-f`: Follow log output (default true).
- `--container`, `-c`: Filter logs by container name (use `mushak containers` to list available names).
- `--key`: Path to SSH key (default `~/.ssh/id_rsa`).

## mushak containers

List all running Docker containers for the application. Useful to discover container names for use with `mushak logs --container`.

```bash
mushak containers
```

**Example:**

```bash
mushak containers
# Output:
# NAMES                              STATUS          PORTS
# mushak-myapp-abc123-web            Up 2 hours      0.0.0.0:8080->80/tcp
# myapp_postgres                     Up 3 days       5432/tcp
```

## mushak redeploy

Trigger a redeployment of the current version on the server. Useful for restarting the application or applying environment changes without pushing new code.

```bash
mushak redeploy
```

## mushak rollback

Rollback to a previous deployment version. Mushak keeps the last 3 Docker images for instant rollbacks without rebuilding.

```bash
mushak rollback [sha]
```

**Arguments:**
- `sha` (optional): The commit SHA to rollback to. If omitted, shows available versions interactively.

**Special values:**
- `-1`: Rollback to the previous version

**Examples:**

```bash
# List available versions interactively
mushak rollback

# Output:
#   SHA        DEPLOYED               STATUS
#   ---        --------               ------
#   abc123d    2024-12-18 10:30:45    current ←
#   def456e    2024-12-17 14:22:10
#   ghi789f    2024-12-16 09:15:33

# Rollback to specific version
mushak rollback def456e

# Rollback to previous version (shorthand)
mushak rollback -1
```

**How it works:**
- Uses pre-built Docker images (no rebuild required)
- Performs health check before switching traffic
- Updates Caddy reverse proxy atomically
- Zero downtime rollback

**Note:** Only versions with cached images can be rolled back to. Mushak automatically keeps the last 3 images for rollback support.

## mushak shell

Opens an interactive bash/shell session directly inside the running application container. This is useful for debugging issues, inspecting files, or checking environment variables in the production environment.

```bash
mushak shell [flags]
```

**Example:**

```bash
# Open a shell to the current app's container
mushak shell

# Verify environment variables inside the container
root@1.2.3.4:/app# env | grep PORT
PORT=8080
```

**Flags:**
- `--key`: Path to SSH key (default `~/.ssh/id_rsa`).

## mushak update

Update mushak to the latest version. Checks for the latest release on GitHub and updates the binary if a newer version is available.

```bash
mushak update [flags]
```

**Flags:**
- `--check`: Check for updates without installing

## mushak version

Print the version number of the installed mushak CLI.

```bash
mushak version
```

## mushak completion

Generate the autocompletion script for the specified shell.

```bash
mushak completion [bash|zsh|fish|powershell]
```

### Bash

```bash
source <(mushak completion bash)
```

### Zsh

```bash
source <(mushak completion zsh)
```

### Fish

```bash
mushak completion fish | source
```
