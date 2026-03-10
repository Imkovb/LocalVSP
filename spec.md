# Local VSP (Virtual Server Provider) Specification

## Overview
Local VSP is a lightweight, self-hosted platform designed to run on Linux VMs (like Ubuntu) or ARM devices (like Raspberry Pi). It provides a web-based management interface to deploy local project folders and static websites from network shares backed by Docker Compose and generated runtime files when needed. It includes a local Gitea instance for repository hosting and documentation, and integrates with Cloudflare Tunnels for secure external access without opening firewall ports.

## Architecture

### Core Components
1. **Management UI (Go + HTMX)**
   - A lightweight, single-binary web server.
   - Provides a dashboard to detect project folders, trigger deployments, manage local sites, and view container status.
   - Interacts with the local Docker daemon to execute `docker compose` commands.
   - Extremely low resource footprint, ideal for Raspberry Pi.

2. **Gitea (Git Server & Wiki)**
   - Self-hosted Git service running in a Docker container.
   - Uses SQLite to minimize memory usage.
   - Hosts local repositories and provides a built-in Wiki for setup instructions and documentation.

3. **Traefik (Reverse Proxy)**
   - Automatically routes traffic to deployed Docker containers based on labels.
   - Listens to the Docker socket to dynamically discover new services.

4. **Cloudflared (Cloudflare Tunnel)**
   - Securely connects the local Traefik instance to the Cloudflare edge.
   - Exposes the Management UI, Gitea, and deployed applications to the internet without requiring port forwarding or public IPs.

### Deployment Flow
1. User places a project folder in `/home/vsp/docker/` or a static website folder in `/home/vsp/html/` via the Samba network share.
2. If a Docker project already contains a compose file, the Management UI uses it directly.
3. If a project folder matches a known runtime type (for example Node.js, Python, Go, PHP, Rust, or static HTML), the Management UI can generate a `Dockerfile` and basic compose file automatically.
4. The Management UI applies an optional generated override file for Local VSP specific routing and host-port exposure.
5. The Management UI runs `docker compose build` and `docker compose up -d` in the project directory or runs an nginx container for a static HTML site.
6. Traefik detects routed containers and forwards traffic to them.
7. Cloudflare Tunnel routes external traffic to Traefik when configured.

## Directory Structure
```text
/opt/localvsp/
├── docker-compose.yml      # Core infrastructure stack
├── .env                    # Environment variables (Cloudflare token, etc.)
├── data/
│   ├── gitea/              # Gitea persistent data
│   └── traefik/            # Traefik configuration and certificates
└── management-ui/          # Go backend and HTMX templates
    ├── main.go
    ├── internal/
    └── web/
        └── templates/
```

## Installation

A single self-contained bash script (`install.sh`) handles the complete bootstrap with one command:

```bash
curl -sSL https://raw.githubusercontent.com/Imkovb/LocalVSP/main/install.sh | sudo bash
```

No files need to be copied to the target machine first. The script will:

1. Install `curl` and `git` if missing.
2. Install Docker and Docker Compose V2 if missing.
3. **Clone the repository** to `/opt/localvsp/` (or pull latest if already present).
4. Prompt for a Cloudflare Tunnel Token and base domain (skippable).
5. Write a `chmod 600` `.env` file.
6. Build and start all services with `docker compose up --build -d`.
7. Run HTTP health checks and wait for services to be ready.
8. Print a summary of all service URLs and next steps.

Re-running the installer performs an in-place update without touching data. If the local checkout contains uncommitted changes or cannot be fast-forwarded safely, the installer leaves the source tree untouched rather than forcing a reset.

For older installs that were copied into `/opt/localvsp` without git metadata, the installer stages a fresh clone, backs up the existing application files, replaces the application tree, and preserves both `.env` and `data/`.

## Requirements for User Projects
Projects can be deployed in one of two ways:

1. Provide your own `docker-compose.yml`.
2. Drop a folder containing a known trigger file so Local VSP can generate container files automatically.

Examples of trigger files currently recognized by the management UI:

- `package.json`
- `requirements.txt`
- `go.mod`
- `composer.json`
- `Cargo.toml`
- `index.html`

To be accessible through a named hostname, a project must either include compatible Traefik labels in its own compose setup or use Local VSP's generated override/subdomain flow.

Static sites are deployed from `/home/vsp/html/<site>` and served through an nginx container managed by Local VSP.

## Requirements for User Compose Projects
If you bring your own compose file and want Local VSP to route traffic automatically, the project should expose a web service port and remain compatible with an external Traefik network.

Example:
```yaml
services:
  myapp:
    image: myapp:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.myapp.rule=Host(`myapp.yourdomain.com`)"
```
