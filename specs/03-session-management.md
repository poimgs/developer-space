# 03 — Session Management

Admins create and manage SpaceSessions — scheduled timeslots when the co-working space is open. Members browse upcoming sessions and see availability. Session modifications trigger Telegram notifications (see [05-telegram-notifications.md](./05-telegram-notifications.md)).

## User Stories

- **As an admin**, I want to create a session with a date, time range, title, and capacity so members know when the space is open.
- **As an admin**, I want to edit a session's details (date, time, title, description, capacity) to accommodate changes.
- **As an admin**, I want to cancel a session so members know it's no longer happening.
- **As a member**, I want to view upcoming sessions so I can decide which to attend.
- **As a member**, I want to see how many spots are left in a session so I know if there's room.
- **As an admin**, I want to create a recurring weekly session so I don't have to create each week individually.

## Acceptance Criteria

1. Only admins can create, update, and cancel sessions.
2. All authenticated members (including admins) can list and view sessions.
3. Creating a session sets status to `scheduled`.
4. Editing date or time on a `scheduled` session changes status to `shifted` and triggers a Telegram notification.
5. Editing other fields (title, description, capacity) does **not** change status.
6. Canceling a session sets status to `canceled` and triggers a Telegram notification. Canceled sessions cannot be edited or un-canceled.
7. Capacity cannot be reduced below the current RSVP count.
8. `end_time` must be after `start_time`.
9. Sessions in the past cannot be created or edited (date must be today or later).
10. Member-facing list defaults to upcoming sessions (today onward), sorted by date ascending.

## Status Lifecycle

```
  scheduled ──▶ shifted
      │              │
      │              │ (further date/time edits stay "shifted")
      │              ▼
      │          shifted
      │
      ▼
  canceled (terminal)
```

Both `scheduled` and `shifted` sessions can be canceled. `canceled` is terminal.

## API Endpoints

### Create Session

```
POST /api/sessions
Authorization: session cookie (admin)

Request:
{
  "title": "Friday Afternoon Session",
  "description": "Open co-working, bring your laptop",   // optional
  "date": "2025-02-14",
  "start_time": "14:00",
  "end_time": "18:00",
  "capacity": 8,
  "repeat_weekly": 0              // optional, 0-12, default 0
}

201 Created (single session, repeat_weekly = 0):
{
  "data": {
    "id": "uuid",
    "title": "Friday Afternoon Session",
    "description": "Open co-working, bring your laptop",
    "date": "2025-02-14",
    "start_time": "14:00",
    "end_time": "18:00",
    "capacity": 8,
    "status": "scheduled",
    "rsvp_count": 0,
    "created_by": "uuid",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}

201 Created (recurring, repeat_weekly > 0 — returns array):
{
  "data": [
    { "id": "uuid-1", "date": "2025-02-14", "status": "scheduled", ... },
    { "id": "uuid-2", "date": "2025-02-21", "status": "scheduled", ... },
    { "id": "uuid-3", "date": "2025-02-28", "status": "scheduled", ... }
  ]
}

422 Unprocessable Entity:
{ "error": "Validation failed", "details": { "end_time": "must be after start_time" } }
```

A Telegram notification is sent when a new session is created (see [05-telegram-notifications.md](./05-telegram-notifications.md)).

### List Sessions

```
GET /api/sessions?from=2025-02-01&to=2025-02-28&status=scheduled
Authorization: session cookie

200 OK:
{
  "data": [
    {
      "id": "uuid",
      "title": "...",
      "description": "...",
      "date": "2025-02-14",
      "start_time": "14:00",
      "end_time": "18:00",
      "capacity": 8,
      "status": "scheduled",
      "rsvp_count": 3,
      "created_by": "uuid",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

Query parameters:
- `from` — start date filter (default: today)
- `to` — end date filter (default: today + 28 days)
- `status` — `scheduled`, `shifted`, `canceled`, or `all` (default: shows `scheduled` + `shifted`)

`rsvp_count` is computed (COUNT of RSVPs for the session).

### Get Session

```
GET /api/sessions/:id
Authorization: session cookie

200 OK:
{ "data": { ...session with rsvp_count } }

404 Not Found:
{ "error": "Session not found" }
```

### Update Session

```
PATCH /api/sessions/:id
Authorization: session cookie (admin)

Request (partial — only changed fields):
{
  "date": "2025-02-15",
  "start_time": "15:00"
}

200 OK:
{ "data": { ...updated session } }

409 Conflict:
{ "error": "Cannot reduce capacity below current RSVP count (currently 5)" }

422 Unprocessable Entity:
{ "error": "Cannot edit a canceled session" }

404 Not Found:
{ "error": "Session not found" }
```

**Backend logic:**
1. Reject if session is `canceled`.
2. If `date` or `start_time` or `end_time` changed → set `status = 'shifted'`, send Telegram notification.
3. If `capacity` is being reduced → verify `new_capacity >= current RSVP count`.
4. Apply updates and return.

### Cancel Session

```
DELETE /api/sessions/:id
Authorization: session cookie (admin)

200 OK:
{ "data": { ...session with status: "canceled" } }

422 Unprocessable Entity:
{ "error": "Session is already canceled" }

404 Not Found:
{ "error": "Session not found" }
```

This is a soft delete — sets `status = 'canceled'`. The session record is preserved. A Telegram notification is sent.

## Recurring Sessions

When `repeat_weekly > 0` on creation, the backend creates **N+1 independent sessions** (the original + N weekly copies) in a single database transaction.

### Behavior

- `repeat_weekly` accepts values 0–12. Values outside this range return 422.
- `repeat_weekly` is a **transient field** — it is not stored on the session record. It only affects creation.
- Each created session has the same title, description, start_time, end_time, capacity, and created_by. Only the `date` advances by 7 days per copy.
- All sessions are fully independent after creation. Editing or canceling one does not affect the others.
- The response is an array of all created sessions when `repeat_weekly > 0`.

### Notifications

- A single summary Telegram message is sent listing all created dates (see [05-notifications.md](./05-telegram-notifications.md)).
- Individual "New Session Created" messages are **not** sent for each copy.

### Example

Creating with `date: "2025-03-07"` and `repeat_weekly: 3` produces 4 sessions:
- 2025-03-07, 2025-03-14, 2025-03-21, 2025-03-28

## UI Design

See [09-design-patterns.md](./09-design-patterns.md) for shared component patterns referenced below.

### Sessions Home Page (`/`)

The sessions list is the app's home page — there is no separate dashboard.

- Sessions are listed under **date group headers** formatted as "Day, Month Date" (e.g., "Monday, March 9").
- **Default view:** today to today + 28 days (4-week window).
- Sorted by date ascending, then start time ascending within each date.
- Admin users see a **"Create Session"** button: floating action button on mobile, top-right button on desktop.
- **Mobile:** cards are full-width stacked. Date headers become sticky on scroll.

### Session Card

Each session is rendered as a card (per card pattern from [09](./09-design-patterns.md#card-pattern)):

- **Title** (left) + **status badge** (right of title, per [status badge pattern](./09-design-patterns.md#status-badges-colored-pills)).
- **Below title:** date + time range line (e.g., "14:00 – 18:00").
- **Capacity display:** "3/8 spots" (or "Full" at capacity).
- **RSVP action button:** see [04-rsvp.md](./04-rsvp.md) for button states.
- **Admin-only actions:** edit (pencil icon) and cancel (X icon) buttons.
- **Canceled cards:** muted opacity (`opacity-50`), strikethrough title, no action buttons.

### Session Detail Page (`/sessions/:id`)

Full session information:
- Title, description, date, time range, capacity, status badge.
- RSVP button (per [04-rsvp.md](./04-rsvp.md)).
- Guest list below (per [04-rsvp.md](./04-rsvp.md#guest-list)).
- **Admin actions:** "Edit" opens the session form pre-populated. "Cancel Session" opens the cancel confirmation modal.

### Session Form (Create / Edit)

Follows [form patterns](./09-design-patterns.md#form-patterns) from 09.

**Fields:**
- Title (required, text input)
- Description (optional, textarea)
- Date (required, date picker)
- Start time (required, time input)
- End time (required, time input)
- Capacity (required, number input, min 1)

**Recurring toggle (create only):**
- "Repeat weekly" checkbox. When checked, reveals a "for N weeks" number input (1–12).
- Not visible on edit — recurring is a creation-time concept only.

### Cancel Session Modal

Per [confirmation modal pattern](./09-design-patterns.md#confirmation-modals) from 09.

- **Title:** "Cancel this session?"
- **Message:** "This will cancel '[Title]' on [Date]. All RSVPed members will be notified. This action cannot be undone."
- **Buttons:** "Keep Session" (secondary) / "Cancel Session" (destructive red).

### Empty State

Per [empty state pattern](./09-design-patterns.md#empty-states) from 09.

- **Icon:** Calendar icon.
- **Heading:** "No upcoming sessions."
- **Admin CTA:** "Create Session" button.
- **Non-admin subtext:** "Check back later."

## Implementation Notes

- `rsvp_count` is not stored on the session — it's computed via `COUNT(*)` from the `rsvps` table (joined or subqueried in the list/get endpoints).
- Consider a database CHECK constraint: `end_time > start_time`.
- Status is an application-level enum stored as varchar. Validate on write.
- Recurring session creation should use a single transaction wrapping all INSERTs.
- `repeat_weekly` is validated (0–12) in the service layer before any database operations.
