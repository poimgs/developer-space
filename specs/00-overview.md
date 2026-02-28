# Developer Co-Working Space App — Overview

A web application for managing a physical co-working space where developers coordinate access, scheduling, and capacity. Admins create sessions representing open-space timeslots; members RSVP to reserve spots. All activity is broadcast to a shared Telegram group chat.

## Tech Stack

| Layer      | Technology              |
|------------|-------------------------|
| Frontend   | React + Vite            |
| Backend    | Go + Chi router         |
| Database   | PostgreSQL              |
| Email      | Resend (magic links)    |
| Notifications | Telegram Bot API (group chat) |
| Deployment | Docker / self-hosted    |

## Data Model

### Member

| Field            | Type         | Constraints                        |
|------------------|--------------|------------------------------------|
| `id`             | UUID         | PK, generated                      |
| `email`          | varchar(255) | UNIQUE, NOT NULL                   |
| `name`           | varchar(255) | NOT NULL                           |
| `telegram_handle`| varchar(255) | nullable                           |
| `is_admin`       | boolean      | NOT NULL, default `false`          |
| `is_active`      | boolean      | NOT NULL, default `true`           |
| `created_at`     | timestamptz  | NOT NULL, default `now()`          |
| `updated_at`     | timestamptz  | NOT NULL, default `now()`          |

### SpaceSession

| Field        | Type         | Constraints                              |
|--------------|--------------|------------------------------------------|
| `id`         | UUID         | PK, generated                            |
| `title`      | varchar(255) | NOT NULL                                 |
| `description`| text         | nullable                                 |
| `date`       | date         | NOT NULL                                 |
| `start_time` | time         | NOT NULL                                 |
| `end_time`   | time         | NOT NULL                                 |
| `capacity`   | integer      | NOT NULL, > 0                            |
| `status`     | varchar(20)  | NOT NULL, default `scheduled` — enum: `scheduled`, `shifted`, `canceled` |
| `created_by` | UUID         | FK → Member(id), NOT NULL                |
| `created_at` | timestamptz  | NOT NULL, default `now()`                |
| `updated_at` | timestamptz  | NOT NULL, default `now()`                |

### RSVP

| Field        | Type        | Constraints                              |
|--------------|-------------|------------------------------------------|
| `id`         | UUID        | PK, generated                            |
| `session_id` | UUID        | FK → SpaceSession(id), NOT NULL          |
| `member_id`  | UUID        | FK → Member(id), NOT NULL                |
| `created_at` | timestamptz | NOT NULL, default `now()`                |

**Unique constraint:** `(session_id, member_id)` — one RSVP per member per session.

### MagicToken

| Field        | Type        | Constraints                              |
|--------------|-------------|------------------------------------------|
| `id`         | UUID        | PK, generated                            |
| `member_id`  | UUID        | FK → Member(id), NOT NULL                |
| `token_hash` | varchar(64) | NOT NULL, indexed                        |
| `expires_at` | timestamptz | NOT NULL                                 |
| `used_at`    | timestamptz | nullable                                 |
| `created_at` | timestamptz | NOT NULL, default `now()`                |

## System Architecture

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

## Spec Index

| #  | File                          | Scope                                    |
|----|-------------------------------|------------------------------------------|
| 00 | `00-overview.md`              | This file — vision, data model           |
| 01 | `01-member-management.md`     | Admin CRUD for members, invitation email |
| 02 | `02-authentication.md`        | Magic link auth flow, profile self-edit  |
| 03 | `03-session-management.md`    | Admin CRUD for sessions, recurring       |
| 04 | `04-rsvp.md`                  | Member RSVP to sessions                  |
| 05 | `05-telegram-notifications.md`| Telegram + email notifications           |
| 06 | `06-cross-cutting-concerns.md`| Middleware, conventions, config           |
| 07 | `07-deployment.md`            | Docker, migrations, seeding              |
| 08 | `08-ui-ux.md`                 | App structure, navigation, workflows     |
| 09 | `09-design-patterns.md`       | Design system, shared component patterns |

See [specs/index.md](./index.md) for full decision log and cross-reference map.

## Decisions Log

See [specs/index.md](./index.md) for the full consolidated decision log with spec references.

| Decision | Rationale |
|----------|-----------|
| Magic links only (no passwords) | Small trusted group — simpler UX, no password management |
| `is_admin` flag on Member | Single-role model is sufficient; no need for RBAC |
| Telegram group chat for notifications | Team already uses Telegram; real-time visibility for all members |
| Output-only Telegram bot (v1) | Keep scope small; command handling is a future enhancement |
| No centralized OpenAPI spec | 7-endpoint app — API contracts live in feature specs |
| Feature-based specs (not layer-based) | Each spec is self-contained: stories, criteria, endpoints, UI design |
| Resend for email delivery | Simple API, good developer experience, generous free tier |
| API-first design | Every feature spec includes full endpoint contracts so any client (web, Telegram bot, future MCP server) can consume them |
| Sessions list as home page | No separate dashboard; sessions are the core experience |
| Mobile-first responsive design | Tailwind mobile-first, top navbar collapses to hamburger |
| Dark mode (class-based) | `darkMode: 'class'`, system preference detection + localStorage |
| Recurring weekly sessions | Simple repeat: N independent sessions, `repeat_weekly` 0–12 |
| Email notifications on cancel/reschedule | Resend emails to RSVPed members, fire-and-forget |
| Member self-edit profile | `PATCH /api/auth/profile` — name + telegram only |
| Optional invitation email | Checkbox on member creation, fire-and-forget via Resend |
