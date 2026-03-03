# Implementation Plan: UI/UX Overhaul

## Context

The Developer Space app is fully functional (Phases 0–7 completed: Go API, React SPA, PostgreSQL, Docker, Telegram notifications). This plan covers the UI/UX overhaul: enhanced profiles, session images, warm visual palette, and redesigned sessions view.

**Specs updated:** `00-overview.md`, `02-authentication.md`, `03-session-management.md`, `08-ui-ux.md`, `09-design-patterns.md`, `index.md`

**Existing codebase:** 194 Go tests, 108 frontend tests, 6 migrations, fully deployed via Docker.

---

## Overhaul Phase 1: Database Migrations + Backend Models ✅ COMPLETED

Add new fields to `members` and `space_sessions` tables, update Go model structs and repository queries.

**Completed:** Migration SQL files, model struct updates (Member: bio, skills, linkedin_url, instagram_handle, github_username; SpaceSession: image_url, location), all repository queries updated (member, session, series, rsvp), service layer updated to pass location through recurring session creation. 9 new unit tests added. All 203 Go tests pass (except pre-existing deploy test), all 108 frontend tests pass.

### Migration: `000007_add_member_profile_fields`

**Up:**
```sql
ALTER TABLE members
  ADD COLUMN bio TEXT,
  ADD COLUMN skills TEXT[] DEFAULT '{}',
  ADD COLUMN linkedin_url VARCHAR(255),
  ADD COLUMN instagram_handle VARCHAR(255),
  ADD COLUMN github_username VARCHAR(255);
```

**Down:**
```sql
ALTER TABLE members
  DROP COLUMN IF EXISTS bio,
  DROP COLUMN IF EXISTS skills,
  DROP COLUMN IF EXISTS linkedin_url,
  DROP COLUMN IF EXISTS instagram_handle,
  DROP COLUMN IF EXISTS github_username;
```

### Migration: `000008_add_session_image_location`

**Up:**
```sql
ALTER TABLE space_sessions
  ADD COLUMN image_url VARCHAR(512),
  ADD COLUMN location TEXT;
```

**Down:**
```sql
ALTER TABLE space_sessions
  DROP COLUMN IF EXISTS image_url,
  DROP COLUMN IF EXISTS location;
```

### Files
- `migrations/000007_add_member_profile_fields.up.sql`
- `migrations/000007_add_member_profile_fields.down.sql`
- `migrations/000008_add_session_image_location.up.sql`
- `migrations/000008_add_session_image_location.down.sql`

### Verify
- `go run ./cmd/api migrate up` adds columns
- `go run ./cmd/api migrate down 2` removes them cleanly
- Existing data is unaffected (all new columns are nullable/have defaults)

---

## Overhaul Phase 2: Visual Refresh — Color Swap ✅ COMPLETED

**Completed:** Replaced indigo accent with amber/gold and cold grays with warm stone neutrals across all frontend files. Updated CSS theme tokens (`--color-primary: #d97706`, `--color-primary-hover: #b45309`, `--color-primary-light: #fffbeb`). Replaced `indigo-*` → `amber-*` across 17 files and `gray-*` → `stone-*` across 18 files. Updated `rounded-lg` → `rounded-xl` on 6 card/modal containers. Added `shadow-sm` to SessionCard. Updated ToastContext test description. Added 10 new visual refresh verification tests. Zero `indigo-` or `gray-` class references remain in the frontend source. All 154 frontend tests pass.

### Files modified
- `frontend/src/index.css` — Theme tokens: `--color-primary` → amber-600 (#d97706), `--color-primary-hover` → amber-700 (#b45309), `--color-primary-light` → amber-50 (#fffbeb)
- All 17 component/page `.tsx` files — `indigo-*` → `amber-*` accent colors
- All 18 component/page `.tsx` files — `gray-*` → `stone-*` neutral colors
- `frontend/src/components/SessionCard.tsx` — `rounded-xl`, `shadow-sm`
- `frontend/src/components/ConfirmModal.tsx` — `rounded-xl`
- `frontend/src/pages/SessionDetailPage.tsx` — `rounded-xl`
- `frontend/src/pages/MembersPage.tsx` — `rounded-xl`
- `frontend/src/pages/SessionCreatePage.tsx` — `rounded-xl`
- `frontend/src/pages/SessionEditPage.tsx` — `rounded-xl`
- `frontend/src/context/__tests__/ToastContext.test.tsx` — Updated test description "indigo border" → "amber border"

### New files
- `frontend/src/components/__tests__/VisualRefresh.test.tsx` — 10 verification tests (SessionCard amber RSVP button, stone card container, rounded-xl + shadow-sm, amber Cancel RSVP, amber attendee pills, amber capacity bar, no indigo classes; Toast amber info border; ConfirmModal stone neutrals + rounded-xl, stone cancel button border)

---

## Overhaul Phase 3: Backend Models & Repositories ✅ COMPLETED (merged into Phase 1)

Model structs, repository queries, and serialization tests completed as part of Phase 1.

**Remaining for later phases:**
- `UpdateImageURL` method on session repository (needed in Phase 8)
- `UpdateMemberRequest` profile fields for admin CRUD (deferred — spec says admins don't edit bio/skills/socials)

### Verify
- New fields appear in API responses from existing endpoints
- Existing tests still pass with updated structs

---

## Overhaul Phase 4: Profile API Changes + Public Profile Endpoint ✅ COMPLETED

**Completed:** Updated `PATCH /api/auth/profile` to accept bio, skills, linkedin_url, instagram_handle, github_username with full validation. Added `GET /api/profiles/{id}` public profile endpoint (auth required, returns public fields only — no email, admin status, or timestamps). Added `PublicMember` model, `GetByIDPublic` repository method, `ProfileHandler`, and route registration. 18 new tests added (12 service-level + 6 handler-level). All existing tests continue to pass.

### Files modified
- `internal/model/member.go` — Added `PublicMember` struct, extended `UpdateMemberRequest` with profile fields
- `internal/service/member.go` — Added `GetByIDPublic` to `MemberRepo` interface
- `internal/service/auth.go` — Rewrote `UpdateProfile` with `ProfileUpdateInput` struct and validation (bio ≤500 chars, skills ≤10 tags trimmed/lowercased, linkedin_url URL format, `@` stripping for instagram/github). Added `GetPublicProfile` method.
- `internal/handler/auth.go` — Updated `UpdateProfile` handler to use `ProfileUpdateInput`
- `internal/repository/member.go` — Added `GetByIDPublic` query (public fields only, active members only), added new field handling in `Update`
- `cmd/api/main.go` — Wired `ProfileHandler` into route registration

### New files
- `internal/handler/profile.go` — `ProfileHandler` with `GetPublicProfile` endpoint
- `internal/handler/profile_test.go` — 6 handler tests (200, 404 not found, 404 inactive, 400 invalid ID, 401 unauth, no private fields)

### Tests added (18 total)
- Service: bio valid, bio too long (422), skills valid (trim+lowercase), skills >10 (422), linkedin valid, linkedin invalid (422), instagram `@` strip, github `@` strip, all fields together, public profile found, not found, inactive
- Handler: GET 200 with data, 404 not found, 404 inactive, 400 invalid ID, 401 unauthenticated, no private fields leaked

---

## Overhaul Phase 5: Frontend Type Updates ✅ COMPLETED

**Completed:** Updated TypeScript types to match new backend API contracts. Added `PublicMember` and `UpdateProfileRequest` interfaces, extended `Member` (bio, skills, linkedin_url, instagram_handle, github_username), `SpaceSession` (image_url, location), `CreateSessionRequest` (location), and `UpdateSessionRequest` (location). Added `getPublicProfile(id)` and `uploadSessionImage(sessionId, file)` API client methods with `uploadRequest` helper for multipart/form-data. 4 new tests added (getPublicProfile success/404, uploadSessionImage success/error). All 112 frontend tests pass. TypeScript compilation clean.

### Files modified
- `frontend/src/types/index.ts` — Added `PublicMember`, `UpdateProfileRequest` interfaces; extended `Member`, `SpaceSession`, `CreateSessionRequest`, `UpdateSessionRequest`
- `frontend/src/api/client.ts` — Added `uploadRequest` helper, `getPublicProfile`, `uploadSessionImage` methods

### Tests added (4 total)
- `src/api/__tests__/client.test.ts` — getPublicProfile fetches by ID, getPublicProfile 404, uploadSessionImage sends FormData, uploadSessionImage error handling

---

## Overhaul Phase 6: Enhanced Profile Page + TagInput Component ✅ COMPLETED

**Completed:** Created reusable `TagInput` component with amber pill styling, Enter/comma to add, × to remove, max enforcement, duplicate prevention, and Backspace-to-delete-last. Enhanced `ProfilePage` with three sections (Identity, About, Social Links): name, telegram, bio with 500-char counter (turns red at ≤50 remaining), skills via TagInput (max 10), LinkedIn URL, Instagram handle, GitHub username. Read-only email + admin display. Save sends all fields via `PATCH /api/auth/profile`. 21 new tests added (10 TagInput + 11 ProfilePage). All 133 frontend tests pass.

### New files
- `frontend/src/components/TagInput.tsx` — Reusable tag input (amber pills, Enter/comma add, × remove, max limit, duplicate prevention, Backspace removes last)
- `frontend/src/components/__tests__/TagInput.test.tsx` — 10 tests
- `frontend/src/pages/__tests__/ProfilePage.test.tsx` — 11 tests (sections render, user data populated, bio counter, red counter at limit, submit all fields, null for empty fields, error toast, saving state, skills removal, null user)

### Files modified
- `frontend/src/pages/ProfilePage.tsx` — Rewritten with sectioned form (Identity, About, Social Links)

---

## Overhaul Phase 7: Member Profile Visibility + Guest List Links ✅ COMPLETED

**Completed:** Created `MemberProfilePage` component with read-only public profile view (name, telegram, bio, skills as amber pills, social links as clickable external links). Added `/profile/:id` route to `App.tsx`. Updated `GuestList` to make attendee names clickable `<Link>` elements pointing to `/profile/:id` with amber styling. Self-view detection shows "Edit Profile" link to `/profile`. 404 handling for missing/inactive members. 17 new tests added (11 MemberProfilePage + 6 updated GuestList). All 144 frontend tests pass.

### New files
- `frontend/src/pages/MemberProfilePage.tsx` — Read-only public profile view (fetches via `getPublicProfile`, conditional sections for bio/skills/links, self-view edit link, 404 state)
- `frontend/src/pages/__tests__/MemberProfilePage.test.tsx` — 11 tests (renders profile data, calls API with route param, social link URLs, self-view edit link, no edit link for others, 404 state, hidden optional sections, loading state, telegram @ stripping, instagram @ stripping in URL, new tab for social links)

### Files modified
- `frontend/src/App.tsx` — Added `/profile/:id` → `MemberProfilePage` route (auth required)
- `frontend/src/components/GuestList.tsx` — Names wrapped in `<Link to={/profile/${id}}>` with amber styling
- `frontend/src/components/__tests__/GuestList.test.tsx` — Updated all 6 tests: added `MemoryRouter` wrapper, verify names are links to correct `/profile/:id` paths

---

## Overhaul Phase 8: Image Upload Backend + Docker/Nginx Config ✅ COMPLETED

**Completed:** Implemented server-side image upload handling with multipart form parsing, magic-byte file type validation (JPEG, PNG, WebP), 5MB size limit via `MaxBytesReader`, server-generated filenames (`{session_id}-{timestamp}.{ext}`), old image cleanup on replacement, and idempotent delete. Added `UpdateImageURL` repository method, `UpdateImageURL`/`ClearImageURL`/`GetImageURL` service methods, `ImageHandler` with `Upload` and `Delete` endpoints. Routes registered under admin-only group. Static file serving via `http.FileServer` in dev, nginx `alias` in production. Docker volumes for image persistence. 15 new tests added. All 218 Go tests pass (except pre-existing deploy test), all 144 frontend tests pass.

### New files
- `internal/handler/image.go` — `ImageHandler` with `Upload` (multipart parse, magic-byte validation, file save, DB update, old file cleanup) and `Delete` (idempotent, file removal, DB clear) methods
- `internal/handler/image_test.go` — 15 tests (JPEG upload 200, PNG upload 200, invalid type 422, session not found 404, invalid ID 400, missing file 400, replaces existing image, file too large 413, delete 200, delete idempotent 200, delete not found 404, delete invalid ID 400, non-admin 403, unauthenticated 401, image URL format)
- `uploads/sessions/.gitkeep` — Ensure directory exists in repo

### Files modified
- `internal/repository/session.go` — Added `UpdateImageURL(ctx, id, imageURL)` method
- `internal/service/session.go` — Added `UpdateImageURL` to `SessionRepo` interface; added `UpdateImageURL`, `ClearImageURL`, `GetImageURL` service methods
- `internal/service/session_test.go` — Added `UpdateImageURL` to mock session repo
- `internal/handler/routes.go` — Added `imageHandler` parameter to `RegisterRoutes`; registered `POST /{id}/image` and `DELETE /{id}/image` in admin group
- `cmd/api/main.go` — Added `os.MkdirAll` for uploads directory, wired `ImageHandler`, registered `http.FileServer` for `/uploads/`
- `frontend/nginx.conf` — Added `/uploads/` location block with alias, 7-day cache, immutable header
- `docker-compose.yml` — Added `uploads` named volume, mounted on both `api` and `frontend` services
- `docker-compose.prod.yml` — Added `uploads` named volume, mounted on both `api` and `frontend` services

---

## Overhaul Phase 9: Sessions View Redesign (DateStrip + ImageUpload + Hero Card) ✅ COMPLETED

**Completed:** Replaced date-grouped session list with hero card + date picker strip layout. Created `DateStrip` component (horizontal scrollable date chips with amber highlight, auto-scroll to selected, tablist accessibility). Created `ImageUpload` component (drag-and-drop with client-side JPEG/PNG/WebP + 5MB validation, upload progress, replace/remove controls for existing images). Added `variant="hero"` prop to `SessionCard` (full-width h-48 image header, larger title, description display, location with map pin icon). Rewrote `SessionsPage` with DateStrip + hero cards (derives date chips from sessions, defaults to nearest upcoming date, filters sessions by selected date). Updated `SessionDetailPage` with hero image, location display, and admin-only `ImageUpload` section. Added location field to `SessionForm` (text input with map pin icon, included in create and edit flows). 17 new tests added (9 DateStrip + 6 SessionCard hero/location/image + 4 SessionForm location). Updated 3 existing test files to include `image_url` and `location` in `makeSession` helpers. All 174 frontend tests pass, all Go tests pass (except pre-existing deploy test).

### New files
- `frontend/src/components/DateStrip.tsx` — Horizontal scrollable date picker (amber-highlighted chips, day abbreviation + date number, tablist role, auto-scroll to selected)
- `frontend/src/components/ImageUpload.tsx` — Drag-and-drop file upload (client-side JPEG/PNG/WebP + 5MB validation, upload progress spinner, replace/remove controls on existing images, toast feedback)
- `frontend/src/components/__tests__/DateStrip.test.tsx` — 9 tests (renders chips, empty state, day/date display, aria-selected, amber highlight, click callback, stone unselected, tablist role, date ordering)

### Files modified
- `frontend/src/pages/SessionsPage.tsx` — Replaced date-grouped list with DateStrip + hero card layout; derives date chips from sessions; defaults to nearest upcoming date; filters by selected date
- `frontend/src/components/SessionCard.tsx` — Added `variant` prop (default/hero); hero shows h-48 image, larger title, description, location with map pin; both variants show location and "spots" suffix on capacity
- `frontend/src/pages/SessionDetailPage.tsx` — Added hero image at top, location display with map pin, admin-only ImageUpload section with cache invalidation
- `frontend/src/components/SessionForm.tsx` — Added location text input with map pin icon between description and date fields; included in create/edit submit handlers
- `frontend/src/components/__tests__/SessionCard.test.tsx` — Added 6 tests (location display, no location, hero image, no image in default, hero description, hero title size); updated makeSession with image_url/location; updated spot count text
- `frontend/src/components/__tests__/SessionForm.test.tsx` — Added 4 location field tests (renders, create includes, edit pre-populates, edit sends changes); updated makeSession
- `frontend/src/components/__tests__/VisualRefresh.test.tsx` — Updated makeSession with image_url/location fields

---

## Overhaul Phase 10: Testing Updates ✅ COMPLETED

Comprehensive test pass to ensure all new features work together and existing tests are updated.

### Backend tests ✅ ALREADY COVERED (added in Phases 1, 4, 8)

All backend test items were completed during their respective implementation phases:

- `internal/model/member_test.go` ✅ — Serialization tests for bio, skills, social fields (Phase 1: `TestMemberProfileFieldsSerialization`, `TestMemberProfileFieldsNullWhenUnset`, `TestMemberEmptySkillsSerialization`)
- `internal/model/session_test.go` ✅ — Serialization tests for image_url, location (Phase 1: `TestSpaceSessionImageURLAndLocation`, `TestSpaceSessionImageURLAndLocationNullWhenUnset`, `TestCreateSessionRequestWithLocation`, etc.)
- `internal/service/auth_test.go` ✅ — Profile update with all new fields, validation edge cases (Phase 4: 12 service-level tests)
- `internal/handler/profile_test.go` ✅ — Public profile endpoint tests (Phase 4: 6 handler tests)
- `internal/handler/image_test.go` ✅ — Image upload endpoint tests (Phase 8: 15 handler tests)
- `internal/handler/auth_test.go` ✅ — Updated me/profile response tests (existing)

### Frontend tests

- `src/components/__tests__/TagInput.test.tsx` ✅ — 10 tests (Phase 6)
- `src/components/__tests__/DateStrip.test.tsx` ✅ — 9 tests (Phase 9)
- `src/components/__tests__/ImageUpload.test.tsx` ✅ — 16 tests (drop zone render, accessible button role, file input accept attribute, image preview, Replace/Remove buttons, invalid file type error toast, file too large error toast, JPEG upload success, PNG upload success, WebP upload success, API error toast, generic error fallback, remove calls delete + onRemove, remove error toast, uploading spinner, disabled buttons during removal)
- `src/components/__tests__/SessionCard.test.tsx` ✅ — 21 tests incl. hero variant, image, location (Phase 9)
- `src/components/__tests__/GuestList.test.tsx` ✅ — 6 tests with profile links (Phase 7)
- `src/pages/__tests__/ProfilePage.test.tsx` ✅ — 11 tests (Phase 6)
- `src/pages/__tests__/MemberProfilePage.test.tsx` ✅ — 11 tests (Phase 7)
- `src/pages/__tests__/SessionsPage.test.tsx` ✅ — 16 tests: loading spinner, empty state (admin/non-admin), DateStrip rendering, default date selection (nearest upcoming + past fallback), date filtering via click, multiple sessions per date, hero variant (description, image, location), Create Session link visibility, date aggregation, aria-selected highlight, canceled session RSVP skip

### Test counts
- **Go tests:** 218+ (all pass except pre-existing deploy test)
- **Frontend tests:** 206 (all pass)

### Verify
- All frontend tests pass: `cd frontend && npx vitest run` ✅
- Manual end-to-end verification (remaining):
  1. Login and edit profile (bio, skills, social links) → save → verify persistence
  2. View another member's profile via guest list link
  3. Admin: create session with location → upload image → verify display
  4. Sessions page: DateStrip navigation, hero card with image + location
  5. Dark mode: all new components render correctly in both modes
  6. Mobile responsive: all new layouts work on small screens

---

## Dependency Graph

```
Overhaul Phase 1 (DB Migrations + Models)          ✅ DONE
  ↓
Overhaul Phase 2 (Visual Refresh)                    ✅ DONE ───┐
  ↓                                                              │
Overhaul Phase 3 (Backend Models & Repos)           ✅ DONE      │
  ↓                                                              │
Overhaul Phase 4 (Profile API + Public Profile)     ✅ DONE      │
  ↓                                                              │
Overhaul Phase 5 (Frontend Types)                   ✅ DONE ─────┘
  ↓
Overhaul Phase 6 (Enhanced Profile Page + TagInput) ✅ DONE
  ↓
Overhaul Phase 7 (Member Profile Visibility + Guest List) ✅ DONE
  ↓
Overhaul Phase 8 (Image Upload Backend + Docker) ✅ DONE
  ↓
Overhaul Phase 9 (Sessions View Redesign) ✅ DONE
  ↓
Overhaul Phase 10 (Testing Updates) ✅ DONE
```

**Parallel opportunities:**
- Overhaul Phase 2 (visual refresh) can run in parallel with Overhaul Phase 4 (backend changes) since they touch different layers
- Overhaul Phase 5 (frontend types) must wait for both Overhaul Phase 2 and Overhaul Phase 4 to complete
- Overhaul Phases 6–9 are sequential (each builds on the previous)

## Key Implementation Notes

1. **Migration safety** — All new columns are nullable or have defaults. Existing data is never affected. Migrations can be rolled back independently.
2. **Color swap approach** — Use global find-and-replace in frontend files. Verify no indigo/gray references remain (except status badge colors which are intentionally different).
3. **Image upload security** — Validate both Content-Type header AND magic bytes. Never trust client-provided filenames. Generate server-side filenames.
4. **TagInput UX** — Tags are trimmed, lowercased, and deduplicated. Max 10 enforced both client-side (disabled input) and server-side (422).
5. **DateStrip data** — Derive available dates from the sessions list response. Group by date, pass to DateStrip. No additional API call needed.
6. **Public profile privacy** — `GET /api/profiles/:id` returns NO email, NO admin status, NO timestamps. Only user-controlled public information.
7. **Image persistence** — Docker volume mount ensures uploads survive container restarts. Nginx serves them directly in production (no Go handler overhead).
8. **Profile self-edit** — `PATCH /api/auth/profile` remains the only profile edit endpoint. New fields are simply added to the accepted set. Admin member CRUD in spec 01 does NOT include bio/skills/socials (members control their own profiles).
