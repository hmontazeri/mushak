# Troubleshooting

Deployment didn't go as planned? Here are common issues and fixes.

## Deployment Failures

### "Health check failed"
This is the most common error. It means your app container started, but Mushak couldn't verify it was ready.

**Fixes:**
1.  **Wrong Port**: Does your app listen on port 3000 but Mushak is checking 8080? Set `internal_port` in `mushak.yaml`.
2.  **Slow Startup**: Does your app take >30s to boot? Increase `health_timeout` in `mushak.yaml`.
3.  **Crash on Boot**: SSH into the server and check docker logs:
    ```bash
    ssh user@server
    docker ps -a
    docker logs <container_id>
    ```

### "Connection refused" or "ssh: unable to authenticate"
This usually means SSH authentication failed.

**Common causes:**
1.  **No SSH key in agent**: Mushak tries SSH agent first. Add your key: `ssh-add ~/.ssh/id_ed25519` (or `id_rsa`)
2.  **Wrong key path**: Specify explicitly: `mushak init --key ~/.ssh/id_ed25519`
3.  **Key not on server**: Ensure your public key is in `~/.ssh/authorized_keys` on the server
4.  **Test SSH directly**: `ssh user@server` - if this works, Mushak should work too

**Priority of SSH keys:**
1. SSH agent (if available)
2. Key specified via `--key` flag
3. Default: `~/.ssh/id_rsa`

### "env file not found" or missing environment variables
Your docker-compose.yml references `.env.prod` but deployment fails.

**Fix:**
1.  **Upload environment file**: `mushak env push` (auto-detects `.env.prod` or `.env`)
2.  **Set individual variables**: `mushak env set DATABASE_PASSWORD=secret`
3.  **During init**: Mushak will prompt to upload if it detects a local env file
4.  **During deploy**: Mushak will prompt to upload if server has no env file

**Priority of environment files:**
1. `.env.prod` on server (`/var/www/<app>/.env.prod`)
2. `.env` on server (`/var/www/<app>/.env`)
3. Not found - deployment may fail if docker-compose requires it

### Wrong service detected for docker-compose
Mushak is trying to expose `postgres` instead of your web service.

**Fix:**
1.  **Rename service**: Include "web" in the name (e.g., `web`, `webapp`, `web-server`)
2.  **Use mushak.yaml**: Create a `mushak.yaml` file with:
    ```yaml
    service_name: your-service-name
    internal_port: 3000
    ```

**Service detection priority:**
1. Services with "web" in the name
2. First service defined in docker-compose.yml
3. Override with `service_name` in mushak.yaml

### Database restarting on every deploy

By default, Mushak automatically detects infrastructure services (databases, caches) and keeps them running across deployments. Only application services are restarted.

**Automatically detected infrastructure:**
- postgres, mysql, mariadb, mongodb, timescale
- redis, memcached
- rabbitmq, elasticsearch

**If your database still restarts:**
1.  Check if it uses a recognized infrastructure image (e.g., `image: postgres:16`)
2.  If using a custom image, add it to `mushak.yaml`:
    ```yaml
    persistent_services:
      - my-database
      - custom-cache
    ```

## Runtime Issues

### 502 Bad Gateway
You deploy successfully, but see a 502 error in the browser.

**Cause:** The container crashed *after* the health check passed, or Caddy can't reach it.
**Fix:** Check Caddy logs and Docker logs on the server.

```bash
sudo journalctl -u caddy -f
```

## Debugging Tips

Mushak provides several commands to debug your deployment:

### View logs
```bash
mushak logs              # View container logs
mushak logs -f           # Follow logs in real-time
mushak logs -n 500       # Show last 500 lines
```

### Access container shell
```bash
mushak shell             # Opens interactive shell in your app container
```

### Manual debugging on server
1.  SSH into your server: `ssh user@server`
2.  Run `docker ps` to see running containers
3.  Run `docker exec -it <container_name> /bin/sh` to get a shell inside your running app
4.  Check environment: `cat /var/www/<app>/.env.prod`
