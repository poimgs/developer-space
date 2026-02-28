# Spec Index

Table of contents, consolidated decision log, and cross-reference map for the co-working space app specifications.

## Spec Table

| # | File | Title | Summary |
|---|------|-------|---------|
| 00 | [00-overview.md](./00-overview.md) | Overview | Vision, tech stack, data model, system architecture |
| 01 | [01-member-management.md](./01-member-management.md) | Member Management | Admin CRUD for members, invitation email, delete guards |
| 02 | [02-authentication.md](./02-authentication.md) | Authentication | Magic link auth flow, session cookies, profile self-edit |
| 03 | [03-session-management.md](./03-session-management.md) | Session Management | Admin CRUD for sessions, recurring weekly sessions, status lifecycle |
| 04 | [04-rsvp.md](./04-rsvp.md) | RSVP | Member RSVP to sessions, atomic capacity check, guest list |
| 05 | [05-telegram-notifications.md](./05-telegram-notifications.md) | Notifications (Telegram + Email) | Telegram broadcast, email to RSVPed members on cancel/reschedule |
| 06 | [06-cross-cutting-concerns.md](./06-cross-cutting-concerns.md) | Cross-Cutting Concerns | Middleware, API conventions, config, rate limiting |
| 07 | [07-deployment.md](./07-deployment.md) | Deployment | Docker Compose, migrations, seeding, production notes |
| 08 | [08-ui-ux.md](./08-ui-ux.md) | UI/UX: App Structure & Workflows | Page inventory, navigation, 8 user workflow descriptions |
| 09 | [09-design-patterns.md](./09-design-patterns.md) | Design Patterns | Color system, dark mode, responsive breakpoints, shared components |

## Decision Log

| # | Decision | Rationale | Spec |
|---|----------|-----------|------|
| 1 | Sessions list as home page | No separate dashboard; sessions are the core experience for all users | 08 |
| 2 | Mobile-first responsive design | Most members check sessions on their phone; Tailwind mobile-first approach | 09 |
| 3 | Sessions grouped by date | Date headers ("Monday, March 9") provide clear temporal context | 08, 03 |
| 4 | 4-week default view | `to` defaults to today + 28 days; balances useful lookahead with performance | 03 |
| 5 | Manual refresh only | React Query refetch on window focus; no polling to minimize server load | 08 |
| 6 | Colored status badges | Green = scheduled, amber = rescheduled, red = canceled; instant visual status | 09 |
| 7 | Illustrated empty states | SVG icon + heading + CTA button; guides users to next action | 09 |
| 8 | Toast notifications for feedback | Auto-dismiss 3–5s; non-blocking, ephemeral feedback pattern | 09 |
| 9 | Modal dialogs for destructive actions | Confirmation required for cancel session, delete member, cancel RSVP | 09 |
| 10 | Neutral palette + indigo accent | Professional, accessible; indigo provides strong contrast in light/dark modes | 09 |
| 11 | Dark mode (class-based + system pref) | Tailwind `darkMode: 'class'`, localStorage override, system preference fallback | 09 |
| 12 | Top navbar with hamburger on mobile | Desktop: visible links; mobile: hamburger → slide-in overlay | 08 |
| 13 | Simple profile page at `/profile` | Members edit name + telegram handle only; no complex settings | 02 |
| 14 | Optional invitation email on member creation | `send_invite` checkbox (default unchecked); fire-and-forget via Resend | 01 |
| 15 | Member self-edit via `PATCH /api/auth/profile` | Name + telegram only; email and admin status controlled by admins | 02 |
| 16 | Recurring sessions (simple weekly repeat) | `repeat_weekly` 0–12; creates N+1 independent sessions in one transaction | 03 |
| 17 | Single summary Telegram message for recurring | One message listing all dates instead of N individual notifications | 05 |
| 18 | Email notifications on cancel/reschedule | Resend emails to RSVPed members; reuses existing Resend integration | 05 |
| 19 | Magic links only (no passwords) | Small trusted group — simpler UX, no password management | 00, 02 |
| 20 | `is_admin` flag on Member | Single-role model is sufficient; no need for RBAC | 00 |
| 21 | Telegram group chat for notifications | Team already uses Telegram; real-time visibility for all members | 00, 05 |
| 22 | Output-only Telegram bot (v1) | Keep scope small; command handling is a future enhancement | 05 |
| 23 | No centralized OpenAPI spec | 7-endpoint app — API contracts live in feature specs | 00 |
| 24 | Feature-based specs (not layer-based) | Each spec is self-contained: stories, criteria, endpoints, UI design | 00 |
| 25 | Resend for email delivery | Simple API, good developer experience, generous free tier | 00 |
| 26 | API-first design | Every feature spec includes full endpoint contracts | 00 |

## Cross-Reference Map

Shows which specs depend on or reference which others.

```
09-design-patterns (foundational)
  ↑ referenced by: 01, 02, 03, 04, 08

08-ui-ux (app structure)
  → references: 01, 02, 03, 04, 05, 09

00-overview (data model)
  ↑ referenced by: all specs

01-member-management
  → references: 02 (self-edit), 09 (UI patterns)
  ↑ referenced by: 08 (workflows)

02-authentication
  → references: 01 (member creation), 09 (UI patterns)
  ↑ referenced by: 08 (workflows)

03-session-management
  → references: 04 (RSVP button), 05 (notifications), 09 (UI patterns)
  ↑ referenced by: 04, 05, 08

04-rsvp
  → references: 03 (session context), 05 (notifications), 09 (UI patterns)
  ↑ referenced by: 03, 08

05-notifications
  → references: 03 (session events), 04 (RSVP events)
  ↑ referenced by: 03, 04, 08

06-cross-cutting-concerns
  ↑ referenced by: 02 (rate limiting), 07 (env vars)

07-deployment
  → references: 06 (env vars)
```
