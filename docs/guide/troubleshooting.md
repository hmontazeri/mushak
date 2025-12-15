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

### "Connection refused" on Git Push
This usually means SSH authentication failed.

**Fix:**
Ensure your SSH key is added to the local agent (`ssh-add`) and the public key is in `~/.ssh/authorized_keys` on the server.

## Runtime Issues

### 502 Bad Gateway
You deploy successfully, but see a 502 error in the browser.

**Cause:** The container crashed *after* the health check passed, or Caddy can't reach it.
**Fix:** Check Caddy logs and Docker logs on the server.

```bash
sudo journalctl -u caddy -f
```

## Debugging Tips

You can effectively debug by "logging in" to your deployed environment.

1.  SSH into your server.
2.  Run `docker ps` to see running containers.
3.  Run `docker exec -it <container_name> /bin/sh` to get a shell inside your running app.
