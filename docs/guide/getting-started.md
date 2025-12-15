# Getting Started

Ready to launch your rocket? 🚀 Follow this guide to set up your first deployment.

## Prerequisites

Before installing Mushak, ensure you have:

1.  **A Server**: A fresh Ubuntu 20.04 (or newer) VPS.
    *   You need `root` or `sudo` access.
    *   SSH key authentication should be enabled.
2.  **A Domain**: A domain name (e.g., `myapp.com`) with an A record pointing to your server's IP.
3.  **Local Machine**: You need `git` installed.

## Installation

### macOS (Homebrew)

> [!NOTE]
> Homebrew support is coming soon.

### Binary Installation

Download the latest binary for your platform from the [GitHub Releases](https://github.com/hmontazeri/mushak/releases) page.

Move it to your path:
```bash
chmod +x mushak
sudo mv mushak /usr/local/bin/
```

### From Source

If you have Go installed:

```bash
git clone https://github.com/hmontazeri/mushak.git
cd mushak
go build -o mushak ./cmd/mushak
sudo mv mushak /usr/local/bin/
```

## First Deployment

### 1. Initialize
Go to your project directory (which must contain a `Dockerfile` or `docker-compose.yml`).

```bash
mushak init \
  --host user@your-server-ip \
  --domain myapp.com \
  --app my-app-name
```

This command will:
- SSH into your server.
- Install Docker, Git, and Caddy automatically.
- set up the remote git repository.

### 2. Deploy
Now, simply run:

```bash
mushak deploy
```

Watch as your "Rocket" lifts off! Mushak will push your code, build it on the server, and deploy it.

### 3. Verify
Visit `https://myapp.com`. Your site should be live with a secure HTTPS certificate provisioned automatically by Caddy.
