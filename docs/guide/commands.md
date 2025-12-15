# CLI Reference

## mushak init

Initializes a new application on the remote server.

```bash
mushak init [flags]
```

**Flags:**
- `--host`: The Server hostname or IP (e.g., `1.2.3.4`). Required.
- `--user`: The SSH username (e.g., `root`). Required.
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
