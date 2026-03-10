# LocalVSP

LocalVSP is a lightweight self-hosted platform for deploying Docker Compose apps and static websites from a local network share, with a Go-based management UI and optional Cloudflare Tunnel exposure.

## What is in this repo

- `management-ui/`: Go management service and HTMX-style templates.
- `docker-compose.yml`: Core platform stack for Traefik, Gitea, Cloudflared, and the management UI.
- `install.sh`: Linux installer/update script for `/opt/localvsp`.
- `_testapp/`: Small sample apps used to validate runtime detection and deployment flows.

## Quick start

### Local install on a Linux host

```bash
git clone https://github.com/Imkovb/LocalVSP.git
cd LocalVSP
sudo bash install.sh
```

### Update an existing install

```bash
curl -sSL https://raw.githubusercontent.com/Imkovb/LocalVSP/main/install.sh | sudo bash
```

The installer now handles both normal upgrades and older pre-git installs in `/opt/localvsp`. Legacy installs are backed up, migrated to a git-managed checkout, and updated while preserving `data/` and `.env`.

### Publish your own fork

```bash
git init -b main
git add .
git commit -m "Initial import"
git remote add origin https://github.com/Imkovb/LocalVSP.git
git push -u origin main
```

## Repository safety notes

- `.env` and runtime data are ignored by git.
- Deployment helper scripts now expect host credentials at runtime instead of storing them in the repository.
- Replace values in `.env.example` before using them in a real environment.

## Runtime requirements

- Linux target with Docker Engine and Docker Compose v2.
- Access to `/home/vsp` for managed app and static-site folders.
- Optional Cloudflare Tunnel token if you want public app/site hostnames.

## Notes

- Default bootstrap credentials inside the product are still intended for first-run local setups. Rotate them after installation if the environment is shared.
- Management UI and Gitea are intended to remain local-only; only apps and static sites with assigned subdomains should be exposed publicly.