# 06 — Cross-Cutting Concerns

Conventions, middleware, and infrastructure that apply across all features.

## Authentication Middleware

Every `/api/*` route (except auth endpoints) requires a valid session cookie. The middleware:

1. Reads the session cookie.
2. Validates the session (signature check or server-side lookup).
3. Loads the `Member` record and attaches it to the request context.
4. Rejects with `401 Unauthorized` if the session is invalid or the member is inactive.

**Unprotected routes:**
- `POST /api/auth/magic-link`
- `GET /api/auth/verify`

**All other routes** require authentication.

## Admin Authorization Middleware

A subset of endpoints require the `is_admin` flag. Apply as a second middleware layer after authentication.

- Returns `403 Forbidden` if `is_admin = false`.
- Applied to:
  - `GET/POST/PATCH/DELETE /api/members` and `GET /api/members/:id`
  - `POST/PATCH/DELETE /api/sessions`

**Non-admin authenticated routes** (any active member):
- `GET /api/sessions` (list sessions)
- `GET /api/sessions/:id` (get session)
- `POST/DELETE /api/sessions/:id/rsvp`
- `GET /api/sessions/:id/rsvps`
- `GET /api/auth/me`
- `PATCH /api/auth/profile`
- `POST /api/auth/logout`

## API Conventions

### Response Envelope

All successful responses use a `data` wrapper:

```json
{ "data": { ... } }       // single resource
{ "data": [ ... ] }        // list of resources
```

Error responses use an `error` field:

```json
{ "error": "Human-readable message" }
{ "error": "Validation failed", "details": { "field": "reason" } }
```

### HTTP Status Codes

| Code | Usage                                          |
|------|------------------------------------------------|
| 200  | Successful read or update                      |
| 201  | Resource created                               |
| 204  | Successful hard delete (no body)               |
| 400  | Malformed request body (bad JSON)              |
| 401  | Not authenticated                              |
| 403  | Authenticated but not authorized (not admin)   |
| 404  | Resource not found                             |
| 409  | Conflict (duplicate, capacity, has RSVPs)      |
| 422  | Validation error (well-formed but invalid data)|
| 429  | Rate limit exceeded                            |
| 500  | Internal server error                          |

### IDs

All entity IDs are UUIDs (v4), generated server-side. Represented as strings in JSON.

### Timestamps

All timestamps are ISO 8601 format in UTC: `2025-01-01T00:00:00Z`. Stored as `timestamptz` in PostgreSQL.

### Dates and Times

- Dates: `YYYY-MM-DD` (e.g., `2025-02-14`)
- Times: `HH:MM` in 24-hour format (e.g., `14:00`)

## Error Handling

- All errors return JSON, never HTML.
- `500 Internal Server Error` responses log the full error server-side but return a generic message to the client: `{ "error": "Internal server error" }`.
- Panics are recovered by middleware and converted to 500 responses.

## Rate Limiting

Applied selectively in v1:

| Endpoint                    | Limit                     |
|-----------------------------|---------------------------|
| `POST /api/auth/magic-link` | 5 requests per email per hour |

Implementation: in-memory rate limiter (e.g., `golang.org/x/time/rate` or a simple map with TTL). No need for Redis in v1.

## CORS Configuration

For local development, the frontend (Vite dev server, typically `localhost:5173`) and the backend (`localhost:8080`) run on different ports.

```
Access-Control-Allow-Origin: {FRONTEND_URL}
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
Access-Control-Allow-Credentials: true
```

`FRONTEND_URL` is set via environment variable. In production, the frontend is served by the same origin or a known domain.

## Structured Logging

Use structured JSON logging (e.g., `log/slog` from the Go standard library).

Log fields per request:
- `method`, `path`, `status`, `duration`, `request_id`
- `member_id` (if authenticated)

Log levels:
- `info` — request/response log, startup
- `warn` — Telegram notification failures, rate limit hits
- `error` — unexpected errors, database failures
- `debug` — token generation, detailed flow (disabled in production)

## Environment Variables

| Variable           | Description                        | Default            |
|--------------------|------------------------------------|--------------------|
| `PORT`             | API server port                    | `8080`             |
| `DATABASE_URL`     | PostgreSQL connection string       | (required)         |
| `FRONTEND_URL`     | Frontend origin for CORS & links   | `http://localhost:5173` |
| `RESEND_API_KEY`   | Resend API key for email delivery  | (required)         |
| `RESEND_FROM_EMAIL`| "From" address for magic links     | (required)         |
| `SESSION_SECRET`   | Secret for signing session cookies | (required)         |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot API token           | (optional)         |
| `TELEGRAM_CHAT_ID`   | Target Telegram group chat ID    | (optional)         |
| `POSTGRES_PASSWORD` | PostgreSQL password (Docker)      | (required)         |
| `FRONTEND_PORT`    | Frontend host port (Docker)        | `3000`             |
| `LOG_LEVEL`        | Logging level                      | `info`             |

## Request ID

Generate a UUID for each incoming request and include it in:
- The response header: `X-Request-Id`
- All log entries for that request
- Error responses (helps with debugging)

## Content-Type

- All API requests and responses use `Content-Type: application/json`.
- Middleware should reject non-JSON request bodies with `415 Unsupported Media Type` on POST/PATCH endpoints.
