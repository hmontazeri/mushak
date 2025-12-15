# Configuration

Mushak follows a "Convention over Configuration" philosophy, but sometimes you need to tweak the settings.

## `mushak.yaml`

You can place a `mushak.yaml` file in the root of your project to override defaults.

```yaml
# The port your application container listens on.
# Default: Automatically detected from EXPOSE in Dockerfile, or defaults to 80/3000/8080 checks.
internal_port: 3000

# The path Mushak probes to check if the app is ready.
# Default: /
health_path: /api/health

# How long to wait for the app to become healthy before failing the deploy.
# Default: 30 (seconds)
health_timeout: 60

# Override which service to expose (for docker-compose with multiple services)
# Default: Automatically detects services with "web" in the name, or uses first service
service_name: api

# Services that should persist across deployments (not restarted)
# Useful for databases, caches, and other stateful services
# Default: Automatically detects common infrastructure (postgres, mysql, redis, mongodb, etc.)
persistent_services:
  - postgres
  - redis
  - custom-database
```

### Persistent Services

By default, Mushak automatically detects and preserves common infrastructure services during redeployments:
- **Databases**: postgres, mysql, mariadb, mongodb, timescale
- **Caches**: redis, memcached
- **Message Queues**: rabbitmq
- **Search**: elasticsearch

These services are:
1. Started once on first deployment
2. **Not restarted** on subsequent deployments (only application services restart)
3. Kept running to avoid downtime and data loss

**Volume Persistence:**

Docker volumes are **ALWAYS preserved** across all deployments, regardless of whether services are marked as persistent. Mushak never removes volumes when cleaning up old containers. This means:
- Database data persists across deployments
- Uploaded files remain intact
- Any data in named volumes or bind mounts is safe
- You can safely redeploy dozens of times without data loss

**Manual Override:**

If you have custom infrastructure or want specific services to persist, add them to `persistent_services`:

```yaml
persistent_services:
  - postgres      # Will persist even if not auto-detected
  - custom-cache  # Your custom service
  - background-processor  # Any service you want to keep running
```

## Environment Variables

You can manage environment variables using the `mushak env set` command.
Variables are stored securely on the server and injected into your application at runtime.

**Environment file priority:**
- Mushak first looks for `.env.prod` on the server (`/var/www/{app}/.env.prod`)
- Falls back to `.env` if `.env.prod` doesn't exist
- Creates `.env.prod` by default if neither exists
- During each deployment, the environment file is copied to the release directory

For `Dockerfile` projects, variables are passed via `--env-file`.
For `Docker Compose` projects, the environment file is placed in the deployment directory, so you can reference variables in your `docker-compose.yml` like `${MY_VAR}` or use `env_file: .env.prod`.

## Docker Configuration

### Dockerfile Projects
If you have a `Dockerfile`, Mushak builds it as a standard image.
- Ensure you `EXPOSE` the port your app listens on.
- Use `CMD` or `ENTRYPOINT` to start your process.

### Docker Compose Projects
If you have a `docker-compose.yml`, Mushak treats it as a service stack.
- **Do not** map ports to the host (e.g., `- "80:80"`). Mushak manages port mapping dynamically to avoid conflicts.
- Mushak will automatically detect the web service by looking for services with "web" in the name (e.g., `web`, `webapp`, `web-server`). If no service with "web" is found, it uses the first service defined. You can override this with a `mushak.yaml` file by specifying `service_name: your-service`.
