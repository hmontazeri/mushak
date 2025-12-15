# Getting Started with Mushak

This guide will walk you through deploying your first application with Mushak.

## Prerequisites

Before you begin, make sure you have:

1. **A server**: Fresh Ubuntu 20.04+ server with SSH access
2. **A domain**: Domain name with DNS pointing to your server
3. **A project**: Application with a `Dockerfile` or `docker-compose.yml`
4. **SSH key**: SSH key set up for passwordless login to your server

## Step-by-Step Tutorial

### Step 1: Prepare Your Server

First, ensure you can SSH into your server:

```bash
ssh user@your-server.com
```

Make sure your user has sudo privileges. Mushak will install Docker, Git, and Caddy automatically.

### Step 2: Prepare Your Project

Your project needs either a `Dockerfile` or `docker-compose.yml`. Here's a simple example:

**Example Dockerfile:**
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

**Example mushak.yaml (optional):**
```yaml
internal_port: 3000
health_path: /health
```

### Step 3: Initialize Mushak

In your project directory, run:

```bash
mushak init \
  --host user@your-server.com \
  --domain app.example.com \
  --app myapp \
  --branch main
```

**What this does:**
- Connects to your server via SSH
- Installs Docker, Git, and Caddy (if not already installed)
- Creates a Git repository at `/var/repo/myapp.git`
- Sets up deployment hooks
- Configures Caddy for reverse proxy
- Adds a Git remote called `mushak` to your local repo

This process takes 2-5 minutes depending on whether dependencies need to be installed.

### Step 4: Deploy Your App

Simply run:

```bash
mushak deploy
```

**What happens:**
1. Your code is pushed to the server's Git repository
2. The server builds your Docker image
3. A new container starts on a random port (8000-9000)
4. Health checks run to ensure the app is working
5. Caddy updates to route traffic to the new container
6. Old containers are stopped and removed

### Step 5: Access Your App

Visit your domain: `https://app.example.com`

Caddy automatically provisions an SSL certificate via Let's Encrypt!

## Common Workflows

### Updating Your App

Just make changes and deploy again:

```bash
git add .
git commit -m "Update feature"
mushak deploy
```

Each deployment creates a new container. Old containers are automatically cleaned up after successful deployment.

### Deploying Multiple Apps

You can deploy multiple apps to the same server:

```bash
# App 1
cd ~/projects/api
mushak init --host user@server.com --domain api.example.com --app api

# App 2
cd ~/projects/web
mushak init --host user@server.com --domain web.example.com --app web
```

Each app is completely isolated with its own:
- Git repository
- Deployment directory
- Caddy configuration
- Docker containers

### Removing an App

To completely remove an app from the server:

```bash
mushak destroy
```

This will:
- Stop and remove containers
- Delete the Git repository
- Remove deployment files
- Remove Caddy configuration
- Remove the local Git remote

## Troubleshooting

### "Health check failed"

Your app isn't responding on the health endpoint. Check:

1. Is your app listening on the correct port?
2. Is the `internal_port` in `mushak.yaml` correct?
3. Does your health endpoint exist?
4. Increase `health_timeout` if your app takes time to start

### "Connection refused" when accessing domain

Check DNS and Caddy:

```bash
# Verify DNS
dig app.example.com

# Check Caddy status
ssh user@server.com sudo systemctl status caddy

# View Caddy logs
ssh user@server.com sudo journalctl -u caddy -f
```

### Container immediately stops

SSH into the server and check Docker logs:

```bash
ssh user@server.com
docker ps -a
docker logs <container-name>
```

## Advanced Usage

### Custom Branch Deployment

Deploy from a different branch:

```bash
mushak init --branch production
```

Now only pushes to the `production` branch will trigger deployments.

### Using Docker Compose

For apps with multiple services, use `docker-compose.yml`:

```yaml
version: '3'
services:
  web:
    build: .
    environment:
      - DATABASE_URL=postgresql://...

  db:
    image: postgres:14
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

**Note:** Don't specify `ports:` in your compose file - Mushak manages this automatically.

### Health Check Configuration

Customize health checks in `mushak.yaml`:

```yaml
internal_port: 8080
health_path: /api/v1/health
health_timeout: 45  # Wait up to 45 seconds
```

## Best Practices

1. **Use Health Checks**: Always implement a health endpoint in your app
2. **Test Locally**: Build and test your Docker image locally first
3. **Monitor Logs**: Check application logs after deployment
4. **Git Tags**: Use Git tags for production deployments
5. **Backups**: Backup your data before running `mushak destroy`

## Next Steps

- Set up CI/CD with GitHub Actions
- Configure environment variables
- Add database backups
- Monitor your application
- Scale horizontally with load balancers

## Need Help?

- GitHub Issues: [github.com/hmontazeri/mushak/issues](https://github.com/hmontazeri/mushak/issues)
- Documentation: See README.md for full command reference
