# 05 — Notifications (Telegram + Email)

All session and RSVP activity is broadcast to a single Telegram group chat via a bot. The bot is **output-only in v1** — it posts messages but does not accept or process commands.

Additionally, **email notifications** are sent to RSVPed members when a session is canceled or rescheduled.

## User Stories

- **As a member**, I want to receive Telegram notifications when a new session is created so I can plan ahead.
- **As a member**, I want to be notified when a session is rescheduled or canceled so I can adjust my plans.
- **As a member**, I want to see when someone RSVPs or cancels so I know who's coming.

## Acceptance Criteria

1. The bot posts to a single Telegram group chat configured via environment variables.
2. Notifications are **fire-and-forget**: failures are logged but never block the triggering operation (RSVP, session create, etc.).
3. Notification is sent asynchronously — after the database transaction commits.
4. If Telegram credentials are not configured (`TELEGRAM_BOT_TOKEN` or `TELEGRAM_CHAT_ID` is empty), notifications are silently skipped (logged at debug level).
5. Messages use Telegram's MarkdownV2 formatting.

## Notification Events & Messages

### New Session Created

Triggered by: `POST /api/sessions` (see [03-session-management.md](./03-session-management.md))

```
📅 New Session

*Friday Afternoon Session*
📆 2025-02-14
🕐 14:00 – 18:00
👥 8 spots available

Open co-working, bring your laptop
```

### Session Rescheduled (Shifted)

Triggered by: `PATCH /api/sessions/:id` when date or time changes

```
🔄 Session Rescheduled

*Friday Afternoon Session*
📆 2025-02-14 → 2025-02-15
🕐 14:00–18:00 → 15:00–19:00
```

Only changed fields (date and/or time) are shown with the arrow notation.

### Session Canceled

Triggered by: `DELETE /api/sessions/:id`

```
❌ Session Canceled

*Friday Afternoon Session*
📆 2025-02-14, 14:00 – 18:00
```

### Member RSVPed

Triggered by: `POST /api/sessions/:id/rsvp` (see [04-rsvp.md](./04-rsvp.md))

```
✅ Jane Doe RSVPed

*Friday Afternoon Session*
📆 2025-02-14, 14:00 – 18:00
👥 4 / 8 spots taken
```

### RSVP Canceled

Triggered by: `DELETE /api/sessions/:id/rsvp`

```
🚫 Jane Doe canceled RSVP

*Friday Afternoon Session*
📆 2025-02-14, 14:00 – 18:00
👥 3 / 8 spots taken
```

### Recurring Sessions Created

Triggered by: `POST /api/sessions` with `repeat_weekly > 0` (see [03-session-management.md](./03-session-management.md#recurring-sessions))

A single summary message replaces individual "New Session Created" messages for the batch:

```
📅 Recurring Sessions Created

*Friday Afternoon Session*
🕐 14:00 – 18:00
👥 8 spots each

📆 2025-02-14
📆 2025-02-21
📆 2025-02-28
📆 2025-03-07
```

## Configuration

| Environment Variable  | Description                            | Required |
|-----------------------|----------------------------------------|----------|
| `TELEGRAM_BOT_TOKEN`  | Bot API token from @BotFather          | Yes*     |
| `TELEGRAM_CHAT_ID`    | Numeric ID of the target group chat    | Yes*     |

*If either is empty, notifications are disabled (no error).

## API Integration

The Telegram bot uses the [Bot API `sendMessage` method](https://core.telegram.org/bots/api#sendmessage):

```
POST https://api.telegram.org/bot{token}/sendMessage
Content-Type: application/json

{
  "chat_id": "{TELEGRAM_CHAT_ID}",
  "text": "...",
  "parse_mode": "MarkdownV2"
}
```

## Email Notifications

In addition to Telegram group messages, email notifications are sent directly to affected members when a session is canceled or rescheduled.

### Triggers

| Event | Recipients | Email Sent |
|-------|-----------|------------|
| Session canceled | All members with active RSVP for that session | Cancel notification |
| Session rescheduled (shifted) | All members with active RSVP for that session | Reschedule notification |

### Email Templates

**Session Canceled:**

```
Subject: Session Canceled: Friday Afternoon Session

Hi Jane,

The session "Friday Afternoon Session" scheduled for February 14, 2025
(14:00 – 18:00) has been canceled.

Your RSVP has been noted. No further action is needed.

— Co-Working Space
```

**Session Rescheduled:**

```
Subject: Session Rescheduled: Friday Afternoon Session

Hi Jane,

The session "Friday Afternoon Session" has been rescheduled:

  Previously: February 14, 2025, 14:00 – 18:00
  Now:        February 15, 2025, 15:00 – 19:00

Your RSVP is still active — no action needed unless the new time doesn't work for you.

— Co-Working Space
```

### Implementation

- Uses the existing **Resend** integration (same `RESEND_API_KEY` and `RESEND_FROM_EMAIL` — no new env vars).
- Emails are sent **fire-and-forget after commit**, same pattern as Telegram notifications.
- Sent via a goroutine — email delivery failures are logged at `warn` level but never block the HTTP response.
- The system queries RSVPed members for the session and sends one email per member.
- Extend the `Notifier` interface with email-capable methods, or add a separate `EmailNotifier` service. The `SessionCanceled` and `SessionShifted` flows trigger both Telegram and email.

## Implementation Notes

- Create a `telegram` package/service with a `Notify(ctx, message)` method.
- The service is initialized with the bot token and chat ID from environment variables.
- If credentials are missing, the service operates as a no-op (returns immediately).
- Use a goroutine for sending so it doesn't block the HTTP response. The caller does not wait for the result.
- Log send failures at `warn` level — include the event type and session ID for debugging.
- Log skipped notifications (missing credentials) at `debug` level.
- MarkdownV2 requires escaping special characters (`_`, `*`, `[`, `]`, `(`, `)`, `~`, `` ` ``, `>`, `#`, `+`, `-`, `=`, `|`, `{`, `}`, `.`, `!`). Build a helper function for this.
- No retry logic in v1 — a single attempt per notification is sufficient.
- The `Notifier` interface should include a `SessionsCreatedRecurring` method (or similar) for the summary Telegram message.
- Reuse the existing Resend client for email notifications. Extend the `Notifier` interface or add a dedicated `EmailNotifier` to handle cancel/reschedule emails.

## Future Enhancements (Out of Scope for v1)

- Incoming bot commands (e.g., `/sessions` to list upcoming sessions from Telegram).
- Per-member notification preferences (opt out of certain event types).
- Direct messages instead of group chat posts.
