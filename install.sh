#!/bin/bash
# Local VSP Installer
# Usage:  sudo bash install.sh
# Or:     curl -sSL https://raw.githubusercontent.com/Imkovb/LocalVSP/main/install.sh | sudo bash
#
# Override repo URL or branch before piping:
#   LOCALVSP_REPO=https://github.com/Imkovb/LocalVSP.git sudo bash install.sh

set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────────────────
REPO_URL="${LOCALVSP_REPO:-https://github.com/Imkovb/LocalVSP.git}"
BRANCH="${LOCALVSP_BRANCH:-main}"
INSTALL_DIR="/opt/localvsp"
DATA_DIR="${INSTALL_DIR}/data"

# ─── Colors & Logging ─────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
log()  { echo -e "${GREEN}[✓]${NC} $*"; }
info() { echo -e "${CYAN}[→]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
fail() { echo -e "${RED}[✗] ERROR:${NC} $*" >&2; exit 1; }
section() { echo -e "\n${BOLD}${CYAN}── $* ──${NC}"; }
random_secret() {
    tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64
}
STAGE_DIR=""
cleanup() {
    if [ -n "${STAGE_DIR:-}" ] && [ -d "${STAGE_DIR}" ]; then
        rm -rf "${STAGE_DIR}"
    fi
}
trap cleanup EXIT

clone_repo_to_stage() {
    cleanup
    STAGE_DIR=$(mktemp -d)
    info "Downloading Local VSP source from ${REPO_URL} (${BRANCH}) …"
    git clone --depth=1 --branch "${BRANCH}" "${REPO_URL}" "${STAGE_DIR}"
}

install_dir_has_files() {
    [ -d "${INSTALL_DIR}" ] && [ -n "$(ls -A "${INSTALL_DIR}" 2>/dev/null)" ]
}

install_stage_into_install_dir() {
    local backup_dir=""

    if install_dir_has_files; then
        backup_dir="${INSTALL_DIR}.backup-$(date +%Y%m%d-%H%M%S)"
        mkdir -p "${backup_dir}"
        info "Backing up current application files to ${backup_dir} …"
        (
            cd "${INSTALL_DIR}"
            tar --exclude='./data' --exclude='./.env' -cf - .
        ) | (
            cd "${backup_dir}"
            tar -xf -
        )

        info "Replacing application files in ${INSTALL_DIR} while preserving data and .env …"
        find "${INSTALL_DIR}" -mindepth 1 -maxdepth 1 ! -name 'data' ! -name '.env' -exec rm -rf {} +
    else
        mkdir -p "${INSTALL_DIR}"
    fi

    (
        cd "${STAGE_DIR}"
        tar --exclude='./data' --exclude='./.env' -cf - .
    ) | (
        cd "${INSTALL_DIR}"
        tar -xf -
    )

    log "Application files installed in ${INSTALL_DIR}"
    if [ -n "${backup_dir}" ]; then
        log "Previous application files backed up to ${backup_dir}"
    fi
}

update_existing_git_checkout() {
    local current_origin=""

    current_origin=$(git -C "${INSTALL_DIR}" remote get-url origin 2>/dev/null || true)
    if [ -n "${current_origin}" ] && [ "${current_origin}" != "${REPO_URL}" ]; then
        info "Updating git remote origin to ${REPO_URL} …"
        git -C "${INSTALL_DIR}" remote set-url origin "${REPO_URL}"
    fi

    info "Existing git checkout found at ${INSTALL_DIR}. Checking for a safe fast-forward update …"
    git -C "${INSTALL_DIR}" fetch --depth=1 origin "${BRANCH}"

    if [ -n "$(git -C "${INSTALL_DIR}" status --porcelain)" ]; then
        warn "Local changes detected in ${INSTALL_DIR}. Skipping automatic source replacement to avoid overwriting work."
        warn "Commit, stash, or remove those changes if you want the installer to update application files."
        return 0
    fi

    if git -C "${INSTALL_DIR}" merge-base --is-ancestor HEAD "origin/${BRANCH}"; then
        git -C "${INSTALL_DIR}" checkout -B "${BRANCH}" "origin/${BRANCH}"
        log "Source updated to latest ${BRANCH}"
        return 0
    fi

    warn "Current checkout cannot be fast-forwarded cleanly. Falling back to a staged replacement update."
    clone_repo_to_stage
    install_stage_into_install_dir
}

# ─── Banner ───────────────────────────────────────────────────────────────────
echo -e "${GREEN}"
cat << 'BANNER'
  _                 _  __     _______ ____
 | |    ___   ___ __ _| \ \   / / ____|  _ \
 | |   / _ \ / __/ _` | |\ \ / /|  _| | |_) |
 | |__| (_) | (_| (_| | | \ V / | |___|  _ <
 |_____\___/ \___\__,_|_|  \_/  |_____|_| \_\
BANNER
echo -e "${NC}  Self-Hosted Virtual Server Provider — Installer\n"

# ─── Pre-flight Checks ────────────────────────────────────────────────────────
section "Pre-flight checks"

[ "$EUID" -ne 0 ] && fail "Please run as root: sudo bash install.sh"

# Detect architecture (for Raspberry Pi support)
ARCH=$(uname -m)
info "Architecture: ${ARCH}"
info "OS: $(. /etc/os-release && echo "${PRETTY_NAME}")"

# Check for required base tools
for cmd in curl git; do
    if ! command -v "$cmd" &>/dev/null; then
        info "Installing missing dependency: ${cmd}"
        apt-get update -qq && apt-get install -y -qq "$cmd"
    fi
done
log "Base dependencies ready"

# ─── Docker ───────────────────────────────────────────────────────────────────
section "Docker"

if ! command -v docker &>/dev/null; then
    info "Docker not found. Installing via get.docker.com …"
    curl -fsSL https://get.docker.com | sh
    # Add default non-root user to docker group if exists
    SUDO_USER_HOME=$(getent passwd "${SUDO_USER:-}" | cut -d: -f6)
    [ -n "${SUDO_USER:-}" ] && usermod -aG docker "${SUDO_USER}" && \
        info "Added ${SUDO_USER} to the docker group (re-login to use docker without sudo)"
    log "Docker installed: $(docker --version)"
else
    log "Docker already installed: $(docker --version)"
fi

# Verify Docker Compose V2 (bundled with modern Docker installs)
if ! docker compose version &>/dev/null; then
    info "Installing docker-compose-plugin …"
    apt-get update -qq
    apt-get install -y -qq docker-compose-plugin
fi
log "Docker Compose ready: $(docker compose version --short)"

# Ensure Docker daemon is running
systemctl enable docker --quiet
systemctl start docker
log "Docker daemon running"

# ─── VSP System User ─────────────────────────────────────────────────────────
section "System user"

if id vsp &>/dev/null; then
    log "User 'vsp' already exists"
else
    useradd -m -s /bin/bash -U vsp
    echo 'vsp:vsp' | chpasswd
    log "Created system user 'vsp' (password: vsp)"
fi

# ─── Samba (Windows File Sharing) ────────────────────────────────────────────
section "Setting up Samba (Windows file sharing)"

apt-get install -y -qq samba samba-common-bin

# Add share block if not already present
if ! grep -q '\[vsp-home\]' /etc/samba/smb.conf; then
    cat >> /etc/samba/smb.conf << 'SAMBA_CONF'

[vsp-home]
   comment = LocalVSP Home
   path = /home/vsp
   browseable = yes
   read only = no
   writable = yes
   valid users = vsp
   force user = vsp
   create mask = 0664
   directory mask = 0775
SAMBA_CONF
    log "Samba share [vsp-home] added to smb.conf"
else
    log "Samba share [vsp-home] already configured"
fi

# Register vsp in Samba's password database (password: vsp)
(echo vsp; echo vsp) | smbpasswd -s -a vsp

systemctl enable smbd --quiet
systemctl restart smbd
log "Samba running — share: \\\\$(hostname -I | awk '{print \$1}')\\vsp-home  (user: vsp / vsp)"

# ─── User Home Folders ────────────────────────────────────────────────────────
section "Creating user home folders"

mkdir -p /home/vsp/docker
mkdir -p /home/vsp/html
mkdir -p /home/vsp/.localvsp
chown -R vsp:vsp /home/vsp/docker /home/vsp/html /home/vsp/.localvsp
log "Created /home/vsp/docker      (place docker-compose projects here)"
log "Created /home/vsp/html        (place flat HTML sites here)"
log "Created /home/vsp/.localvsp   (autostart + subdomain config)"

# ─── Fetch / Update Source ────────────────────────────────────────────────────
section "Fetching Local VSP source"

if [ -d "${INSTALL_DIR}/.git" ]; then
    update_existing_git_checkout
elif install_dir_has_files; then
    warn "Legacy install detected at ${INSTALL_DIR} without git metadata. Migrating it to a managed checkout."
    clone_repo_to_stage
    install_stage_into_install_dir
else
    clone_repo_to_stage
    install_stage_into_install_dir
fi

# ─── Directory Structure ──────────────────────────────────────────────────────
section "Preparing data directories"

mkdir -p "${DATA_DIR}/gitea"
mkdir -p "${DATA_DIR}/traefik"
if [ ! -f "${DATA_DIR}/traefik/routes.yml" ]; then
        cat > "${DATA_DIR}/traefik/routes.yml" << 'TRAEFIK_ROUTES'
http:
    routers: {}
    services: {}
TRAEFIK_ROUTES
fi
chmod 700 "${INSTALL_DIR}"
log "Directories ready"

# ─── Configuration ────────────────────────────────────────────────────────────
section "Configuration"

ENV_FILE="${INSTALL_DIR}/.env"

# If .env already exists, load existing values as defaults
CF_TOKEN_DEFAULT=""
DOMAIN_DEFAULT="localvsp.local"
GITEA_SECRET_DEFAULT=""
if [ -f "${ENV_FILE}" ]; then
    CF_TOKEN_DEFAULT=$(grep -Po '(?<=CLOUDFLARE_TUNNEL_TOKEN=).*' "${ENV_FILE}" || true)
    DOMAIN_DEFAULT=$(grep -Po '(?<=VSP_DOMAIN=).*' "${ENV_FILE}" || true)
    GITEA_SECRET_DEFAULT=$(grep -Po '(?<=GITEA_SECRET_KEY=).*' "${ENV_FILE}" || true)
    warn "Existing .env found — shown values are current settings."
fi

if [ -z "${GITEA_SECRET_DEFAULT}" ]; then
    GITEA_SECRET_DEFAULT="$(random_secret)"
fi

# Cloudflare Tunnel Token
echo ""
echo -e "  ${BOLD}Cloudflare Tunnel Token${NC}"
echo "  Get yours at: https://one.dash.cloudflare.com → Networks → Tunnels"
echo -e "  ${YELLOW}Leave blank to skip Cloudflare (local access only)${NC}"
echo -n "  Token [${CF_TOKEN_DEFAULT:0:8}...]: "
read -r CF_TOKEN
CF_TOKEN="${CF_TOKEN:-$CF_TOKEN_DEFAULT}"

# Domain
echo ""
echo -e "  ${BOLD}Base Domain${NC}"
echo "  Used for app/site subdomains (e.g. blog.yourdomain.com)"
echo -n "  Domain [${DOMAIN_DEFAULT}]: "
read -r VSP_DOMAIN
VSP_DOMAIN="${VSP_DOMAIN:-$DOMAIN_DEFAULT}"

# Write .env while preserving any unrelated keys already present
TMP_ENV=$(mktemp)
if [ -f "${ENV_FILE}" ]; then
    grep -Ev '^(CLOUDFLARE_TUNNEL_TOKEN|VSP_DOMAIN|GITEA_SECRET_KEY)=' "${ENV_FILE}" > "${TMP_ENV}" || true
fi
{
    cat "${TMP_ENV}" 2>/dev/null || true
    echo "CLOUDFLARE_TUNNEL_TOKEN=${CF_TOKEN}"
    echo "VSP_DOMAIN=${VSP_DOMAIN}"
    echo "GITEA_SECRET_KEY=${GITEA_SECRET_DEFAULT}"
} > "${TMP_ENV}.new"
mv "${TMP_ENV}.new" "${ENV_FILE}"
rm -f "${TMP_ENV}"
chmod 600 "${ENV_FILE}"
log ".env written to ${ENV_FILE}"

# ─── Build & Start ────────────────────────────────────────────────────────────
section "Building and starting services"

cd "${INSTALL_DIR}"

# Determine which compose profiles to activate
PROFILES=""
if [ -n "${CF_TOKEN}" ]; then
    PROFILES="--profile cloudflare"
    info "Cloudflare Tunnel enabled"
else
    warn "No Cloudflare token — cloudflared will not start. Services will be local-only."
fi

info "Running docker compose up --build (this may take a few minutes on first run) …"
# shellcheck disable=SC2086
docker compose ${PROFILES} up --build --detach --remove-orphans
log "Stack started"

# ─── Health Checks ────────────────────────────────────────────────────────────
section "Waiting for services to become healthy"

check_http() {
    local name="$1" url="$2" retries=20 wait=3
    info "Waiting for ${name} at ${url} …"
    for i in $(seq 1 $retries); do
        if curl -sf --max-time 3 "${url}" -o /dev/null 2>/dev/null; then
            log "${name} is up"
            return 0
        fi
        sleep "$wait"
    done
    warn "${name} did not respond in time (it may still be starting)"
    return 1
}

check_http "Management UI" "http://localhost:8081"
check_http "Gitea"         "http://localhost:3000"

# ─── Gitea Admin Setup ────────────────────────────────────────────────────────
section "Configuring Gitea admin account"

info "Creating admin user (vsp:vsp) …"
if docker exec -u git gitea gitea admin user create \
    --admin \
    --username vsp \
    --password vsp \
    --email admin@localvsp.local \
    --must-change-password=false 2>&1 | grep -qE 'created|already'; then
    log "Admin user 'vsp' ready"
elif docker exec -u git gitea gitea admin user list 2>&1 | grep -q 'vsp'; then
    log "Admin user 'vsp' already exists"
else
    # User creation output is sometimes on stderr — run once more and show output
    docker exec -u git gitea gitea admin user create \
        --admin \
        --username vsp \
        --password vsp \
        --email admin@localvsp.local \
        --must-change-password=false 2>&1 || \
    warn "Gitea user creation returned non-zero — user may already exist"
fi

# ─── Done ─────────────────────────────────────────────────────────────────────
HOST_IP=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}${BOLD}  Local VSP is running!${NC}"
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
if [ -n "${VSP_DOMAIN}" ]; then
echo -e "  Example: subdomain ${BOLD}blog${NC} + domain ${BOLD}${VSP_DOMAIN}${NC} → ${BOLD}http://blog.${VSP_DOMAIN}${NC}"
echo -e "  Cloudflare tunnel setup: point ${BOLD}*.${VSP_DOMAIN}${NC} to ${BOLD}http://traefik:80${NC} inside the tunnel."
echo -e "  ${YELLOW}Note:${NC} Only apps/sites with an assigned subdomain are public. Management UI and Gitea stay local-only."
echo -e "  ${YELLOW}Tip:${NC} Add ${BOLD}${HOST_IP}${NC} to your DNS or hosts file for *.${VSP_DOMAIN}"
else
echo -e "  Assign a base domain later in Settings if you want public app/site hostnames through Cloudflare."
echo -e "  ${YELLOW}Note:${NC} Management UI and Gitea stay local-only by default."
fi
echo -e "  ${BOLD}Management UI${NC}  →  http://${HOST_IP}:8081"
echo -e "  ${BOLD}Gitea          ${NC}  →  http://${HOST_IP}:3000"
echo -e "  ${BOLD}Traefik Panel  ${NC}  →  http://${HOST_IP}:8080"
echo ""
if [ -n "${CF_TOKEN}" ]; then
echo -e "  ${BOLD}Cloudflare${NC}     →  Configure public hostnames at one.dash.cloudflare.com"
echo -e "                     *.${VSP_DOMAIN}       →  http://traefik:80"
fi
echo ""
echo -e "  ${CYAN}Gitea login:${NC}            http://${HOST_IP}:3000  —  ${BOLD}vsp / vsp${NC}"
echo ""
echo -e "  ${BOLD}Network Share (Windows)${NC}"
echo -e "  Explorer address bar:   ${BOLD}\\\\\\\\${HOST_IP}\\\\vsp-home${NC}  (user: vsp / vsp)"
echo -e "  Deploy folders:         ${BOLD}\\\\\\\\${HOST_IP}\\\\vsp-home\\\\docker${NC}   ← docker-compose projects"
echo -e "                          ${BOLD}\\\\\\\\${HOST_IP}\\\\vsp-home\\\\html${NC}      ← flat HTML sites"
echo ""
echo -e "  ${BOLD}Traefik Subdomain Routing${NC}"
echo -e "  Set a subdomain per site/project in the dashboard (stopped state only)."
echo -e "  Example: subdomain ${BOLD}blog${NC} + domain ${BOLD}${VSP_DOMAIN}${NC} → ${BOLD}http://blog.${VSP_DOMAIN}${NC}"
echo -e "  Cloudflare tunnel setup: point ${BOLD}*.${VSP_DOMAIN}${NC} to ${BOLD}http://traefik:80${NC} inside the tunnel."
echo -e "  ${YELLOW}Tip:${NC} Add ${BOLD}${HOST_IP}${NC} to your DNS or hosts file for *.${VSP_DOMAIN}"
echo ""
echo -e "  To view logs:    ${BOLD}cd ${INSTALL_DIR} && docker compose logs -f${NC}"
echo -e "  To update:       ${BOLD}curl -sSL <install_url> | sudo bash${NC}"
echo ""
