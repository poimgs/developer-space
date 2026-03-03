# 09 — Design Patterns

Shared design tokens and reusable component patterns for the co-working space app. All feature specs (01–04) reference this document for visual consistency. Built with Tailwind CSS.

## Color System

### Token Table

| Token | Light Mode | Dark Mode | Usage |
|-------|-----------|-----------|-------|
| `bg-primary` | `white` | `stone-900` | Page background |
| `bg-secondary` | `stone-50` | `stone-800` | Cards, panels |
| `bg-tertiary` | `stone-100` | `stone-700` | Hover states, nested panels |
| `text-primary` | `stone-900` | `stone-100` | Headings, body text |
| `text-secondary` | `stone-500` | `stone-400` | Subtext, labels, timestamps |
| `accent` | `amber-600` | `amber-400` | Text on light backgrounds (links, labels, active states) |
| `accent-hover` | `amber-700` | `amber-300` | Hover on accent text elements |
| `border` | `stone-200` | `stone-700` | Card borders, dividers |
| `status-scheduled` | `green-100` text `green-800` | `green-900` text `green-300` | Scheduled badge |
| `status-rescheduled` | `amber-100` text `amber-800` | `amber-900` text `amber-300` | Rescheduled (shifted) badge |
| `status-canceled` | `red-100` text `red-800` | `red-900` text `red-300` | Canceled badge |

### Accent Color

Amber/gold is the primary accent throughout the app. `amber-500` (#F59E0B) is used for filled backgrounds where text is white (primary buttons, selected chips) — white-on-amber meets WCAG AA. `amber-600` (#D97706) is used for text on light backgrounds (links, secondary labels, active states) to meet WCAG AA 4.5:1 contrast. Combined with warm stone-based neutrals, it creates a warm, inviting atmosphere that reflects the co-working community feel.

## Dark Mode

### Strategy

Tailwind `darkMode: 'class'` — the `dark` class on the `<html>` element controls the mode.

### Initialization Logic

1. On first load, check `localStorage.getItem('theme')`.
2. If `'dark'` → add `dark` class to `<html>`.
3. If `'light'` → remove `dark` class.
4. If no stored preference → detect `window.matchMedia('(prefers-color-scheme: dark)')`. If matches, add `dark` class.

### Toggle Behavior

The toggle lives in the user menu dropdown. Clicking it:
1. Toggles the `dark` class on `<html>`.
2. Persists the new value to `localStorage.setItem('theme', 'dark' | 'light')`.

### Implementation

- Utility file: `src/lib/darkMode.ts` — exports `initTheme()` (called on app load) and `toggleTheme()`.
- Component: `ThemeToggle.tsx` — renders a sun/moon icon button, calls `toggleTheme()`.

## Typography & Spacing

Use Tailwind defaults throughout. No custom fonts.

- **Headings:** `text-2xl font-bold` (page titles), `text-lg font-semibold` (section headers), `text-base font-medium` (card titles).
- **Body:** `text-sm` or `text-base` depending on context.
- **Small text:** `text-xs` for badges, timestamps, helper text.
- **Spacing scale:** Tailwind default (4px base). Consistent padding: cards use `p-4` or `p-6`, page sections use `space-y-6`.

## Responsive Breakpoints

Tailwind defaults:

| Breakpoint | Min Width | Usage |
|------------|-----------|-------|
| `sm` | 640px | Minor layout adjustments |
| `md` | 768px | Primary mobile/desktop breakpoint |
| `lg` | 1024px | Wide desktop optimizations |

### Mobile (< 768px)

- Full-width cards, single column layout
- Hamburger navigation (slide-in overlay)
- Tables collapse to card lists (one card per row)
- Modals expand to near-full-screen
- Sticky date headers on session list scroll
- Form submit buttons full-width

### Desktop (>= 768px)

- Multi-column grids where appropriate
- Standard table layouts
- Centered modals with `max-w-md`
- Top navbar with visible links
- Form submit buttons right-aligned

## Shared Component Patterns

The following patterns are defined once here and referenced by feature specs. Each feature spec describes its specific content (text, icons, fields) but follows the layout and behavior defined below.

### Status Badges (Colored Pills)

Visual indicators for session status: scheduled, rescheduled (shifted), canceled.

```
Classes: rounded-full px-2.5 py-0.5 text-xs font-medium
```

| Status | Label | Light Colors | Dark Colors |
|--------|-------|-------------|-------------|
| `scheduled` | Scheduled | `bg-green-100 text-green-800` | `dark:bg-green-900 dark:text-green-300` |
| `shifted` | Rescheduled | `bg-amber-100 text-amber-800` | `dark:bg-amber-900 dark:text-amber-300` |
| `canceled` | Canceled | `bg-red-100 text-red-800` | `dark:bg-red-900 dark:text-red-300` |

### Toast Notifications

Ephemeral feedback messages for user actions.

**Position:** Top-right of viewport, stacked vertically with `gap-2`.

**Types:**

| Type | Left Border Color | Icon | Auto-dismiss |
|------|------------------|------|-------------|
| Success | `border-l-4 border-green-500` | Checkmark | 3 seconds |
| Error | `border-l-4 border-red-500` | X circle | 5 seconds |
| Info | `border-l-4 border-amber-500` | Info circle | 3 seconds |

**Behavior:**
- Slide in from right, fade out on dismiss.
- Manual dismiss via X button.
- Maximum 3 toasts visible at once; oldest dismissed when a 4th arrives.

**Implementation:** React context (`ToastProvider`) + `useToast()` hook. Each toast has a unique ID and a timeout that auto-removes it.

### Confirmation Modals

Required for all destructive actions.

**Layout:**
- Centered overlay with semi-transparent backdrop (`bg-black/50`).
- White (dark: `stone-800`) card with `rounded-xl p-6 max-w-md` (desktop), near-full-width on mobile.
- Content: title (bold) + descriptive message + two buttons.
- Buttons: secondary "Cancel" (left) + destructive action (right, red background).

**Behavior:**
- Clicking backdrop or pressing Escape closes the modal (equivalent to Cancel).
- Focus trapped inside the modal while open.

**Used for:** Cancel session, delete member, cancel RSVP.

### Empty States

Displayed when a list has no items.

**Layout:** Centered in the content area.
- SVG icon or illustration (muted color, `w-16 h-16`).
- Heading: `text-lg font-medium text-primary`.
- Subtext: `text-sm text-secondary`.
- CTA button (optional): primary button style.

Each feature spec defines its own icon, heading, subtext, and CTA. This pattern only defines the layout.

### Form Patterns

Consistent form behavior across all create/edit forms.

- **Labels:** Above inputs, `text-sm font-medium text-primary`.
- **Inputs:** Full-width, `border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2`, focus ring in accent color (`focus:ring-amber-500`).
- **Validation errors:** Below the field, `text-xs text-red-600 dark:text-red-400`, shown inline after submission or on blur.
- **Submit button:** Full-width on mobile, right-aligned on desktop.
- **Loading state:** Submit button shows a spinner icon + disabled (`opacity-50 cursor-not-allowed`) during submission. Button text changes (e.g., "Saving...").
- **Required fields:** Marked with a red asterisk (`*`) next to the label.

### Card Pattern

Container for list items (sessions, members on mobile).

```
Classes: bg-secondary border border-border rounded-xl p-4 shadow-sm
```

- Light: `bg-stone-50 border-stone-200`
- Dark: `dark:bg-stone-800 dark:border-stone-700`
- Hover state (if clickable): `hover:bg-tertiary hover:shadow-md` transition.
- Used for session cards, member cards on mobile, profile cards.

### Table Pattern

Used for structured data on desktop (members list).

- **Desktop (>= 768px):** Standard `<table>` with:
  - Header row: `bg-stone-100 dark:bg-stone-700 text-secondary text-xs uppercase tracking-wider`.
  - Body rows: `border-b border-border`, hover highlight.
  - Row actions as icon buttons (right-aligned column).
- **Mobile (< 768px):** Collapses to card list. Each row becomes a card (per card pattern above) with key-value pairs stacked vertically and actions inline.

## Button Styles

| Variant | Classes | Usage |
|---------|---------|-------|
| Primary | `bg-amber-500 hover:bg-amber-600 text-white dark:bg-amber-500 dark:hover:bg-amber-400` | Main actions (RSVP, Save, Create) |
| Secondary | `bg-transparent border border-amber-500 text-amber-700 hover:bg-amber-50 dark:border-amber-400 dark:text-amber-400 dark:hover:bg-amber-950` | Cancel RSVP, secondary actions |
| Destructive | `bg-red-600 hover:bg-red-700 text-white dark:bg-red-500 dark:hover:bg-red-400` | Delete, Cancel Session |
| Disabled | `bg-stone-200 text-stone-400 cursor-not-allowed dark:bg-stone-700 dark:text-stone-500` | Full sessions, loading states |

All buttons: `rounded-lg px-4 py-2 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500`.

## Additional Component Patterns

### TagInput

Reusable tag input for skills and similar tag-based fields.

**Layout:**
- Text input with pill-shaped tags displayed inline above or inside the input area.
- Each tag has an "×" remove button.
- Tags are added by pressing Enter or comma. Duplicates are silently ignored.
- Maximum tag count enforced (e.g., 10 for skills). Input is disabled when max reached.

**Styling:**
- Tags: `bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300 rounded-full px-2.5 py-0.5 text-xs font-medium` (same shape as status badges).
- Remove button: `ml-1 text-amber-700 hover:text-amber-800 dark:text-amber-400 dark:hover:text-amber-200`.
- Input: standard form input styling from [form patterns](#form-patterns).

**Props:** `value: string[]`, `onChange: (tags: string[]) => void`, `max?: number`, `placeholder?: string`.

### DateStrip

Horizontal scrollable date picker for the sessions home page.

**Layout:**
- Horizontally scrollable strip of date "chips", one per date that has a session.
- Each chip shows the day-of-week abbreviation (Mon, Tue, ...) and the day number.
- The selected/active date chip is highlighted with the accent color.
- Dates without sessions are not shown (only dates with at least one session appear).
- Auto-scrolls to center the selected date on load.

**Styling:**
- Container: `flex gap-2 overflow-x-auto py-2 px-1 scrollbar-hide`.
- Chip (default): `flex flex-col items-center px-3 py-2 rounded-xl text-sm cursor-pointer bg-stone-100 text-stone-600 dark:bg-stone-800 dark:text-stone-400 min-w-[3.5rem]`.
- Chip (selected): `bg-amber-500 text-white dark:bg-amber-500 dark:text-white shadow-md`.
- Day label: `text-xs font-medium uppercase`.
- Date number: `text-lg font-bold`.

**Behavior:**
- Clicking a date chip selects that date and displays its session(s) as hero card(s) below.
- On page load, defaults to the nearest upcoming date with a session.
- If multiple sessions exist on the same date, show them as stacked hero cards.

**Props:** `dates: { date: string; sessionCount: number }[]`, `selected: string`, `onSelect: (date: string) => void`.

### ImageUpload

File upload component for session images (admin only).

**Layout:**
- Dropzone area with a dashed border, camera/upload icon, and "Click or drag to upload" text.
- If an image is already set, shows a thumbnail preview with a "Remove" / "Replace" button.
- File type hint: "JPEG, PNG, or WebP, max 5MB".

**Styling:**
- Dropzone: `border-2 border-dashed border-stone-300 dark:border-stone-600 rounded-xl p-8 text-center cursor-pointer hover:border-amber-500 dark:hover:border-amber-400 transition-colors`.
- Preview: `relative rounded-xl overflow-hidden` with overlay controls on hover.
- Upload progress: amber-colored progress bar.

**Behavior:**
- Accepts click-to-browse or drag-and-drop.
- Validates file type (JPEG, PNG, WebP) and size (max 5MB) client-side before upload.
- Shows upload progress indicator.
- On success, displays thumbnail preview and emits the new `image_url`.
- On error, shows toast with error message.

**Props:** `sessionId: string`, `currentImageUrl?: string`, `onUpload: (imageUrl: string) => void`.
