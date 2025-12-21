# What is Mushak?

**Mushak** (Persian for "Rocket" 🚀) is a CLI tool designed to bring a PaaS-like deployment experience to your own Linux VPS or Home lab. It allows you to deploy Docker-based applications with a simple `git push`, handling build triggers, health checks, and zero-downtime traffic switching automatically.

## Why Mushak?

In the modern devops landscape, developers often face a binary choice:
1. **Expensive PaaS**: Heroku, Vercel, Railway. Easy to use, but costs scale rapidly and you don't own the infrastructure.
2. **Complex DIY**: Kubernetes, manual Docker management, Nginx config hacking. Cheap infrastructure, but high maintenance and complexity.
3. **Different use cases**: Kamal, Nomad, Nomad. Cheap infrastructure, but needs config / env management and some orchestration.

**Mushak** sits in the middle. It gives you the "Heroku experience" (`git push` to deploy) on your own $5 DigitalOcean/Hetzner droplet.

### Targeted Audience

Mushak is **not** for everyone.
- It is **not** for large enterprises needing complex orchestration (use K8s).
- It is **not** for serverless functions (use Vercel/Lambda).
- It is **not** for multi-server orchestration (use kamal or k8s).

It **is** for:
- **Solo Developers** shipping side projects.
- **Small Teams** who want to iterate fast without a dedicated DevOps engineer.
- **Agencies** managing multiple client projects on a few efficient servers.

## Core Philosophy

### 1. Zero Configuration
If your project has a `Dockerfile` or `docker-compose.yml`, it should just work. Mushak infers how to build and run your app. You shouldn't need to write a complex CI/CD pipeline just to get a container running.

### 2. Zero Downtime
Deployments should never break the site. Mushak spins up the new version alongside the old one, waits for it to be healthy, and only then switches the traffic. If the new version fails, the old one stays active.

### 3. Simplicity > Flexibility
Mushak makes opinionated choices (Linux, Caddy, Docker) to keep the tool simple and reliable. We prioritize "just working" over infinite configurability.
