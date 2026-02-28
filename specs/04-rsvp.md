# 04 — RSVP

Members RSVP to sessions to reserve a spot. Capacity is enforced atomically. Every RSVP and cancellation triggers a Telegram notification (see [05-telegram-notifications.md](./05-telegram-notifications.md)).

## User Stories

- **As a member**, I want to RSVP to an upcoming session so I have a guaranteed spot.
- **As a member**, I want to cancel my RSVP so the spot is freed for others.
- **As a member**, I want to see who else is attending a session so I know who I'll be co-working with.

## Acceptance Criteria

1. Any authenticated active member can RSVP to a session.
2. A member can only RSVP once per session. Duplicate attempts return `409 Conflict`.
3. RSVP is rejected if the session is at capacity → `409 Conflict` with a clear message.
4. RSVP is rejected if the session is `canceled`.
5. Canceling an RSVP frees one spot.
6. The guest list (who RSVPed) is visible to all authenticated members.
7. A member cannot RSVP to a past session.
8. Each RSVP triggers a Telegram notification: "[Name] RSVPed to [Session Title]".
9. Each RSVP cancellation triggers a Telegram notification: "[Name] canceled RSVP for [Session Title]".

## API Endpoints

### RSVP to Session

```
POST /api/sessions/:id/rsvp
Authorization: session cookie

Request: (empty body — member is derived from session)

201 Created:
{
  "data": {
    "id": "uuid",
    "session_id": "uuid",
    "member_id": "uuid",
    "created_at": "2025-01-01T00:00:00Z"
  }
}

409 Conflict (already RSVPed):
{ "error": "You have already RSVPed to this session" }

409 Conflict (at capacity):
{ "error": "This session is full" }

422 Unprocessable Entity (canceled session):
{ "error": "Cannot RSVP to a canceled session" }

422 Unprocessable Entity (past session):
{ "error": "Cannot RSVP to a past session" }

404 Not Found:
{ "error": "Session not found" }
```

**Backend logic (atomic capacity check):**
1. Begin transaction.
2. SELECT session with `FOR UPDATE` (row lock).
3. Verify session is not canceled and not in the past.
4. Count existing RSVPs for this session.
5. If count >= capacity → 409.
6. Check for existing RSVP by this member → 409 if exists.
7. Insert RSVP.
8. Commit transaction.
9. Send Telegram notification (fire-and-forget, outside transaction).

### Cancel RSVP

```
DELETE /api/sessions/:id/rsvp
Authorization: session cookie

204 No Content

404 Not Found (no RSVP to cancel):
{ "error": "You have not RSVPed to this session" }

422 Unprocessable Entity (past session):
{ "error": "Cannot modify RSVP for a past session" }
```

### List RSVPs (Guest List)

```
GET /api/sessions/:id/rsvps
Authorization: session cookie

200 OK:
{
  "data": [
    {
      "id": "uuid",
      "member": {
        "id": "uuid",
        "name": "Jane Doe",
        "telegram_handle": "janedoe"
      },
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}

404 Not Found:
{ "error": "Session not found" }
```

Returns member name and telegram handle (not email) for the guest list. Sorted by `created_at` ascending (first to RSVP listed first).

## UI Design

See [09-design-patterns.md](./09-design-patterns.md) for shared component patterns referenced below.

### RSVP Button States (Session Card)

The RSVP button on each session card changes based on context:

| State | Button | Style | Action |
|-------|--------|-------|--------|
| Available | "RSVP" | Primary indigo ([button styles](./09-design-patterns.md#button-styles)) | `POST /api/sessions/:id/rsvp` |
| Already RSVPed | "Cancel RSVP" | Secondary/outline ([button styles](./09-design-patterns.md#button-styles)) | Opens cancel RSVP modal |
| Full (not RSVPed) | "Full" | Disabled gray | No action |
| Canceled session | *(no button)* | — | — |
| Past session | *(no button)* | — | — |

### RSVP Action Feedback

- **On RSVP success:** Toast "You're in!" (success type, 3s auto-dismiss). Spots count updates optimistically before server response.
- **On cancel RSVP success:** Toast "RSVP canceled." (info type, 3s auto-dismiss). Spots count updates optimistically.
- **On error:** Toast with error message (error type, 5s auto-dismiss). Optimistic update rolls back.

### Cancel RSVP Confirmation Modal

Per [confirmation modal pattern](./09-design-patterns.md#confirmation-modals) from 09.

- **Title:** "Cancel your RSVP?"
- **Message:** "You'll lose your spot for '[Title]' on [Date]."
- **Buttons:** "Keep RSVP" (secondary) / "Cancel RSVP" (destructive red).

### Capacity Display

Shown on both the session card and the session detail page.

- **Format:** "3 / 8 spots" when spots remain.
- **Full:** "Full" text when `rsvp_count >= capacity`.
- Text color: `text-secondary` normally, `text-red-600` when full.

### Guest List (Session Detail Page)

Displayed below session info on `/sessions/:id`.

- Shows attendee **name** and **telegram handle** (not email — privacy).
- Ordered by RSVP time (`created_at` ascending) — first to RSVP listed first.
- **Empty state:** "No RSVPs yet. Be the first!" (inline text, no illustration needed).

## Implementation Notes

- The unique constraint `(session_id, member_id)` on the `rsvps` table prevents duplicates at the database level.
- Use `SELECT ... FOR UPDATE` within a transaction for the capacity check to prevent race conditions.
- The Telegram notification is sent **after** the transaction commits, not inside it. Notification failures must not roll back the RSVP.
- The guest list endpoint joins `rsvps` with `members` to return member names.
