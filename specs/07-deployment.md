# 07 — Deployment

Self-hosted deployment using Docker Compose. The stack consists of three services: API, frontend, and PostgreSQL.

## Docker Compose Services

```yaml
# docker-compose.yml
services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.api
    ports:
      - "${PORT:-8080}:8080"
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "${FRONTEND_PORT:-3000}:80"

  postgres:
    image: postgres:16-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: coworkspace
      POSTGRES_USER: coworkspace
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U coworkspace"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

## Multi-Stage Builds

### API Dockerfile (`Dockerfile.api`)

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /api /api
COPY migrations/ /migrations/
EXPOSE 8080
CMD ["/api"]
```

### Frontend Dockerfile (`Dockerfile.frontend`)

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

## Database Migrations

Use [golang-migrate](https://github.com/golang-migrate/migrate) for database schema migrations.

Migration files live in `migrations/` directory:

```
migrations/
  000001_create_members.up.sql
  000001_create_members.down.sql
  000002_create_space_sessions.up.sql
  000002_create_space_sessions.down.sql
  000003_create_rsvps.up.sql
  000003_create_rsvps.down.sql
  000004_create_magic_tokens.up.sql
  000004_create_magic_tokens.down.sql
```

Migrations run automatically on API startup (before the HTTP server starts), or manually via CLI:

```bash
# Run migrations up
go run ./cmd/api migrate up

# Roll back last migration
go run ./cmd/api migrate down 1
```

## Initial Admin Seeding

A CLI command to create the first admin member:

```bash
go run ./cmd/api seed-admin --email admin@example.com --name "Admin User"
```

This creates a member with `is_admin = true` and `is_active = true`. Exits with an error if the email already exists. After seeding, the admin can log in via magic link and create other members.

## Local Development Workflow

### Prerequisites

- Go 1.23+
- Node.js 20+
- PostgreSQL 16 (or use Docker Compose)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for manual migration management)

### Setup

```bash
# Clone and enter project
git clone <repo-url>
cd developer-space

# Copy environment config
cp .env.example .env
# Edit .env with your values

# Start database (if using Docker)
docker compose up postgres -d

# Run migrations
go run ./cmd/api migrate up

# Seed initial admin
go run ./cmd/api seed-admin --email you@example.com --name "Your Name"

# Start API server (with hot reload via air)
air

# In another terminal — start frontend
cd frontend
npm install
npm run dev
```

### Hot Reload

- **API:** Use [air](https://github.com/air-verse/air) for automatic Go rebuilds on file change.
- **Frontend:** Vite's built-in HMR.

## `.env.example`

```bash
# Server
PORT=8080
FRONTEND_URL=http://localhost:5173

# Database
DATABASE_URL=postgres://coworkspace:localdev@localhost:5432/coworkspace?sslmode=disable
POSTGRES_PASSWORD=localdev

# Auth
SESSION_SECRET=change-me-to-a-random-string

# Email (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxx
RESEND_FROM_EMAIL=noreply@yourdomain.com

# Telegram (optional — leave empty to disable)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Logging
LOG_LEVEL=debug
```

## Production Notes

### HTTPS

The application does not terminate TLS. Use a reverse proxy:

- **Caddy** (recommended — automatic HTTPS via Let's Encrypt):
  ```
  yourdomain.com {
      handle /api/* {
          reverse_proxy api:8080
      }
      handle {
          reverse_proxy frontend:80
      }
  }
  ```
- **Nginx** or **Traefik** are also suitable.

### Backups

- Use `pg_dump` on a cron schedule for PostgreSQL backups:
  ```bash
  pg_dump -U coworkspace coworkspace | gzip > /backups/coworkspace_$(date +%Y%m%d).sql.gz
  ```
- Store backups off-host (S3, rsync to another server, etc.).

### Security Checklist

- [ ] `SESSION_SECRET` is a long, random string (32+ characters).
- [ ] `POSTGRES_PASSWORD` is strong and not the default.
- [ ] `.env` file is not committed to version control.
- [ ] HTTPS is enforced via reverse proxy.
- [ ] Database port (5432) is not exposed to the public internet.
- [ ] `FRONTEND_URL` is set to the actual production domain (for CORS and magic link URLs).
- [ ] Resend sender domain is verified.

## Project Structure

```
developer-space/
├── cmd/
│   └── api/              # Application entrypoint
│       └── main.go
├── internal/
│   ├── handler/          # HTTP handlers
│   ├── middleware/        # Auth, admin, logging, CORS
│   ├── model/            # Data structures
│   ├── repository/       # Database queries
│   ├── service/          # Business logic
│   └── telegram/         # Telegram notification service
├── migrations/           # SQL migration files
├── frontend/             # React + Vite app
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── specs/                # This spec directory
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.frontend
├── .env.example
├── go.mod
└── go.sum
```
