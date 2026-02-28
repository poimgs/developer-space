# 02 — Authentication

Authentication uses passwordless magic links sent via email. Members click a link to authenticate; the server issues a session cookie. There are no passwords, OAuth flows, or registration pages.

See [01-member-management.md](./01-member-management.md) for how members are created (admin-only).

## User Stories

- **As a member**, I want to enter my email and receive a login link so I can access the app without a password.
- **As a member**, I want to stay logged in across browser sessions so I don't have to re-authenticate frequently.
- **As a member**, I want to log out explicitly when I choose.
- **As a member**, I want to see my own profile information once logged in.
- **As a member**, I want to edit my name and Telegram handle so I can keep my profile accurate without asking an admin.

## Acceptance Criteria

1. A member requests a magic link by submitting their email address.
2. If the email belongs to an active member, the server sends a magic link via Resend. If not, the server responds with the same `200 OK` (no user enumeration).
3. The magic link contains a single-use token valid for **15 minutes**.
4. Tokens are stored as SHA-256 hashes — the raw token only exists in the email link.
5. Clicking the link verifies the token. On success, the server sets an `HttpOnly`, `Secure`, `SameSite=Lax` session cookie.
6. Session duration: **30 days** from creation. No sliding expiration in v1.
7. Inactive members (`is_active = false`) are rejected at verification with a generic error.
8. Used or expired tokens are rejected.
9. Rate limiting: max **5 magic link requests per email per hour**.

## Magic Link Flow

```
Member                  Frontend               Backend               Resend
  │                        │                      │                     │
  │  Enter email           │                      │                     │
  │───────────────────────▶│                      │                     │
  │                        │  POST /auth/magic-link                     │
  │                        │─────────────────────▶│                     │
  │                        │                      │  Send email         │
  │                        │                      │────────────────────▶│
  │                        │  200 OK              │                     │
  │                        │◀─────────────────────│                     │
  │  "Check your email"    │                      │                     │
  │◀───────────────────────│                      │                     │
  │                        │                      │                     │
  │  Click link in email   │                      │                     │
  │───────────────────────▶│                      │                     │
  │                        │  GET /auth/verify?token=abc123             │
  │                        │─────────────────────▶│                     │
  │                        │                      │  Validate token     │
  │                        │                      │  Create session     │
  │                        │  Set-Cookie + redirect to /                │
  │                        │◀─────────────────────│                     │
  │  Logged in             │                      │                     │
  │◀───────────────────────│                      │                     │
```

## API Endpoints

### Request Magic Link

```
POST /api/auth/magic-link

Request:
{ "email": "dev@example.com" }

200 OK (always — prevents user enumeration):
{ "message": "If this email is registered, a login link has been sent." }

429 Too Many Requests:
{ "error": "Too many requests. Try again later." }
```

**Backend logic:**
1. Look up active member by email.
2. If not found or inactive → return 200 (do nothing).
3. Generate a cryptographically random token (32 bytes, base64url-encoded).
4. Store `SHA-256(token)` in `magic_tokens` table with `expires_at = now() + 15 min`.
5. Send email via Resend with link: `{FRONTEND_URL}/auth/verify?token={raw_token}`.

### Verify Token

```
GET /api/auth/verify?token={raw_token}

302 Redirect to / (on success, with Set-Cookie header)

401 Unauthorized (invalid/expired/used token):
{ "error": "Invalid or expired login link." }
```

**Backend logic:**
1. Hash the provided token with SHA-256.
2. Look up the hash in `magic_tokens` where `used_at IS NULL` and `expires_at > now()`.
3. If not found → 401.
4. Mark token as used (`used_at = now()`).
5. Look up the member; verify `is_active = true`.
6. Create a session record or sign a session cookie.
7. Set cookie and redirect to `/`.

### Get Current User

```
GET /api/auth/me
Authorization: session cookie

200 OK:
{
  "data": {
    "id": "uuid",
    "email": "dev@example.com",
    "name": "Jane Doe",
    "telegram_handle": "janedoe",
    "is_admin": false
  }
}

401 Unauthorized:
{ "error": "Not authenticated" }
```

### Logout

```
POST /api/auth/logout
Authorization: session cookie

200 OK (clears cookie):
{ "message": "Logged out" }
```

### Update Profile (Self-Edit)

```
PATCH /api/auth/profile
Authorization: session cookie

Request (partial update — only include fields to change):
{
  "name": "Jane Smith",
  "telegram_handle": "janesmith"
}

200 OK:
{
  "data": {
    "id": "uuid",
    "email": "dev@example.com",
    "name": "Jane Smith",
    "telegram_handle": "janesmith",
    "is_admin": false
  }
}

422 Unprocessable Entity:
{ "error": "Validation failed", "details": { "name": "required" } }
```

**Backend logic:**
1. Only `name` and `telegram_handle` are accepted. All other fields (email, is_admin, is_active) are ignored.
2. `name` cannot be empty — return 422 if an empty string is provided.
3. `telegram_handle` strips leading `@` on input (same as admin member update).
4. Returns the updated profile in the same format as `GET /api/auth/me`.

### Route Structure

```
POST /api/auth/magic-link     [rate-limited, public]
GET  /api/auth/verify          [public]
POST /api/auth/logout          [auth required]
GET  /api/auth/me              [auth required]
PATCH /api/auth/profile        [auth required]
```

## UI Design

See [09-design-patterns.md](./09-design-patterns.md) for shared component patterns referenced below.

### Login Page (`/login`)

Centered card layout with app title at the top.

- Single email input field + "Send login link" button.
- Follows [form patterns](./09-design-patterns.md#form-patterns) from 09.
- After submission: success message "Check your email for a login link." replaces the form.
- Error state: only for rate limiting (429) — "Too many requests. Try again later."
- No registration link — members are created by admins.

### Verify Page (`/auth/verify`)

- Shows a **loading spinner** during token verification.
- On success: backend 302 redirect + cookie set. The browser follows the redirect to `/`.
- On failure (expired/used/invalid token): error message "Invalid or expired login link." + "Back to login" link.

### Profile Page (`/profile`)

Simple inline form — not a modal.

- **Fields:** Name (required), Telegram Handle (optional, with `@` prefix hint).
- Pre-populated from `GET /api/auth/me` on page load.
- "Save" button → `PATCH /api/auth/profile` → success toast "Profile updated."
- Follows [form patterns](./09-design-patterns.md#form-patterns) from 09.
- Members can only edit their own name and telegram handle. Email and admin status are read-only (shown as plain text, not editable).

### User Menu (Navbar)

Dropdown triggered by clicking the member's name in the top navbar.

- **Items:**
  1. "Profile" — navigates to `/profile`.
  2. Dark mode toggle — sun/moon icon + label (see [09-design-patterns.md](./09-design-patterns.md#dark-mode)).
  3. "Log out" — calls `POST /api/auth/logout`, clears auth state, redirects to `/login`.

## UI Notes

- `/login` page: single email input + "Send login link" button.
- After submission: "Check your email for a login link" message.
- `/auth/verify?token=...` page: calls the verify endpoint, shows a spinner, then redirects or shows an error.
- Authenticated users see a user menu with their name and a "Log out" action.

## Security Notes

- **No user enumeration:** The magic link request endpoint always returns 200.
- **Token hashing:** Raw tokens are never stored; only SHA-256 hashes are persisted.
- **Single use:** Tokens are invalidated immediately on verification.
- **Short-lived:** 15-minute expiry limits the window for token interception.
- **Rate limiting:** 5 requests per email per hour prevents abuse. See [06-cross-cutting-concerns.md](./06-cross-cutting-concerns.md).
- **Cookie flags:** `HttpOnly` (no JS access), `Secure` (HTTPS only), `SameSite=Lax` (CSRF mitigation).
- **Inactive members:** Blocked at the verification step, not at the request step.

## Implementation Notes

- Use `crypto/rand` for token generation.
- Use `crypto/sha256` for hashing.
- Session management can be implemented with a signed cookie (e.g., `securecookie`) or a server-side session table. A signed cookie is simpler for v1.
- Clean up expired/used tokens periodically (background goroutine or cron).
