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

## mushak logs (Coming Soon)

View real-time logs from your deployed application.

## mushak ssh (Coming Soon)

Quickly SSH into the server or the running application container.
