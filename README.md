# Developer Co-Working Space

A web application for managing a physical co-working space where developers coordinate access, scheduling, and capacity. Admins create sessions representing open-space timeslots; members RSVP to reserve spots. All activity is broadcast to a shared Telegram group chat.

## Tech Stack

| Layer          | Technology                          |
|----------------|-------------------------------------|
| Frontend       | React + Vite                        |
| Backend        | Go + Chi router                     |
| Database       | PostgreSQL 16                       |
| Email          | Resend (magic-link authentication)  |
| Notifications  | Telegram Bot API (group chat)       |
| Deployment     | Docker / self-hosted                |

## Architecture

```
┌─────────────┐       ┌──────────────────────────┐       ┌────────────┐
│             │       │         Go API            │       │            │
│  React SPA  │──────▶│  Chi router + middleware  │──────▶│ PostgreSQL │
│  (Vite)     │ HTTP  │                           │       │            │
│             │◀──────│  /api/*                   │◀──────│            │
└─────────────┘       └──────────┬───────┬────────┘       └────────────┘
                                 │       │
                          Resend │       │ Telegram
                          API    │       │ Bot API
                                 ▼       ▼
                          ┌──────┐  ┌──────────┐
                          │Email │  │ TG Group  │
                          │Inbox │  │ Chat      │
                          └──────┘  └──────────┘
```

## Prerequisites

- Go 1.23+
- Node.js 20+
- PostgreSQL 16 (or Docker)
- [air](https://github.com/air-verse/air) (Go hot-reload, optional)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (optional, for manual migrations)

## Quick Start (Docker Compose)

```bash
cp .env.example .env
# Edit .env with your values
docker compose up
```

This starts the API (`:8080`), frontend (`:3000`), and PostgreSQL. Migrations run automatically on API startup.

## Local Development

```bash
# 1. Copy environment config
cp .env.example .env
# Edit .env with your values

# 2. Start PostgreSQL
docker compose up postgres -d

# 3. Run migrations
go run ./cmd/api migrate up

# 4. Seed initial admin
go run ./cmd/api seed-admin --email you@example.com --name "Your Name"

# 5. Start API (with hot reload)
air

# 6. In another terminal — start frontend
cd frontend
npm install
npm run dev
```

The frontend runs at `http://localhost:5173` and proxies API requests to `:8080`.

## Environment Variables

Copy `.env.example` and fill in your values. Key variables:

| Variable             | Description                              |
|----------------------|------------------------------------------|
| `PORT`               | API server port (default `8080`)         |
| `FRONTEND_URL`       | Frontend origin for CORS & magic links   |
| `DATABASE_URL`       | PostgreSQL connection string             |
| `SESSION_SECRET`     | Random string for session signing        |
| `RESEND_API_KEY`     | Resend API key for magic-link emails     |
| `RESEND_FROM_EMAIL`  | Sender address for outgoing email        |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token (optional)            |
| `TELEGRAM_CHAT_ID`   | Telegram group chat ID (optional)        |

## Project Structure

```
developer-space/
├── cmd/
│   └── api/              # Application entrypoint
│       └── main.go
├── internal/
│   ├── config/           # Environment configuration
│   ├── database/         # Database connection & migrations
│   ├── handler/          # HTTP handlers
│   ├── middleware/        # Auth, admin, logging, CORS
│   ├── model/            # Data structures
│   ├── repository/       # Database queries
│   ├── response/         # HTTP response helpers
│   ├── service/          # Business logic
│   └── telegram/         # Telegram notification service
├── migrations/           # SQL migration files
├── frontend/             # React + Vite app
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── specs/                # Feature specifications
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.frontend
├── .env.example
├── go.mod
└── go.sum
```

## Production Deployment

### Architecture

```
Internet --> Caddy (:443 TLS, :80 redirect) --> frontend nginx (:80 internal)
                                                    |           |
                                                /api/*         /*
                                                    v           v
                                              Go API :8080   React SPA
                                                    |
                                              PostgreSQL :5432
```

Only ports 22, 80, and 443 are publicly accessible. Caddy handles TLS automatically via Let's Encrypt.

### VPS Setup

Requires a VPS running Ubuntu 22.04/24.04 with 1GB+ RAM and a domain with a DNS A record pointing to the VPS IP.

#### 1. Install Docker

```bash
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

#### 2. Create a deploy user

```bash
sudo adduser --disabled-password --gecos "" deploy
sudo usermod -aG docker deploy
```

#### 3. Configure firewall

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

#### 4. Clone the repository

```bash
sudo -u deploy git clone https://github.com/poimgs/developer-space.git /home/deploy/developer-space
```

If the repo is private, add a [deploy key](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/managing-deploy-keys) first and clone via SSH.

#### 5. Configure environment

```bash
cd /home/deploy/developer-space
sudo -u deploy cp .env.example .env
sudo -u deploy nano .env
```

Set production values — generate secure passwords with:

```bash
openssl rand -base64 32   # for POSTGRES_PASSWORD
openssl rand -base64 48   # for SESSION_SECRET
```

Key production settings:

| Variable | Production value |
|----------|-----------------|
| `DOMAIN` | Your domain (e.g. `app.example.com`) |
| `FRONTEND_URL` | `https://your-domain.example.com` |
| `DATABASE_URL` | `postgres://coworkspace:<password>@postgres:5432/coworkspace?sslmode=disable` |
| `LOG_LEVEL` | `info` |

#### 6. Start the application

```bash
cd /home/deploy/developer-space
sudo -u deploy docker compose -f docker-compose.prod.yml up -d
```

Caddy automatically provisions a TLS certificate on first startup.

#### 7. Create an admin user

```bash
cd /home/deploy/developer-space
docker compose -f docker-compose.prod.yml exec api /api seed-admin --email admin@example.com --name "Admin Name"
```

#### 8. Verify

```bash
curl https://your-domain.example.com/health
# Should return: {"data":{"status":"ok"}}
```

### CI/CD with GitHub Actions

Pushes to `main` auto-deploy to the VPS via SSH.

#### Deploy SSH key

```bash
ssh-keygen -t ed25519 -f deploy_key -C "github-actions-deploy"
```

- Add the **public key** to `/home/deploy/.ssh/authorized_keys` on the VPS
- Add the **private key** as a GitHub repository secret

#### GitHub Secrets

Add in repo Settings > Secrets and variables > Actions:

| Secret | Value |
|--------|-------|
| `VPS_HOST` | VPS public IP address |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | Contents of the deploy private key |
| `DOMAIN` | Your production domain |

### CI Testing

Pull requests to `main` automatically run Go tests (with PostgreSQL) and frontend tests (Vitest).

## Specs

Detailed feature specifications live in the [`specs/`](specs/) directory:

| # | File | Scope |
|---|------|-------|
| 00 | [Overview](specs/00-overview.md) | Vision, data model, architecture |
| 01 | [Member Management](specs/01-member-management.md) | Admin CRUD, invitation email |
| 02 | [Authentication](specs/02-authentication.md) | Magic-link auth, profile self-edit |
| 03 | [Session Management](specs/03-session-management.md) | Admin CRUD, recurring sessions |
| 04 | [RSVP](specs/04-rsvp.md) | Member RSVP to sessions |
| 05 | [Telegram Notifications](specs/05-telegram-notifications.md) | Telegram + email notifications |
| 06 | [Cross-Cutting Concerns](specs/06-cross-cutting-concerns.md) | Middleware, conventions, config |
| 07 | [Deployment](specs/07-deployment.md) | Docker, migrations, seeding |
| 08 | [UI/UX](specs/08-ui-ux.md) | App structure, navigation, workflows |
| 09 | [Design Patterns](specs/09-design-patterns.md) | Design system, shared components |
