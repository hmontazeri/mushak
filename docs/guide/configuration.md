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
```

## Docker Configuration

### Dockerfile Projects
If you have a `Dockerfile`, Mushak builds it as a standard image.
- Ensure you `EXPOSE` the port your app listens on.
- Use `CMD` or `ENTRYPOINT` to start your process.

### Docker Compose Projects
If you have a `docker-compose.yml`, Mushak treats it as a service stack.
- **Do not** map ports to the host (e.g., `- "80:80"`). Mushak manages port mapping dynamically to avoid conflicts.
- Mushak will look for the service named `web`, `app`, or the first service defined to route traffic to.
