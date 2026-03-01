#!/usr/bin/env bash
set -euo pipefail

# ─── VPS Setup Script ────────────────────────────────────────────────
# Run as root on a fresh Ubuntu 22.04/24.04 droplet.
# Usage: bash scripts/setup-vps.sh
# ─────────────────────────────────────────────────────────────────────

REPO_URL="${REPO_URL:-https://github.com/poimgs/developer-space.git}"
APP_DIR="/opt/developer-space"
DEPLOY_USER="deploy"

echo "==> Updating system packages"
apt-get update && apt-get upgrade -y

echo "==> Installing Docker"
apt-get install -y ca-certificates curl gnupg
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  > /etc/apt/sources.list.d/docker.list

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo "==> Creating deploy user"
if ! id "$DEPLOY_USER" &>/dev/null; then
  adduser --disabled-password --gecos "" "$DEPLOY_USER"
fi
usermod -aG docker "$DEPLOY_USER"

# Set up SSH directory for deploy user
mkdir -p /home/$DEPLOY_USER/.ssh
chmod 700 /home/$DEPLOY_USER/.ssh
touch /home/$DEPLOY_USER/.ssh/authorized_keys
chmod 600 /home/$DEPLOY_USER/.ssh/authorized_keys
chown -R $DEPLOY_USER:$DEPLOY_USER /home/$DEPLOY_USER/.ssh
echo "  -> Add the deploy SSH public key to /home/$DEPLOY_USER/.ssh/authorized_keys"

echo "==> Configuring UFW firewall"
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "==> Cloning repository"
if [ ! -d "$APP_DIR" ]; then
  git clone "$REPO_URL" "$APP_DIR"
  chown -R $DEPLOY_USER:$DEPLOY_USER "$APP_DIR"
else
  echo "  -> $APP_DIR already exists, skipping clone"
fi

echo "==> Creating .env template"
if [ ! -f "$APP_DIR/.env" ]; then
  cat > "$APP_DIR/.env" <<'ENVEOF'
# Server
PORT=8080
FRONTEND_URL=https://CHANGE_ME_DOMAIN

# Database (internal Docker network)
DATABASE_URL=postgres://coworkspace:CHANGE_ME@postgres:5432/coworkspace?sslmode=disable
POSTGRES_PASSWORD=CHANGE_ME

# Auth
SESSION_SECRET=CHANGE_ME

# Email (Resend)
RESEND_API_KEY=CHANGE_ME
RESEND_FROM_EMAIL=CHANGE_ME

# Telegram (optional)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Logging
LOG_LEVEL=info
ENVEOF
  chown $DEPLOY_USER:$DEPLOY_USER "$APP_DIR/.env"
  echo "  -> Edit $APP_DIR/.env with real production values"
else
  echo "  -> $APP_DIR/.env already exists, skipping"
fi

echo ""
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. Add deploy SSH public key to /home/$DEPLOY_USER/.ssh/authorized_keys"
echo "  2. Edit $APP_DIR/.env with production values"
echo "  3. cd $APP_DIR && docker compose -f docker-compose.prod.yml up -d"
echo "  4. Seed admin: docker compose -f docker-compose.prod.yml exec api /api seed-admin --email <email> --name <name>"
