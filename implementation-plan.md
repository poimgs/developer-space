# Implementation Plan: Co-Working Space Management App

## Context

Building a greenfield web app for managing a physical co-working space. The project has detailed specs (`specs/00-09`) but zero existing code. The app allows admins to manage members and sessions, members to RSVP, and broadcasts activity to a Telegram group + email.

**Stack:** Go (Chi) + PostgreSQL + React (Vite, Tailwind, React Query) + Docker
**Auth:** Passwordless magic links via Resend
**Notifications:** Telegram Bot API (fire-and-forget)
**Testing:** Go tests + React tests alongside each phase

---

## Phase 0: Project Scaffolding ✅ COMPLETED

Establish Go module, directory structure, config loading, database connection, middleware stack, and a health endpoint.

**Status:** Done — all files created, 18 unit tests passing.

### Files to create
- `go.mod` — module init, deps: `chi/v5`, `pgx/v5`, `golang-migrate/v4`, `gorilla/securecookie`, `google/uuid`, `joho/godotenv`, `go-chi/cors`
- `.env.example` — all env vars from spec 06
- `.gitignore` — Go, `.env`, `tmp/`, `frontend/node_modules/`, `frontend/dist/`
- `.air.toml` — hot reload config
- `cmd/api/main.go` — config load, DB connect, Chi router, global middleware, `GET /health`, graceful shutdown
- `internal/config/config.go` — `Config` struct, `Load()`, validation, defaults
- `internal/database/database.go` — `Connect()` (pgxpool), `Close()`
- `internal/response/response.go` — `JSON(w, status, data)`, `Error(w, status, msg)`, `ValidationError(w, details)`
- `internal/middleware/request_id.go` — UUID per request, `X-Request-Id` header + context
- `internal/middleware/logging.go` — `log/slog` JSON, logs method/path/status/duration/request_id
- `internal/middleware/cors.go` — wraps go-chi/cors, origin from `FRONTEND_URL`, credentials allowed
- `internal/middleware/content_type.go` — reject non-JSON POST/PATCH with 415
- `internal/middleware/recovery.go` — panic recovery, logs + returns 500 JSON

### Tests
- `internal/config/config_test.go` — env loading, defaults, required validation
- `internal/response/response_test.go` — envelope format verification
- `internal/middleware/*_test.go` — request ID generation, content type rejection

### Verify
- `go run ./cmd/api` starts, connects to PostgreSQL
- `curl localhost:8080/health` → `{"data":{"status":"ok"}}` with `X-Request-Id` header
- Structured JSON logs appear in terminal

---

## Phase 1: Database Migrations & Models ✅ COMPLETED

Create all migration files and Go model structs. API auto-runs migrations on startup.

**Status:** Done — 8 migration files (4 up + 4 down), 4 model files, migrate runner, CLI subcommands, 17 new unit tests passing (30 total).

### Migrations (in `migrations/`)
1. `000001_create_members` — members table with `UNIQUE(email)`, `updated_at` default
2. `000002_create_space_sessions` — sessions table with `CHECK(end_time > start_time)`, `CHECK(capacity > 0)`, status enum constraint, FK to members
3. `000003_create_rsvps` — rsvps table with `UNIQUE(session_id, member_id)`, FKs
4. `000004_create_magic_tokens` — magic_tokens table with index on `token_hash`

### Files
- `internal/database/migrate.go` — `RunMigrations()`, `MigrateDown()` using golang-migrate with pgx5 driver, called from main before server start
- `internal/model/member.go` — `Member`, `CreateMemberRequest`, `UpdateMemberRequest` structs
- `internal/model/session.go` — `SpaceSession`, `CreateSessionRequest`, `UpdateSessionRequest` structs
- `internal/model/rsvp.go` — `RSVP`, `RSVPWithMember`, `RSVPMember` structs
- `internal/model/magic_token.go` — `MagicToken` struct
- Update `cmd/api/main.go` — CLI subcommands: `migrate up`, `migrate down N`, `seed-admin --email --name`. Auto-runs migrations on server start.

### Tests
- `internal/database/migrate_test.go` — pgxURL conversion (postgres://, postgresql://, passthrough, empty)
- `internal/model/member_test.go` — JSON serialization, nil telegram_handle, create/update request parsing
- `internal/model/session_test.go` — JSON serialization, omitempty on rsvp_count, create/update request parsing
- `internal/model/rsvp_test.go` — RSVP and RSVPWithMember JSON serialization
- `internal/model/magic_token_test.go` — token_hash omitted from JSON, used_at nullable

### Verify
- `go run ./cmd/api migrate up` creates all tables
- `go run ./cmd/api migrate down 4` drops them cleanly
- Constraints work: `end_time < start_time` rejected, duplicate email rejected

---

## Phase 2: Member Management (Admin CRUD) ✅ COMPLETED

5 endpoints for admin member directory. Auth middleware not enforced yet (open for testing, locked in Phase 3).

**Status:** Done — repository, service, handler, and route registration created. `seed-admin` CLI already existed from Phase 1. 29 new unit tests passing (59 total). `NoopEmailSender` used until Resend integration in Phase 3.

### Files
- `internal/repository/member.go` — `Create`, `List` (with active filter), `GetByID`, `GetByEmail`, `Update` (dynamic PATCH), `Delete`, `HasRSVPs`
- `internal/service/member.go` — business logic: validation, strip `@` from telegram_handle, delete guard (check RSVPs), `send_invite` handling (fire-and-forget invitation email via Resend). Sentinel errors: `ErrNotFound`, `ErrDuplicateEmail`, `ErrHasRSVPs`. `ValidationError` type with field-level details. `MemberRepo` interface for testability. `EmailSender` interface + `NoopEmailSender`.
- `internal/handler/member.go` — HTTP handlers mapping service errors to status codes (201/200/204/404/409/422)
- `internal/handler/routes.go` — central route registration, member routes under `/api/members`
- `cmd/api/main.go` — updated to wire repository → service → handler → routes

### Tests
- `internal/service/member_test.go` — 16 unit tests with mocked repository (create, validate, duplicate email, telegram strip, email normalization, get/update/delete, active filter, empty telegram handle)
- `internal/handler/member_test.go` — 13 httptest request/response tests (201/422/409 create, 200 list, 404/400 get, 200/404 update, 204/409/404 delete, 400 invalid JSON)

### Verify
- `seed-admin` creates an admin member
- Full CRUD via curl: POST (201), GET list (200), GET by id (200), PATCH (200), DELETE (204)
- Duplicate email → 409, delete with RSVPs → 409

---

## Phase 3: Authentication (Magic Links) ✅ COMPLETED

Full magic link flow, session cookies, auth/admin middleware. After this phase all endpoints are protected.

**Status:** Done — repository, email service, auth service, auth/admin/rate-limit middleware, auth handler, route wiring, and background token cleanup all created. 42 new unit tests passing (107 total).

### Files
- `internal/repository/magic_token.go` — `Create`, `FindValidByHash`, `MarkUsed`, `CountRecentByEmail`, `CleanExpired`
- `internal/service/email.go` — `ResendEmailSender` with `SendMagicLink()` and `SendInvitation()` via Resend HTTP API. Replaces `NoopEmailSender` in main.go wiring.
- `internal/service/auth.go` — `RequestMagicLink` (generate token, SHA-256 hash, store, send email — always returns nil to prevent enumeration), `VerifyToken` (hash lookup, mark used, check active), `CreateSessionCookie` / `ValidateSessionCookie` (gorilla/securecookie, 30-day, HttpOnly, Secure derived from `IsSecure()`, SameSite=Lax), `UpdateProfile` (name + telegram_handle only, validates name not empty, strips `@`), `CleanExpiredTokens` (background cleanup). `MagicTokenRepo` and `MagicLinkSender` interfaces for testability.
- `internal/middleware/auth.go` — validate cookie, load member via `CookieValidator` + `MemberLookup` interfaces, store in context, 401 if invalid/inactive. Helper: `MemberFromContext(ctx)`
- `internal/middleware/admin.go` — check `is_admin`, 403 if not
- `internal/middleware/rate_limit.go` — in-memory per-key `RateLimiter` with configurable max + window. `RateLimit` middleware factory with key extraction function. Rate limiting for magic link is enforced at the service layer (5/email/hour via `CountRecentByEmail`).
- `internal/handler/auth.go` — `RequestMagicLink` (always 200, 429 on rate limit), `Verify` (302 redirect + set cookie or 401), `Me` (200 with user), `Logout` (clear cookie, 200), `UpdateProfile` (PATCH, 200 or 422)
- Update `internal/handler/routes.go` — auth routes (public: magic-link, verify; authenticated: me, logout, profile). Member routes now protected with auth + admin middleware.
- Update `cmd/api/main.go` — wires `MagicTokenRepository`, `ResendEmailSender`, `AuthService`, `AuthHandler`. Background goroutine cleans expired tokens every hour.

### Route structure after this phase
```
POST  /api/auth/magic-link     [public]
GET   /api/auth/verify          [public]
POST  /api/auth/logout          [auth required]
GET   /api/auth/me              [auth required]
PATCH /api/auth/profile         [auth required]
/api/members/*                  [auth + admin required]
```

### Tests
- `internal/service/auth_test.go` — 18 tests: token generation/hashing, request magic link (active/unknown/inactive/rate-limited/empty), verify token (valid/invalid/expired/used/inactive member), cookie encode/decode/invalid/clear, profile update (valid name/empty name/strip @), clean expired
- `internal/middleware/auth_test.go` — 6 tests: valid cookie, no cookie, invalid cookie, inactive member, member not found, empty context
- `internal/middleware/admin_test.go` — 3 tests: admin, non-admin, no member in context
- `internal/middleware/rate_limit_test.go` — 5 tests: allows up to max, blocks after max, different keys, resets after window, middleware returns 429
- `internal/handler/auth_test.go` — 10 tests: magic link 200/unknown/bad JSON, verify missing/invalid token, me unauthenticated/authenticated, logout, profile update 200/422

### Verify
- Magic link flow end-to-end (check DB for token, verify endpoint sets cookie)
- `GET /api/auth/me` returns user with valid cookie
- `PATCH /api/auth/profile` updates name + telegram, ignores other fields
- `PATCH /api/auth/profile` with empty name → 422
- Unauthenticated → 401, non-admin → 403 on member endpoints
- 6th magic link request in 1 hour → 429
- Expired/used token → 401

---

## Phase 4: Session Management ✅ COMPLETED

Admin CRUD for sessions, all members can list/view. Telegram notifications stubbed via `Notifier` interface.

**Status:** Done — Notifier interface, session repository, service, handler, and route wiring created. 27 new unit tests passing (134 total). `NoopNotifier` used until Telegram integration in Phase 5. `GetByIDForUpdate` deferred to Phase 5 (RSVP atomicity).

### Files
- `internal/service/notifier.go` — `Notifier` interface: `SessionCreated`, `SessionsCreatedRecurring`, `SessionShifted`, `SessionCanceled`, `MemberRSVPed`, `MemberCanceledRSVP`
- `internal/service/noop_notifier.go` — no-op implementation (used until Phase 5)
- `internal/repository/session.go` — `Create`, `CreateBatch` (multi-insert in single transaction for recurring), `List` (with date range + status filters, includes computed `rsvp_count` via LEFT JOIN, `to` defaults to today+28d), `GetByID` (with rsvp_count), `Update` (dynamic PATCH with optional status change), `Cancel` (set status=canceled), `GetRSVPCount`
- `internal/service/session.go` — validation (date >= today, end > start, capacity > 0, `repeat_weekly` 0–12), status lifecycle (scheduled→shifted on date/time change, →canceled is terminal), capacity guard (cannot reduce below RSVP count), `createRecurring` method (creates N+1 sessions in single transaction, fires `SessionsCreatedRecurring` notifier), notifier calls via fire-and-forget goroutines. Sentinel errors: `ErrSessionNotFound`, `ErrSessionCanceled`, `ErrAlreadyCanceled`, `ErrCapacityBelowRSVPs`.
- `internal/handler/session.go` — HTTP handlers, `created_by` extracted from auth context via `MemberFromContext`. Returns `any` (single session or array) on create when `repeat_weekly > 0`.
- Update `internal/handler/routes.go` — session routes: GET list/detail (auth required), POST/PATCH/DELETE (admin required). Added `sessionHandler` parameter to `RegisterRoutes`.
- Update `cmd/api/main.go` — wires `SessionRepository`, `NoopNotifier`, `SessionService`, `SessionHandler`.

### Route structure after this phase
```
GET    /api/sessions          [auth required]
GET    /api/sessions/{id}     [auth required]
POST   /api/sessions          [auth + admin required]
PATCH  /api/sessions/{id}     [auth + admin required]
DELETE /api/sessions/{id}     [auth + admin required]
```

### Status lifecycle
```
scheduled → shifted (date/time changed) → canceled (terminal)
scheduled → canceled (terminal)
```

### Tests
- `internal/service/session_test.go` — 27 unit tests with mocked repository and notifier: create (valid, missing title, invalid date, past date, end before start, zero capacity, repeat_weekly > 12, multiple errors, recurring N+1), get by ID (found, not found), update (title only, date→shifted, start_time→shifted, already shifted stays shifted, canceled rejected, not found, capacity below RSVPs, capacity above RSVPs OK, empty title, invalid date format, end before start cross-field), cancel (scheduled, shifted, already canceled, not found), list empty slice

### Verify
- Admin creates session (201, status=scheduled)
- Admin creates recurring session (`repeat_weekly: 3`) → 201 returns array of 4 sessions with correct dates
- `repeat_weekly: 13` → 422
- Update date → status becomes "shifted"
- Cancel → status "canceled", further edits rejected (422)
- Non-admin can list/get (200) but not create/update/cancel (403)
- Session list defaults to today to today+28d, sorted by date

---

## Phase 5: RSVP & Telegram Notifications ✅ COMPLETED

Atomic RSVP with capacity enforcement + Telegram broadcast + email notifications. Completes the entire backend API.

**Status:** Done — all RSVP, Telegram, and email notification features implemented. 194 total unit tests passing.

**RSVP Status:** Done — repository (atomic `SELECT FOR UPDATE` transaction), service (error mapping, fire-and-forget notifications), handler (full error→status code mapping), and route wiring created. 27 RSVP unit tests passing.

**Telegram Status:** Done — MarkdownV2 escape helper, TelegramService (HTTP client, no-op when unconfigured), TelegramNotifier (implements Notifier interface with 6 message types), and main.go wiring (auto-selects TelegramNotifier or NoopNotifier based on env vars).

**Email Notification Status:** Done — `NotificationEmailSender` interface, `RSVPMemberLister` interface, email templates (cancel/reschedule per spec 05), `SessionService` integration (fire-and-forget goroutines after cancel/shift), `ListEmailsBySession` repository method, main.go wiring. 15 new unit tests passing (194 total).

### RSVP files ✅
- `internal/repository/rsvp.go` — `CreateAtomic` uses transaction with `SELECT FOR UPDATE` for atomic capacity check (BEGIN → lock session row → verify not canceled/past → count RSVPs → check capacity → check duplicate → INSERT → COMMIT). Returns `RSVPTxResult` (RSVP + locked session). Also: `Delete` (with past-session guard), `ListBySession` (JOIN members, return name + telegram_handle, ordered by created_at). Sentinel errors: `ErrRSVPSessionCanceled`, `ErrRSVPSessionPast`, `ErrRSVPSessionFull`, `ErrRSVPDuplicate`, `ErrRSVPNotFound`.
- `internal/service/rsvp.go` — `RSVPRepo` interface, `MemberGetter` interface, maps repo errors to service-level errors, fires notifier **after** transaction commits (goroutine, fire-and-forget)
- `internal/handler/rsvp.go` — `RSVP` (201), `CancelRSVP` (204), `ListRSVPs` (200). Member ID from auth context. Full error→HTTP status mapping (404/409/422).
- Updated `internal/handler/routes.go` — RSVP routes under `/api/sessions/{id}/rsvp` and `/api/sessions/{id}/rsvps` (auth required).
- Updated `cmd/api/main.go` — wires `RSVPRepository`, `RSVPService`, `RSVPHandler`.

### RSVP tests ✅
- `internal/service/rsvp_test.go` — 13 unit tests: create (success, not found, canceled, past, full, duplicate), cancel (success, not found, rsvp not found, past), list (success, empty, not found)
- `internal/handler/rsvp_test.go` — 14 httptest tests: RSVP 201/404/409-dup/409-full/422-canceled/422-past/400-invalid-id, cancel 204/404-session/404-rsvp/422-past, list 200/404/200-empty

### Route structure after RSVP
```
POST   /api/sessions/{id}/rsvp    [auth required]
DELETE /api/sessions/{id}/rsvp    [auth required]
GET    /api/sessions/{id}/rsvps   [auth required]
```

### Telegram files ✅
- `internal/telegram/escape.go` — `EscapeMarkdownV2()` for all 18 special chars + backslash (escaped first to avoid double-escaping)
- `internal/telegram/telegram.go` — `TelegramService` with `SendMessage()`. POST to Telegram Bot API with MarkdownV2 parse_mode. No-op if `TELEGRAM_BOT_TOKEN` or `TELEGRAM_CHAT_ID` are empty. Logs failures at warn level, never returns error to caller. `apiBase` field allows test injection.
- `internal/telegram/notifier.go` — implements `service.Notifier` interface. Formats all 6 message types per spec 05 templates: SessionCreated, SessionsCreatedRecurring (summary), SessionShifted, SessionCanceled, MemberRSVPed, MemberCanceledRSVP. All user content is MarkdownV2-escaped.

### Email notification files ✅
- `internal/service/email_notification.go` — `NotificationEmailSender` and `RSVPMemberLister` interfaces, `RSVPRecipient` model (in `model/rsvp.go`), `SetEmailNotifier()` to configure email deps on `SessionService`, `sendCancelEmails()` and `sendShiftedEmails()` methods (fire-and-forget, query RSVPed members, render templates, send per-member emails). Nil-safe when email not configured.
- `internal/service/email_templates.go` — `cancelEmailSubject/Body()`, `rescheduleEmailSubject/Body()`, `formatDateHuman()`. Templates match spec 05 exactly (plain text, "— Co-Working Space" sign-off). Reschedule email includes old and new date/time with "Previously:" / "Now:" format.
- `internal/service/email.go` — added `SendNotificationEmail()` method to `ResendEmailSender` (reuses private `send()`, logs failures at warn level).
- `internal/repository/rsvp.go` — added `ListEmailsBySession()` (JOINs rsvps + members, returns `[]model.RSVPRecipient` with name + email).
- `internal/service/session.go` — added `emailSender` and `rsvpLister` fields. `Cancel()` fires `go s.sendCancelEmails(session)`. `Update()` fires `go s.sendShiftedEmails(existing, session)` when date/time changes (passes both old and new session for email template).
- `cmd/api/main.go` — calls `sessionSvc.SetEmailNotifier(emailSender, rsvpRepo)` after construction.

### Update wiring ✅
- `cmd/api/main.go` — wires `TelegramService` + `TelegramNotifier` when credentials configured, falls back to `NoopNotifier`. Wires `ResendEmailSender` + `RSVPRepository` as email notifier on `SessionService`. Logs which mode is active at startup.

### Telegram tests ✅
- `internal/telegram/escape_test.go` — 22 tests: all 18 special chars individually, backslash-first ordering, plain text, empty string, compound escaping, real-world formats (time, date, RSVP count)
- `internal/telegram/telegram_test.go` — 7 tests: enabled/disabled (4 combos), skip when disabled, sends correct payload to API (verifies chat_id, text, parse_mode, URL path), handles API error, handles network error
- `internal/telegram/notifier_test.go` — 9 tests: all 6 message types with content verification, no-description variant, empty recurring list, special char escaping in session titles

### Email notification tests ✅
- `internal/service/email_notification_test.go` — 15 tests: formatDateHuman (4 cases), cancelEmailSubject, cancelEmailBody (content verification), rescheduleEmailSubject, rescheduleEmailBody (old/new time verification), sendCancelEmails (sends to all recipients, no recipients, nil dependencies), sendShiftedEmails (sends to all recipients, nil dependencies), integration: cancel triggers email, shift triggers email, title-only no email, cancel without email notifier still works

### Verify
- RSVP → 201, duplicate → 409, full → 409, canceled → 422, past → 422
- Cancel RSVP → 204, non-existent → 404
- Guest list shows names + telegram handles
- Session `rsvp_count` updates correctly
- Telegram messages appear in group (with valid creds), including recurring summary
- Without Telegram creds: no errors, debug log only
- Cancel session with RSVPs → Telegram broadcast + email sent to each RSVPed member
- Reschedule session with RSVPs → Telegram broadcast + email with old/new times

---

## Phase 6: Frontend (React + Vite + Tailwind + React Query)

Complete React SPA. Built in sub-steps. See [08-ui-ux.md](./specs/08-ui-ux.md) for app structure and workflows, [09-design-patterns.md](./specs/09-design-patterns.md) for visual patterns.

### Setup ✅ COMPLETED
- Scaffold with `npm create vite@latest frontend -- --template react-ts`
- Add deps: `react-router-dom`, `@tanstack/react-query`, `tailwindcss`, `@tailwindcss/vite`
- Add dev deps: `vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`, `jsdom`
- `vite.config.ts` — proxy `/api` to `localhost:8080`, Tailwind v4 Vite plugin
- `vitest.config.ts` — jsdom environment, setup file for jest-dom matchers

### Step 6a: Design System + API Client + Auth ✅ COMPLETED

**Status:** Done — all design system components, API client, TypeScript types, auth context, route guards, login/verify/profile pages, app shell with responsive layout, and routing created. 50 frontend unit tests passing.

#### Files created
- `src/index.css` — Tailwind v4 import + custom theme tokens (status colors, primary/destructive)
- `index.html` — inline dark mode init script (prevents flash), `dark:` body classes
- `src/lib/darkMode.ts` — `initTheme()` (localStorage → system pref → apply `dark` class), `toggleTheme()` (toggle + persist), `isDarkMode()`
- `src/components/ThemeToggle.tsx` — sun/moon icon button, calls `toggleTheme()`
- `src/components/Toast.tsx` — toast display, reads from `ToastContext`, dismiss button
- `src/context/ToastContext.tsx` — `ToastProvider`, `useToast()`, `addToast(msg, type)`, auto-dismiss (3s success/info, 5s error), max 3 visible
- `src/components/ConfirmModal.tsx` — reusable modal (title, message, cancel + destructive button, backdrop click dismiss, Escape key)
- `src/components/EmptyState.tsx` — icon slot, heading, subtext, optional CTA button
- `src/components/StatusBadge.tsx` — colored pill (green scheduled, amber rescheduled/shifted, red canceled)
- `src/api/client.ts` — fetch wrapper (`credentials: 'include'`, JSON, `ApiError` class with status/body), `api.get/post/patch/delete`
- `src/types/index.ts` — TypeScript interfaces matching Go models (`Member`, `SpaceSession`, `RSVP`, `RSVPWithMember`, `RSVPMember`, request types, `APIResponse`, `APIError`)
- `src/context/AuthContext.tsx` — `AuthProvider`, `useAuth()` with `user/loading/refresh/logout` state
- `src/components/ProtectedRoute.tsx` — redirect to `/login` if unauthenticated, loading spinner
- `src/components/AdminRoute.tsx` — redirect to `/` if not admin
- `src/pages/LoginPage.tsx` — centered card, email input → POST magic-link → confirmation message, rate limit (429) error handling
- `src/pages/VerifyPage.tsx` — reads token from URL, refreshes auth, redirects to `/`
- `src/pages/ProfilePage.tsx` — inline form (name + telegram_handle), PATCH `/api/auth/profile`, success toast
- `src/components/Layout.tsx` — responsive top navbar (desktop: links; mobile: hamburger → overlay), dark mode toggle, logout
- `src/App.tsx` — QueryClientProvider + AuthProvider + ToastProvider + BrowserRouter, route tree with protected/admin routes
- `src/main.tsx` — StrictMode render entry

#### Tests (50 passing)
- `src/components/__tests__/StatusBadge.test.tsx` — 7 tests: all 3 statuses render correct labels, unknown status fallback, correct color classes per status
- `src/components/__tests__/EmptyState.test.tsx` — 7 tests: heading, subtext, no subtext, CTA button, click handler, no button, icon rendering
- `src/components/__tests__/ConfirmModal.test.tsx` — 8 tests: open/closed rendering, confirm/cancel clicks, Escape key, backdrop click, content click no-dismiss, custom confirm label
- `src/context/__tests__/ToastContext.test.tsx` — 7 tests: add toast, multiple toasts, max 3 limit, dismiss click, auto-dismiss 3s (success), auto-dismiss 5s (error), info type default styling
- `src/lib/__tests__/darkMode.test.ts` — 8 tests: localStorage dark/light, system preference dark/light, toggle to dark/light with persistence, isDarkMode true/false
- `src/api/__tests__/client.test.ts` — 9 tests: GET credentials/content-type, GET JSON parsing, POST/PATCH/DELETE methods, 204 returns undefined, ApiError on non-ok, error status/body/message, non-JSON error fallback
- `src/pages/__tests__/LoginPage.test.tsx` — 4 tests: renders form, shows confirmation after submit, rate limit error, back to form from confirmation

### Step 6b: Sessions + RSVP ✅ COMPLETED

**Status:** Done — backend `user_rsvped` enhancement, SessionsPage with date grouping, SessionCard with 4 RSVP states, SessionDetailPage with guest list, GuestList component, optimistic RSVP updates, cancel RSVP confirmation modal. 20 new frontend tests passing (70 total). 194 Go tests passing.

#### Backend enhancement (user_rsvped)
- `internal/repository/session.go` — `List()` and `GetByID()` now accept `*uuid.UUID` memberID parameter. Added `EXISTS(SELECT 1 FROM rsvps WHERE session_id = s.id AND member_id = $N)` subquery to populate `user_rsvped` boolean. Passes `uuid.Nil` when no member context.
- `internal/service/session.go` — `SessionRepo` interface, `List()`, and `GetByID()` updated to accept `*uuid.UUID`. Internal calls from `Update()`/`Cancel()` pass `nil`.
- `internal/handler/session.go` — `List` and `GetByID` handlers extract member from auth context and pass `memberID` to service.

#### Frontend files created
- `src/components/SessionCard.tsx` — title + status badge (right), time range, capacity ("3/8 spots" or "Full"), RSVP button (4 states: RSVP/Cancel RSVP/Full/hidden for canceled). Admin: edit icon link. Canceled: muted opacity + strikethrough. Cancel RSVP opens ConfirmModal. Exported `formatDateLabel()` helper.
- `src/components/GuestList.tsx` — attendee names + telegram handles (strips leading `@`), ordered by RSVP time, empty state "No RSVPs yet. Be the first!"
- `src/pages/SessionsPage.tsx` — React Query fetch, date grouping with sticky headers on mobile, empty state with calendar icon, optimistic RSVP/cancel mutations with rollback, toast notifications ("You're in!" / "RSVP canceled.").
- `src/pages/SessionDetailPage.tsx` — full session detail (title, description, date, time, capacity, status badge), RSVP button with same 4 states, guest list via separate query, back link to sessions list, cancel RSVP ConfirmModal.
- Updated `src/App.tsx` — replaced SessionsPage placeholder with real import, added `/sessions/:id` route for SessionDetailPage.

#### Tests (20 new, 70 total)
- `src/components/__tests__/SessionCard.test.tsx` — 14 tests: renders title/time, spot count, status badge, RSVP button when available, calls onRSVP, Cancel RSVP button when RSVPed, confirm modal on cancel click, calls onCancelRSVP after confirm, Full badge when full, no buttons when canceled, opacity/line-through for canceled, Cancel RSVP when full+RSVPed, formatDateLabel (2 cases)
- `src/components/__tests__/GuestList.test.tsx` — 6 tests: empty state message, renders member names, telegram handle display, handle without @, null telegram handle, multiple guests

### Step 6c: Admin session management ✅ COMPLETED

**Status:** Done — SessionForm component, create/edit pages, cancel session flow with confirmation modals, admin action buttons on SessionCard and SessionDetailPage. 16 new frontend tests passing (86 total frontend, 194 Go tests).

#### Files created
- `src/components/SessionForm.tsx` — create/edit form (title, description, date, start/end time, capacity). Client-side validation (required fields, end > start, capacity >= 1). **Recurring toggle:** "Repeat weekly" checkbox → "for N weeks" number input (1–12), only on create. Edit mode sends only changed fields (dynamic PATCH). Loading state with "Saving..." button.
- `src/pages/SessionCreatePage.tsx` — admin-only create page, wraps SessionForm, handles recurring response (array vs single), toast notifications ("Session created." / "Created N sessions."), navigates to home on success.
- `src/pages/SessionEditPage.tsx` — admin-only edit page, fetches session data, pre-populates SessionForm, blocks editing canceled sessions, invalidates queries on success, navigates to detail page.

#### Files updated
- `src/components/SessionCard.tsx` — added `onCancelSession` prop, cancel session button (X icon) for admins, cancel session ConfirmModal ("Cancel this session?" with warning about notifications). Edit link now points to `/sessions/:id/edit`.
- `src/pages/SessionsPage.tsx` — "Create Session" button (top-right, admin only), cancel session mutation with toast, passes `onCancelSession` to SessionCard. Admin-specific empty state subtext.
- `src/pages/SessionDetailPage.tsx` — admin action buttons (Edit link + "Cancel Session" red button), cancel session mutation with ConfirmModal, navigates to home on cancel success.
- `src/App.tsx` — added `/sessions/new` and `/sessions/:id/edit` routes under AdminRoute.

#### Tests (16 new, 86 total)
- `src/components/__tests__/SessionForm.test.tsx` — 16 tests: create mode (empty fields, submit button, repeat weekly checkbox, week count input toggle, required field validation, end time validation, valid submit, repeat_weekly submit, loading state), edit mode (pre-populated fields, Save Changes button, no repeat weekly, only changed fields sent, empty object when unchanged, null description), required field indicators

### Step 6d: Admin member management ✅ COMPLETED

**Status:** Done — MembersPage with desktop table / mobile card list, active filter (Active/Inactive/All), MemberForm for create/edit, full CRUD mutations, activate/deactivate toggle, delete with confirmation modal + RSVP guard error handling. 22 new frontend tests passing (108 total frontend, 194 Go tests).

#### Files created
- `src/components/MemberForm.tsx` — create/edit form (email, name, telegram_handle with `@` prefix hint, is_admin checkbox, send_invite checkbox on create only). Client-side validation (email required + format, name required). Edit mode hides email + send_invite, sends only changed fields (dynamic PATCH). Cancel button, loading state.
- `src/pages/MembersPage.tsx` — React Query fetch with `active` filter param, segmented toggle (Active/Inactive/All), desktop table (Name/Email/Telegram/Role/Status/Actions columns), mobile card list, inline create/edit form panel, activate/deactivate toggle button per row, delete with ConfirmModal, empty state with people icon + "Add Member" CTA. Full error→toast mapping (409 for duplicate email and RSVP guard).

#### Files updated
- `src/App.tsx` — replaced Step 6d placeholder with real `MembersPage` import.

#### Tests (22 new, 108 total)
- `src/components/__tests__/MemberForm.test.tsx` — 22 tests: create mode (empty fields, Add Member button, Cancel button, onCancel callback, required validation, email format validation, valid submit, admin+invite checkboxes, empty telegram excluded, loading state, @ prefix hint), edit mode (pre-populated, no email field, no send_invite, Save Changes button, only changed fields, empty when unchanged, is_admin toggle, null telegram, name required), required field indicators (2 create, 1 edit)

### Step 6e: Profile Page ✅ COMPLETED

**Status:** Done — implemented as part of Step 6a. ProfilePage with inline form (name + telegram_handle), PATCH call, success/error toasts, loading state.

- `src/pages/ProfilePage.tsx` — inline form with Name (required) + Telegram Handle (optional). Pre-populated from `GET /api/auth/me`. Save → `PATCH /api/auth/profile` → success toast "Profile updated." Per spec 02 UI design.

### Step 6f: App Shell + Navigation ✅ COMPLETED

**Status:** Done — implemented as part of Step 6a. Layout with responsive navbar, hamburger menu, admin links, dark mode toggle, routing with all protected/admin routes.

- `src/main.tsx` — QueryClientProvider + RouterProvider
- `src/App.tsx` — route definitions: `/login`, `/auth/verify`, `/` (sessions), `/sessions/:id`, `/members` (admin), `/profile`
- `src/components/Layout.tsx` — responsive top navbar. Desktop: logo + "Sessions" + "Members" (admin) + user menu dropdown (Profile, dark mode toggle, Logout). Mobile: hamburger → slide-in overlay. Per spec 08 navigation structure.

### Tests
- Component tests with Vitest + React Testing Library
- Key flows: login form submission, session card RSVP toggle, member form validation, toast display, modal confirm/cancel, dark mode toggle

### Verify
- `npm run dev` at localhost:5173
- Full login flow via magic link
- Sessions page with date grouping, RSVP working, counts updating, optimistic updates
- Recurring session creation (repeat weekly toggle)
- Admin session and member management
- Profile self-edit at `/profile`
- Dark mode toggle persists across page reloads
- Toast notifications for all actions
- Confirmation modals for destructive actions
- Mobile responsive: hamburger nav, card lists, sticky date headers
- Logout works, protected routes redirect

---

## Phase 7: Docker & Deployment ✅ COMPLETED

Containerize everything with Docker Compose.

**Status:** Done — Dockerfile.api (multi-stage Go build with CGO_ENABLED=0, ca-certificates, migrations), Dockerfile.frontend (multi-stage Node build with nginx runtime), nginx.conf (SPA fallback + /api proxy to backend service), docker-compose.yml (3 services with postgres healthcheck dependency). 28 new unit tests passing (222 Go total).

### Files
- `Dockerfile.api` — multi-stage: `golang:1.25-alpine` build → `alpine:3.19` runtime (static binary + migrations + ca-certificates)
- `Dockerfile.frontend` — multi-stage: `node:20-alpine` build → `nginx:alpine` serving dist
- `frontend/nginx.conf` — `try_files $uri /index.html` for SPA routing, `proxy_pass http://api:8080` for `/api/` requests
- `docker-compose.yml` — 3 services: `api` (depends on postgres healthy, env_file), `frontend` (depends on api), `postgres:16-alpine` (pgdata volume, healthcheck)

### Tests
- `internal/deploy/deploy_test.go` — 28 tests: Dockerfile.api (exists, multi-stage, go build, CGO disabled, migrations, port, ca-certs), Dockerfile.frontend (exists, multi-stage, npm ci, build, nginx runtime, dist copy, nginx conf), nginx.conf (exists, SPA fallback, API proxy, port 80), docker-compose.yml (exists, 3 services, dockerfile refs, postgres image/healthcheck, healthy dependency, volume, API port, env_file)

### Verify
- `docker compose up --build` starts full stack
- API at :8080, frontend at :3000, migrations auto-run
- `docker compose exec api /api seed-admin --email admin@test.com --name Admin`
- Full login and RSVP flow works through Docker
- `docker compose down -v` cleans up completely

---

## Dependency Graph

```
Phase 0 (Scaffolding)
  ↓
Phase 1 (Migrations + Models)
  ↓
Phase 2 (Member CRUD)
  ↓
Phase 3 (Auth + Middleware)
  ↓
Phase 4 (Sessions)
  ↓
Phase 5 (RSVP + Telegram)
  ↓
Phase 6 (Frontend)
  ↓
Phase 7 (Docker)
```

## Key Implementation Notes

1. **RSVP atomicity** — `SELECT FOR UPDATE` within a single transaction is critical. Notification sent *after* commit, not inside transaction.
2. **Cookie Secure flag** — derive from `FRONTEND_URL` scheme. `false` for `http://localhost`, `true` otherwise.
3. **Magic link verify flow** — browser navigates directly to `GET /api/auth/verify?token=...`, backend sets cookie + 302 redirects to frontend `/`.
4. **`user_rsvped` enhancement** — ✅ Done in Step 6b. `EXISTS` subquery on `List()` and `GetByID()`, memberID extracted from auth context in handler.
5. **Dynamic PATCH queries** — build SQL UPDATE with only provided fields. Always use parameterized queries.
6. **MarkdownV2 escaping** — escape all special chars in user content before sending to Telegram. Test thoroughly.
7. **Recurring sessions** — single transaction wrapping all INSERTs. `repeat_weekly` capped at 12. Summary Telegram notification (one message, not N). `repeat_weekly` is transient (not stored).
8. **Profile self-edit** — `PATCH /api/auth/profile` accepts only `name` + `telegram_handle`. All other fields are silently ignored. Name cannot be empty (422).
9. **Email notifications** — reuse existing Resend client (`RESEND_API_KEY`, `RESEND_FROM_EMAIL`). Fire-and-forget after transaction commit, same pattern as Telegram. Triggered on session cancel + reschedule, sent to all RSVPed members.
10. **Invitation email** — `send_invite` is transient (not stored on member). Fire-and-forget via Resend after member creation.
11. **Dark mode** — Tailwind `darkMode: 'class'` on `<html>`. Check localStorage first, fall back to `prefers-color-scheme: dark`. Persist toggle choice to localStorage.
