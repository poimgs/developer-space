# 09 — Design Patterns

Shared design tokens and reusable component patterns for the co-working space app. All feature specs (01–04) reference this document for visual consistency. Built with Tailwind CSS.

## Color System

### Token Table

| Token | Light Mode | Dark Mode | Usage |
|-------|-----------|-----------|-------|
| `bg-primary` | `white` | `gray-900` | Page background |
| `bg-secondary` | `gray-50` | `gray-800` | Cards, panels |
| `bg-tertiary` | `gray-100` | `gray-700` | Hover states, nested panels |
| `text-primary` | `gray-900` | `gray-100` | Headings, body text |
| `text-secondary` | `gray-500` | `gray-400` | Subtext, labels, timestamps |
| `accent` | `indigo-600` | `indigo-400` | Buttons, links, active states |
| `accent-hover` | `indigo-700` | `indigo-300` | Hover on accent elements |
| `border` | `gray-200` | `gray-700` | Card borders, dividers |
| `status-scheduled` | `green-100` text `green-800` | `green-900` text `green-300` | Scheduled badge |
| `status-rescheduled` | `amber-100` text `amber-800` | `amber-900` text `amber-300` | Rescheduled (shifted) badge |
| `status-canceled` | `red-100` text `red-800` | `red-900` text `red-300` | Canceled badge |

### Accent Color

Indigo is the primary accent throughout the app — used for primary buttons, links, active nav items, and focus rings. It provides a professional, accessible contrast on both light and dark backgrounds.

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
| Info | `border-l-4 border-indigo-500` | Info circle | 3 seconds |

**Behavior:**
- Slide in from right, fade out on dismiss.
- Manual dismiss via X button.
- Maximum 3 toasts visible at once; oldest dismissed when a 4th arrives.

**Implementation:** React context (`ToastProvider`) + `useToast()` hook. Each toast has a unique ID and a timeout that auto-removes it.

### Confirmation Modals

Required for all destructive actions.

**Layout:**
- Centered overlay with semi-transparent backdrop (`bg-black/50`).
- White (dark: `gray-800`) card with `rounded-lg p-6 max-w-md` (desktop), near-full-width on mobile.
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
- **Inputs:** Full-width, `border rounded-md px-3 py-2`, focus ring in accent color.
- **Validation errors:** Below the field, `text-xs text-red-600 dark:text-red-400`, shown inline after submission or on blur.
- **Submit button:** Full-width on mobile, right-aligned on desktop.
- **Loading state:** Submit button shows a spinner icon + disabled (`opacity-50 cursor-not-allowed`) during submission. Button text changes (e.g., "Saving...").
- **Required fields:** Marked with a red asterisk (`*`) next to the label.

### Card Pattern

Container for list items (sessions, members on mobile).

```
Classes: bg-secondary border border-border rounded-lg p-4
```

- Light: `bg-gray-50 border-gray-200`
- Dark: `dark:bg-gray-800 dark:border-gray-700`
- Hover state (if clickable): `hover:bg-tertiary` transition.
- Used for session cards, member cards on mobile.

### Table Pattern

Used for structured data on desktop (members list).

- **Desktop (>= 768px):** Standard `<table>` with:
  - Header row: `bg-tertiary text-secondary text-xs uppercase tracking-wider`.
  - Body rows: `border-b border-border`, hover highlight.
  - Row actions as icon buttons (right-aligned column).
- **Mobile (< 768px):** Collapses to card list. Each row becomes a card (per card pattern above) with key-value pairs stacked vertically and actions inline.

## Button Styles

| Variant | Classes | Usage |
|---------|---------|-------|
| Primary | `bg-indigo-600 hover:bg-indigo-700 text-white dark:bg-indigo-500 dark:hover:bg-indigo-400` | Main actions (RSVP, Save, Create) |
| Secondary | `bg-transparent border border-indigo-600 text-indigo-600 hover:bg-indigo-50 dark:border-indigo-400 dark:text-indigo-400 dark:hover:bg-indigo-950` | Cancel RSVP, secondary actions |
| Destructive | `bg-red-600 hover:bg-red-700 text-white dark:bg-red-500 dark:hover:bg-red-400` | Delete, Cancel Session |
| Disabled | `bg-gray-200 text-gray-400 cursor-not-allowed dark:bg-gray-700 dark:text-gray-500` | Full sessions, loading states |

All buttons: `rounded-md px-4 py-2 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500`.
