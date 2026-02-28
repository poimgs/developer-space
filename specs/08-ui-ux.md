# 08 — UI/UX: App Structure & Workflows

High-level app structure, navigation, page inventory, and all user workflows for the co-working space app. This spec defines **what pages exist and how users move between them**. It does not define component-level design — visual patterns live in [09-design-patterns.md](./09-design-patterns.md) and feature-specific UI details live in each feature spec (01–04).

**Frontend stack:** React + Vite + Tailwind CSS + React Query + React Router

## Page Inventory

| Route | Page | Access | Purpose | Feature Spec |
|-------|------|--------|---------|--------------|
| `/login` | Login | Public | Email input for magic link | [02](./02-authentication.md) |
| `/auth/verify` | Verify | Public | Token verification + redirect | [02](./02-authentication.md) |
| `/` | Sessions (Home) | Auth | Upcoming sessions, RSVP | [03](./03-session-management.md), [04](./04-rsvp.md) |
| `/sessions/:id` | Session Detail | Auth | Full info, guest list, RSVP | [03](./03-session-management.md), [04](./04-rsvp.md) |
| `/members` | Members | Admin | Member directory, CRUD | [01](./01-member-management.md) |
| `/profile` | Profile | Auth | Self-edit name + telegram | [02](./02-authentication.md) |

**Access levels:**
- **Public** — no authentication required.
- **Auth** — any authenticated active member.
- **Admin** — authenticated member with `is_admin = true`.

Unauthenticated users hitting Auth/Admin routes are redirected to `/login`. Non-admin users hitting Admin routes are redirected to `/`.

## Navigation Structure

### Desktop (>= 768px)

Top navbar spanning the full width:

```
┌─────────────────────────────────────────────────────┐
│  [Logo/AppName]   Sessions   Members*      [User ▾] │
└─────────────────────────────────────────────────────┘
```

- **Left:** App name or logo, links to `/`.
- **Center-left:** "Sessions" link (always visible), "Members" link (admin only).
- **Right:** User menu dropdown showing the member's name.

**User menu dropdown items:**
1. **Profile** — links to `/profile`.
2. **Dark mode toggle** — sun/moon icon with label, calls `toggleTheme()` (see [09-design-patterns.md](./09-design-patterns.md#dark-mode)).
3. **Log out** — calls `POST /api/auth/logout`, clears auth state, redirects to `/login`.

### Mobile (< 768px)

Top navbar collapses to:

```
┌─────────────────────────────────┐
│  [Logo/AppName]         [☰]    │
└─────────────────────────────────┘
```

Hamburger icon opens a slide-in overlay from the right:
- Sessions link
- Members link (admin only)
- Profile link
- Dark mode toggle
- Log out button

Overlay closes on navigation or tapping outside.

## User Workflows

### 1. First-Time Admin Setup

**Actor:** Admin (first user)
**Precondition:** App deployed, database migrated, no members exist.

1. Admin runs `seed-admin` CLI command to create their member record.
2. Admin navigates to `/login` and enters their email.
3. Admin receives magic link email, clicks it.
4. Browser hits `/auth/verify?token=...` → backend verifies → sets cookie → redirects to `/`.
5. Sessions home page loads showing an **empty state**: calendar icon, "No upcoming sessions", and a "Create Session" CTA button.
6. Admin clicks "Create Session", fills the form (title, date, time, capacity), submits.
7. First session appears on the home page. Telegram notification is sent to the group.
8. Admin navigates to `/members` and clicks "Add Member" to start building the member directory.

### 2. Member Login

**Actor:** Any member (created by admin)
**Precondition:** Member record exists with an active status.

1. Member navigates to `/login`.
2. Enters their email and clicks "Send login link".
3. Screen shows: "Check your email for a login link."
4. Member opens their email and clicks the magic link.
5. Browser navigates to `/auth/verify?token=...`.
6. Verify page shows a loading spinner during verification.
7. On success: backend sets session cookie and 302 redirects to `/`.
8. Sessions home page loads with upcoming sessions.
9. On failure (expired/used token): error message + "Back to login" link.

### 3. Member RSVP

**Actor:** Authenticated member
**Precondition:** At least one upcoming session exists with available capacity.

1. Member views the sessions home page (`/`). Sessions are grouped under date headers ("Monday, March 9").
2. Each session card shows: title, time range, capacity ("3/8 spots"), and an "RSVP" button.
3. Member clicks "RSVP" on a session card.
4. Button state changes to "Cancel RSVP" (secondary/outline style). Spot count updates optimistically.
5. Success toast: "You're in!" (auto-dismisses after 3 seconds).
6. Telegram notification sent: "[Name] RSVPed to [Session Title]".
7. **To cancel:** Member clicks "Cancel RSVP". Confirmation modal appears: "Cancel your RSVP? You'll lose your spot for '[Title]' on [Date]."
8. Member confirms → RSVP removed, button reverts to "RSVP", spot count updates.
9. Toast: "RSVP canceled." Telegram notification sent.

### 4. Admin Session Management

**Actor:** Admin
**Precondition:** Admin is authenticated.

1. Admin views sessions home page. A "Create Session" button is visible (floating action button on mobile, top-right button on desktop).
2. **Create:** Admin clicks "Create Session" → session form opens with fields: title, description (optional), date, start time, end time, capacity. Admin fills and submits → 201 → session appears in list, success toast "Session created." Telegram notification sent.
3. **Edit:** Admin clicks pencil icon on a session card → session form opens pre-populated. Admin changes fields and saves → 200 → card updates. If date/time changed, status becomes "Rescheduled" and Telegram notification sent. If session has RSVPs and date/time changed, email notifications are also sent to RSVPed members.
4. **Cancel:** Admin clicks X icon → confirmation modal: "Cancel this session? This will cancel '[Title]' on [Date]. All RSVPed members will be notified. This action cannot be undone." Admin confirms → session status becomes "canceled", card shows muted/strikethrough style. Telegram notification + email to RSVPed members sent.

### 5. Recurring Session Creation

**Actor:** Admin
**Precondition:** Admin is creating a new session.

1. On the session creation form, admin checks the "Repeat weekly" checkbox.
2. A number input appears: "for N weeks" (1–12).
3. Admin sets the repeat count (e.g., 4 weeks).
4. On submit, the backend creates N+1 independent sessions in a single transaction (the original + N weekly copies).
5. All sessions appear in the sessions list under their respective date headers.
6. Success toast: "Created 5 sessions."
7. A single summary Telegram message is sent listing all created dates.
8. After creation, each session is independent — editing or canceling one does not affect the others.

### 6. Admin Member Management

**Actor:** Admin
**Precondition:** Admin is authenticated, on `/members` page.

1. Members page shows a table (desktop) or card list (mobile) with columns: Name, Email, Telegram, Admin badge, Active status.
2. **Active filter** toggle above the table — defaults to showing active members only.
3. **Create:** Admin clicks "Add Member" → form opens with fields: email, name, telegram handle, is_admin checkbox, "Send invitation email" checkbox (default unchecked). Admin fills and submits → 201 → member appears in list, success toast "Member added." If invite checkbox was checked, a welcome email is sent with the app URL.
4. **Edit:** Admin clicks pencil icon → form opens pre-populated. Admin updates fields and saves → success toast "Member updated."
5. **Deactivate:** Admin clicks the active status toggle → member becomes inactive. Inactive members cannot log in or RSVP.
6. **Reactivate:** Admin toggles back → member is active again.
7. **Delete:** Admin clicks trash icon → confirmation modal: "Delete this member? This action cannot be undone." If the member has RSVPs, an error toast appears: "Cannot delete member with existing RSVPs. Deactivate instead." If no RSVPs, deletion proceeds → success toast "Member deleted."

### 7. Member Profile Self-Edit

**Actor:** Authenticated member (any role)
**Precondition:** Member is logged in.

1. Member clicks their name in the navbar → user menu dropdown → "Profile".
2. Navigates to `/profile`.
3. Profile page shows an inline form with two editable fields: **Name** (required) and **Telegram Handle** (optional).
4. Fields are pre-populated from `GET /api/auth/me`.
5. Member edits their name or telegram handle and clicks "Save".
6. `PATCH /api/auth/profile` is called with the updated fields.
7. Success toast: "Profile updated."
8. The navbar user menu reflects the updated name.

### 8. Cancel/Reschedule Notification Flow

**Actor:** System (triggered by admin action)
**Precondition:** A session with active RSVPs is canceled or rescheduled.

1. Admin cancels or reschedules a session (via workflows 4 above).
2. Backend commits the session status change.
3. **Telegram broadcast:** A message is posted to the group chat with session details (cancel or reschedule template per [05-notifications.md](./05-telegram-notifications.md)).
4. **Email notifications:** The system queries all members with active RSVPs for the affected session. For each RSVPed member, a notification email is sent via Resend:
   - **Cancel email:** Subject "Session Canceled: [Title]", body includes session details and a note that their RSVP has been removed.
   - **Reschedule email:** Subject "Session Rescheduled: [Title]", body includes old and new date/time and a note that their RSVP is still active.
5. Emails are sent fire-and-forget (same pattern as Telegram — failures logged, never block the admin's response).
6. Admin sees a success toast confirming their action. They do not see individual email delivery status.

## Data Fetching Strategy

- **React Query** manages all server state. Configured with:
  - `refetchOnWindowFocus: true` — manual refresh when user returns to the tab.
  - No polling — data is fetched on mount and on explicit user actions.
  - Stale time: 30 seconds for session lists, 0 for mutations.
- **Optimistic updates** for RSVP actions: UI updates immediately, rolls back on server error.
- **4-week default view:** Sessions page fetches `GET /api/sessions?from=today&to=today+28d` by default.
