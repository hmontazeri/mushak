# Getting Started

Ready to launch your rocket? 🚀 Follow this guide to set up your first deployment.

## Prerequisites

Before installing Mushak, ensure you have:

1.  **A Server**: A fresh Linux VPS (Ubuntu strongly recommended).
    *   You need `root` or `sudo` access.
    *   SSH key authentication should be enabled.
2.  **A Domain**: A domain name (e.g., `myapp.com`) with an A record pointing to your server's IP.
3.  **Local Machine**: You need `git` installed.

## Installation

### Quick Install (macOS & Linux)

The easiest way to install Mushak is via the installation script, which downloads the latest release for your platform:

```bash
curl -sL https://raw.githubusercontent.com/hmontazeri/mushak/main/install.sh | sh
```

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
mushak init root@your-server-ip
```

Mushak will prompt you for:
- **Domain**: The domain name for your app (e.g., `myapp.com`)
- **App name**: Defaults to your current directory name (press Enter to use default)

You can also provide these values as flags if you prefer:

```bash
mushak init root@your-server-ip --domain myapp.com --app my-app-name
```

This command will:
- SSH into your server
- Install Docker, Git, and Caddy automatically
- Set up the remote git repository
- **Detect and upload `.env.prod` if it exists** (you'll be prompted)

If you have a `.env.prod` file, Mushak will detect it and ask if you want to upload it to the server:

```bash
✓ Found local .env.prod with 5 variables (DATABASE_PASSWORD, API_KEY, +3 more)
→ Upload to server? [Y/n]:
```

### 2. Deploy
Now, simply run:

```bash
mushak deploy
```

Watch as your deployment happens! Mushak will:
- Check for environment files (prompt to upload if missing)
- Push your code
- Build containers on the server
- Run health checks
- Switch traffic with zero downtime

### 3. Verify
Visit `https://myapp.com`. Your site should be live with a secure HTTPS certificate provisioned automatically by Caddy.

## Troubleshooting

### SSH Setup

If you're using a non-default SSH key (not `~/.ssh/id_rsa`), add it to your SSH agent:

```bash
# Add your SSH key
ssh-add ~/.ssh/id_ed25519

# Verify it's added
ssh-add -l
```

Or specify the key explicitly with the `--key` flag when running `mushak init`.
