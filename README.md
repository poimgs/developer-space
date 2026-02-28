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
