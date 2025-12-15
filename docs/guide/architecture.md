# Nitty Gritty Details

This section dives into the internals of Mushak. You don't need to know this to use the tool, but it helps if you want to understand the magic.

## The Deployment Lifecycle

When you run `mushak deploy`, the following sequence occurs:

1.  **Git Push**: Your code is pushed over SSH to a bare Git repository on the server at `/var/repo/<app>.git`.
2.  **Post-Receive Hook**: The git hook triggers the Mushak server-side binary.
3.  **Checkout**: Mushak checks out your code into a timestamped directory in `/var/www/<app>/<deployment_id>`.
4.  **Build**:
    *   It checks for `docker-compose.yml`. If found, it runs `docker-compose build`.
    *   Otherwise, it runs `docker build`.
5.  **Run**:
    *   It finds a free port between 8000 and 9000.
    *   It spins up the container(s), mapping the internal port (e.g., 3000) to this random host port.
6.  **Health Check**:
    *   Mushak polls `http://localhost:<random_port>/<health_path>` repeatedly.
    *   It waits up to `health_timeout` seconds.
7.  **Switch Traffic**:
    *   Once healthy, Mushak updates the global Caddy configuration via the Caddy API.
    *   Caddy instantly points the domain to the new port. This atomic switch ensures zero downtime.
8.  **Cleanup**:
    *   Mushak stops the old container(s) and removes the old directory to save space.

## Directory Structure

On your server, Mushak organizes files as follows:

```
/
├── var/
│   ├── repo/
│   │   └── myapp.git/       # Bare git repository
│   └── www/
│       └── myapp/
│           ├── current/     # Symlink to active deployment
│           └── 20231215-103022-a1b2c3d/  # Actual deployment code
└── etc/
    └── caddy/
        ├── Caddyfile        # Master Caddyfile
        └── apps/
            └── myapp.caddy  # App-specific config
```

## Security Models

- **SSH**: All transport happens over standard SSH. Mushak relies on your existing SSH key configuration.
- **Isolation**: Each app runs in its own Docker container/network. They cannot access each other unless explicitly linked (or via public internet).
- **HTTPS**: Caddy automatically manages Let's Encrypt certificates. You get an A+ SSL rating out of the box.
