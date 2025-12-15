---
layout: home

hero:
  name: "Mushak"
  text: "Deploy like a rocket"
  tagline: Zero-config, zero-downtime deployment for your own VPS or Home lab
  image:
    src: /logo-mushak.svg
    alt: Mushak Logo
  actions:
    - theme: brand
      text: What is Mushak?
      link: /guide/what-is-mushak
    - theme: alt
      text: Get Started
      link: /guide/getting-started

features:
  - title: Zero Configuration
    details: Works out-of-the-box with Dockerfile or docker-compose.yml. No complex YAML pipelines needed.
  - title: Zero Downtime
    details: Uses Caddy reverse proxy to atomically switch traffic to new containers only after they pass health checks.
  - title: Own Your Data
    details: Deploy to your own VPS or Home lab. No vendor lock-in, no per-seat pricing, just you and your server.
  - title: Smart Builds
    details: Automatically detects your project type and builds strictly isolated containers for every deployment.
---
