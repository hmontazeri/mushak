# <img src="docs/public/logo-mushak.svg" width="150" align="center" /> Mushak

**Zero-Config, Zero-Downtime deployments to your Linux server**

Mushak is a CLI tool that brings PaaS-like deployment experience to your own Linux VPS. Deploy your Docker-based applications with a simple `git push`, complete with automatic builds, health checks, and zero-downtime switching.

## Why Mushak?

I absolutly love Docker and Docker Compose. But it has 2 big issues:

1. You need script or some kind of automation to get docker compose up and running on a remote server and then configure env. etc. 
2. You will face downtime with when redeployen a new version of your app. 

These 2 isses is why Mushak exists. It is a zero-config, zero-downtime alternative to self hosted PaaS you find on the market. 

So its basically GIT + Caddy + Docker (Compose) = Mushak.

## Features

- **Zero Configuration**: Works out-of-the-box with `Dockerfile` or `docker-compose.yml`
- **Zero Downtime**: Uses Caddy reverse proxy for atomic traffic switching
- **Multi-App Support**: Deploy multiple apps on the same server
- **Smart Builds**: Automatically detects and handles both Dockerfile and Docker Compose projects
- **Health Checks**: Ensures new deployments are healthy before switching traffic
- **Automatic Cleanup**: Manages old containers and deployment artifacts
- **Branch Control**: Configure which Git branch triggers deployments
- **Self-Updating**: Built-in update mechanism

## Prerequisites

- **Local Machine**: Git installed
- **Server**: Fresh Linux server (Ubuntu strongly recommended) with SSH access
- **Domain**: DNS pointing to your server (for HTTPS via Caddy)
- **Project**: Application with `Dockerfile` or `docker-compose.yml`

## Installation

### macOS (Homebrew)

```bash
# Coming soon
brew install mushak
```

### From Source

```bash
git clone https://github.com/hmontazeri/mushak.git
cd mushak
go build -o mushak ./cmd/mushak
sudo mv mushak /usr/local/bin/
```

### Binary Download

Download the latest binary from [GitHub Releases](https://github.com/hmontazeri/mushak/releases)

## Quick Start

### 1. Initialize

In your project directory:

```bash
mushak init \
  --host user@your-server.com \
  --domain app.example.com \
  --app myapp \
  --branch main
```

This will:
- Install Docker, Git, and Caddy on your server
- Set up a Git repository
- Configure deployment hooks
- Add a Git remote named `mushak`

### 2. Deploy

```bash
mushak deploy
```

Your app will be built, health-checked, and deployed with zero downtime!

### 3. Access Your App

Visit `https://app.example.com` (make sure DNS is configured)

## How It Works

1. **Git Push**: `mushak deploy` pushes your code to the server's bare Git repository
2. **Hook Triggered**: A post-receive hook on the server starts the deployment process
3. **Build**: Detects Dockerfile or docker-compose.yml and builds your app
4. **Port Assignment**: Finds a free port (8000-9000) for the new container
5. **Health Check**: Polls the health endpoint (default: `/`) for up to 30 seconds
6. **Traffic Switch**: Updates Caddy configuration to point to the new container
7. **Cleanup**: Stops and removes old containers

## Configuration

### Optional `mushak.yaml` in Your Project

```yaml
# Port your app listens on inside the container
internal_port: 3000

# Health check endpoint
health_path: /api/health

# Health check timeout in seconds
health_timeout: 60
```

## Commands

### `mushak init`

Initialize a new app on the server.

```bash
mushak init --host user@server.com --domain app.com --app myapp --branch main
```

**Flags:**
- `--host`: Server address (user@hostname)
- `--domain`: Domain name for the app (required)
- `--app`: App name (default: current directory name)
- `--branch`: Git branch to deploy (default: main)
- `--key`: SSH key path (default: ~/.ssh/id_rsa)
- `--port`: SSH port (default: 22)

### `mushak deploy`

Deploy the current branch to the server.

```bash
mushak deploy
```

**Flags:**
- `-f, --force`: Force push to server

### `mushak destroy`

Remove an app from the server.

```bash
mushak destroy
```

**Flags:**
- `--host`: Server address (optional if config exists)
- `--user`: SSH username (optional if config exists)
- `--app`: App name (optional if config exists)
- `--force`: Skip confirmation prompt

### `mushak update`

Update mushak to the latest version.

```bash
mushak update
```

**Flags:**
- `--check`: Check for updates without installing

### `mushak version`

Show the current version.

```bash
mushak version
```

### `mushak logs`

Stream logs from your application.

```bash
mushak logs
```

**Flags:**
- `-f, --follow`: Follow log output
- `-n, --lines`: Number of lines to show (default: 100)

### `mushak exec`

Open an interactive shell in your application container.

```bash
mushak exec
```

### Environment Variable Management

Mushak provides several commands to manage environment variables:

**Priority system:**
1. First checks for `.env.prod` on the server
2. Falls back to `.env` if `.env.prod` doesn't exist
3. Creates `.env.prod` by default if neither exists
4. During deployment, copies the environment file to each release

**Commands:**

```bash
# Set individual variables and redeploy
mushak env set DATABASE_PASSWORD=secret RAILS_MASTER_KEY=abc123

# Upload entire .env file
mushak env push              # Auto-detects .env.prod, .env.production, or .env
mushak env push .env.prod    # Upload specific file

# Download from server
mushak env pull              # Downloads to local .env.prod

# Compare local vs server
mushak env diff              # Shows differences
```


## Examples

### Node.js App with Dockerfile

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

```yaml
# mushak.yaml
internal_port: 3000
health_path: /health
```

### Multi-Service App with Docker Compose

This example shows a complete setup with web server, background worker, and database using environment variables.

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      POSTGRES_USER: ${DATABASE_USERNAME:-postgres}
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
      POSTGRES_DB: myapp_production
    # Don't expose ports externally - services communicate internally
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  web:
    build: .
    depends_on:
      postgres:
        condition: service_healthy
    env_file:
      - .env.prod
    environment:
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      RAILS_ENV: production
    # Don't specify ports - Mushak handles this dynamically
    volumes:
      - app_storage:/app/storage
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  worker:
    build: .
    command: bundle exec sidekiq
    depends_on:
      postgres:
        condition: service_healthy
    env_file:
      - .env.prod
    environment:
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      RAILS_ENV: production
    volumes:
      - app_storage:/app/storage

volumes:
  postgres_data:
  app_storage:
```

```yaml
# mushak.yaml
internal_port: 3000
health_path: /health
health_timeout: 60
```

```bash
# .env.prod (gitignored)
DATABASE_PASSWORD=your-secure-password
DATABASE_USERNAME=postgres
RAILS_MASTER_KEY=your-master-key
SECRET_KEY_BASE=your-secret-key
```

**How Mushak handles this setup:**

1. **Service Detection**: Mushak automatically detects the `web` service for routing (ignores `postgres` and `worker`)
2. **Port Mapping**: Creates a dynamic port mapping (e.g., 8000:3000) only for the `web` service
3. **Environment Files**:
   - During `mushak init`, prompts to upload your local `.env.prod`
   - During deployment, copies `/var/www/myapp/.env.prod` to each release directory
   - All services can access the variables via `env_file: .env.prod`
4. **Service Startup Order**:
   - `postgres` starts first, waits for health check
   - `web` and `worker` start after postgres is healthy
   - Only `web` is exposed via Caddy reverse proxy
5. **Internal Communication**: Services communicate via internal Docker network (no exposed ports needed)
6. **Smart Redeployments**:
   - **Infrastructure services** (`postgres`) are automatically detected and kept running
   - **Application services** (`web`, `worker`) are rebuilt and restarted
   - Database stays up, no connection drops, no data loss

**Deployment flow:**

```bash
# Initialize (auto-uploads .env.prod if exists)
mushak init --host 1.2.3.4 --user root --domain myapp.com --app myapp
# ✓ Found local .env.prod with 4 variables (DATABASE_PASSWORD, RAILS_MASTER_KEY, +2 more)
# → Upload to server? [Y/n]: y
# ✓ Uploaded .env.prod to server

# First Deploy
mushak deploy
# → Environment file: .env.prod ✓
# → Detecting build method...
#   Service name: web (detected web service)
#   Infrastructure services: postgres
#   Application services: web worker
# → Building and starting containers...
#   postgres: started ✓
#   web: building... healthy ✓
#   worker: building... running ✓

# Second Deploy (after code changes)
mushak deploy
# → Environment file: .env.prod ✓
#   Infrastructure services: postgres
#   Application services: web worker
# → Ensuring infrastructure services are running...
#   postgres: already running ✓ (not restarted)
# → Building and deploying application services...
#   web: building... healthy ✓
#   worker: building... running ✓
```

**Managing environment variables:**

```bash
# Update variables and redeploy
mushak env set DATABASE_PASSWORD=new-password

# Upload entire .env.prod file
mushak env push

# Download from server (for team sync)
mushak env pull

# Compare local vs server
mushak env diff
```

**Important notes:**

- **No port conflicts**: Don't specify `ports:` in your docker-compose.yml for the web service. Mushak dynamically assigns ports (8000-9000 range) to avoid conflicts between deployments.
- **Service naming**: Services with "web" in the name are automatically detected. If you use a different name, create a `mushak.yaml` with `service_name: your-service`.
- **Container naming**: Avoid setting `container_name:` in docker-compose.yml. Let Mushak manage naming for proper isolation between deployments. If you must use custom names, `mushak logs` and `mushak shell` will still work.
- **Database persistence**: Volumes persist across deployments. Database data is not lost during redeployments.
- **Zero-downtime**: Old containers keep running until new ones pass health checks, then Mushak switches traffic and cleans up old containers.

## Multi-App Deployment

Deploy multiple apps on the same server:

```bash
# App 1
cd ~/projects/app1
mushak init --host user@server.com --domain app1.com --app app1

# App 2
cd ~/projects/app2
mushak init --host user@server.com --domain app2.com --app app2
```

Each app gets its own:
- Git repository
- Deployment directory
- Caddy configuration
- Port assignment

## Server Structure

```
/var/repo/
  ├── app1.git/          # Bare Git repo
  ├── app2.git/

/var/www/
  ├── app1/
  │   ├── abc123/        # Deployment by commit SHA
  │   └── def456/
  ├── app2/

/etc/caddy/
  ├── Caddyfile          # Main config (imports apps)
  └── apps/
      ├── app1.caddy     # Per-app reverse proxy config
      └── app2.caddy
```

## Troubleshooting

### Deployment Failed

Check the output from `mushak deploy`. Common issues:
- Health check timeout (increase in `mushak.yaml`)
- Port already in use (mushak will find another)
- Build errors (fix Dockerfile/docker-compose.yml)

### SSH Connection Failed

- Verify SSH access: `ssh user@server.com`
- Check SSH key path: `--key ~/.ssh/id_rsa`
- Ensure server allows SSH connections

### Domain Not Working

- Verify DNS points to your server: `dig app.example.com`
- Check Caddy is running: `ssh user@server.com sudo systemctl status caddy`
- View Caddy logs: `ssh user@server.com sudo journalctl -u caddy -f`

## Security Considerations

- Uses SSH key-based authentication
- Server requires passwordless sudo for Docker and Caddy operations
- HTTPS automatically configured by Caddy
- Containers run in isolated Docker networks

## Roadmap

- [ ] GitHub Actions integration
- [ ] Rollback to previous deployment
- [x] Log viewing (`mushak logs`)
- [x] SSH access (`mushak exec`)
- [x] Environment variable management (`mushak env set`)
- [ ] Database migrations support
- [ ] Custom health check commands

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - see LICENSE file for details

## Credits

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Caddy](https://caddyserver.com/) - Reverse proxy
- [Docker](https://www.docker.com/) - Containerization
- [go-github-selfupdate](https://github.com/rhysd/go-github-selfupdate) - Self-update

## Support

- GitHub Issues: [github.com/hmontazeri/mushak/issues](https://github.com/hmontazeri/mushak/issues)
- Documentation: [github.com/hmontazeri/mushak/wiki](https://github.com/hmontazeri/mushak/wiki)
