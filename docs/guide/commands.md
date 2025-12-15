# CLI Reference

## mushak init

Initializes a new application on the remote server.

```bash
mushak init [flags]
```

**Flags:**
- `--host`: The SSH user and host (e.g., `root@1.2.3.4`). Required.
- `--domain`: The domain name for this app (e.g., `example.com`). Required.
- `--app`: The internal name for the app. Defaults to current directory name.
- `--branch`: The git branch to track. Defaults to checked-out branch.
- `--key`: Path to private SSH key. Defaults to `~/.ssh/id_rsa`.

## mushak deploy

Deploys the current project state to the server.

```bash
mushak deploy [flags]
```

**Flags:**
- `--force`, `-f`: Force push to the git remote. Useful if history diverged.
- `--branch`: Deploy a specific local branch instead of the current one.

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
- `--key`: Path to SSH key (default `~/.ssh/id_rsa`).

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
