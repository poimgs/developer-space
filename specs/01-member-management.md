# 01 — Member Management

Admins manage the member directory. Every person who can access the co-working space has a Member record. Members are created by admins (no self-registration); authentication is handled separately via magic links (see [02-authentication.md](./02-authentication.md)).

## User Stories

- **As an admin**, I want to add a new member so they can log in and RSVP to sessions.
- **As an admin**, I want to view all members so I can see who has access.
- **As an admin**, I want to edit a member's details (name, email, telegram handle, admin flag) so records stay accurate.
- **As an admin**, I want to deactivate a member so they can no longer log in or RSVP, without losing historical data.
- **As an admin**, I want to reactivate a previously deactivated member.
- **As an admin**, I want to optionally send an invitation email when creating a member so they know they have access.

## Acceptance Criteria

1. Only authenticated admins (`is_admin = true`) can access member management endpoints.
2. Email must be unique across all members (active and inactive). Attempting to create a duplicate returns `409 Conflict`.
3. Deactivating a member sets `is_active = false`. Deactivated members cannot authenticate (magic link verification rejects them).
4. Deleting a member is a hard delete. It should only be allowed if the member has no RSVPs. If the member has RSVPs, the admin should deactivate instead.
5. `telegram_handle` is optional — stored without the leading `@`.
6. Listing members supports filtering by `is_active` (default: `true`).
7. When `send_invite` is `true` on creation, a welcome email is sent to the member with the app URL. The email is fire-and-forget — delivery failures do not block the creation response.
8. Members can edit their own name and telegram handle via `/profile` — see [02-authentication.md](./02-authentication.md#update-profile-self-edit).

## API Endpoints

### Create Member

```
POST /api/members
Authorization: session cookie (admin)

Request:
{
  "email": "dev@example.com",
  "name": "Jane Doe",
  "telegram_handle": "janedoe",   // optional
  "is_admin": false,              // optional, default false
  "send_invite": false            // optional, default false — transient, not stored
}

201 Created:
{
  "data": {
    "id": "uuid",
    "email": "dev@example.com",
    "name": "Jane Doe",
    "telegram_handle": "janedoe",
    "is_admin": false,
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}

409 Conflict:
{ "error": "A member with this email already exists" }

422 Unprocessable Entity:
{ "error": "Validation failed", "details": { "email": "required" } }
```

### List Members

```
GET /api/members?active=true
Authorization: session cookie (admin)

200 OK:
{
  "data": [
    { "id": "uuid", "email": "...", "name": "...", "telegram_handle": "...", "is_admin": false, "is_active": true, "created_at": "...", "updated_at": "..." }
  ]
}
```

Query parameters:
- `active` — `true` (default), `false`, or `all`

### Get Member

```
GET /api/members/:id
Authorization: session cookie (admin)

200 OK:
{ "data": { ...member } }

404 Not Found:
{ "error": "Member not found" }
```

### Update Member

```
PATCH /api/members/:id
Authorization: session cookie (admin)

Request (partial update — only include fields to change):
{
  "name": "Jane Smith",
  "telegram_handle": "janesmith",
  "is_admin": true,
  "is_active": false
}

200 OK:
{ "data": { ...updated member } }

409 Conflict:
{ "error": "A member with this email already exists" }

404 Not Found:
{ "error": "Member not found" }
```

### Delete Member

```
DELETE /api/members/:id
Authorization: session cookie (admin)

204 No Content

409 Conflict:
{ "error": "Cannot delete member with existing RSVPs. Deactivate instead." }

404 Not Found:
{ "error": "Member not found" }
```

## Invitation Email

When `send_invite` is `true` on member creation, a welcome email is sent via Resend.

```
Subject: You've been added to [App Name]

Hi Jane,

You've been added to the co-working space. You can now log in and
RSVP to upcoming sessions.

Log in here: {FRONTEND_URL}/login

— Co-Working Space
```

- Uses the existing Resend integration (`RESEND_API_KEY`, `RESEND_FROM_EMAIL`).
- Fire-and-forget: sent via goroutine after the member is created. Failures logged at `warn` level.
- `send_invite` is transient — it is not stored on the member record.

## UI Design

See [09-design-patterns.md](./09-design-patterns.md) for shared component patterns referenced below.

### Members Page (`/members`)

Admin-only page. Accessible from the "Members" link in the navbar.

- **Desktop:** Table layout with columns: Name, Email, Telegram, Admin (badge), Active (status indicator).
- **Mobile:** Table collapses to card list (per [table pattern](./09-design-patterns.md#table-pattern) from 09). One card per member with key-value pairs stacked vertically and inline actions.

### Active Filter

Toggle above the table — defaults to showing **active members only**. Options: "Active" (default), "Inactive", "All".

### "Add Member" Button

Positioned above the table (right-aligned on desktop, full-width on mobile). Opens the member creation form as a modal or inline panel.

### Member Creation Form

Follows [form patterns](./09-design-patterns.md#form-patterns) from 09.

**Fields:**
- Email (required)
- Name (required)
- Telegram Handle (optional, with `@` prefix hint)
- Is Admin (checkbox, default unchecked)
- "Send invitation email" (checkbox, default unchecked)

### Row Actions

- **Edit** (pencil icon) → opens the member form pre-populated for editing.
- **Activate/Deactivate** toggle → immediately toggles `is_active` via `PATCH`.
- **Delete** (trash icon) → opens confirmation modal per [confirmation modal pattern](./09-design-patterns.md#confirmation-modals).

### Delete Guard

If the member has existing RSVPs, deletion is blocked. The server returns 409, and the UI shows an **error toast**: "Cannot delete member with existing RSVPs. Deactivate instead."

### Empty State

Per [empty state pattern](./09-design-patterns.md#empty-states) from 09.

- **Icon:** People/users icon.
- **Heading:** "No members yet."
- **CTA:** "Add Member" button.

## Implementation Notes

- Email uniqueness enforced at the database level with a unique index.
- `updated_at` is set automatically on every UPDATE via a trigger or application code.
- `telegram_handle` is stored without `@` prefix; the application strips it on input if present.
