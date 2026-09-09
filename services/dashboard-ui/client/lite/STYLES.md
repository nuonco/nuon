# Style guidelines

The visual language of `client/lite/` — themes, tokens, type, spacing,
elevation, motion, and the accessibility floor.

This document answers **"what should this look like?"**. For *what to build and
how to organise it*, read [DESIGN.md](./DESIGN.md), which holds the five UX
patterns and the atomic design method. Engineering conventions are in
[DEV.md](./DEV.md), copy rules in [COPY.md](./COPY.md), and per-component rules
in each component's `Overview` story.

## Themes

Three themes — light, dark and high contrast. **System is not a fourth theme**;
it is the preference to follow the OS, and it resolves to light or dark.

- **Preference** (`light` / `dark` / `high-contrast` / `system`) is what the user
  picks. Persisted in `localStorage` under `nuon-lite-theme`.
- **Resolved theme** (`light` / `dark` / `high-contrast`) is what is on screen.
  Under `system` it tracks `prefers-color-scheme` live, with no reload.

Applied as `data-theme` on `<html>`: absent means follow the OS, otherwise the
named theme wins. Light and dark are defined once with CSS `light-dark()`, and
`data-theme="light"` / `"dark"` only switch `color-scheme` — so there is a single
definition per token, and a `system` user's first paint is correct with no
render-blocking script. High contrast is a flat override block, which is the
shape any future theme takes.

**High contrast** is the deliberate outlier: `#0000aa` blue surfaces, white body
text, `#ffff00` for tertiary text and accents. Every pair clears WCAG AAA
(13.3:1 white on blue, 12.4:1 yellow on blue). It is a real accessibility
option, not a novelty — **treat a regression in it as a regression.** Note that
`.status-tint` inverts strategy there (dark fill plus a `currentColor` outline
instead of a light wash) because the wash is what caps contrast.

## Tokens

Tokens live in `styles.css` in two layers:

1. **Primitives** — `--brand-*`, `--neutral-*`, `--error-*`. Raw values.
2. **Semantic roles** — `--surface-*`, `--text-*`, `--divider`, `--action-*`,
   `--status-*`, and per-component groups (`--card-*`, `--button-*`, `--field-*`,
   `--menu-item-*`, `--popover-*`, `--badge-*`, `--code-*`, `--diff-*`).

Names come from nuon.co, so the site and the app stay in step. `@theme inline`
maps the semantic layer onto Tailwind utilities — `bg-surface-default`,
`text-secondary`, `border-divider`, `text-status-error`.

### Rules

- **No `dark:` variants.** If a component needs one, a token is missing. This is
  what lets another theme or a white-label drop in without touching components.
- **No raw colour values and no stock Tailwind colour utilities.** No
  `bg-[#4cc9f0]`, no `text-zinc-500`. Semantic utilities only.
- **Primitives are referenced only by the semantic layer**, never by a component.
  Wanting `--brand-primary` in a component means a semantic token is missing.
- **Every token is defined in both themes.** `light-dark()` enforces this by
  construction — a token with one value is a token that does not theme.
- **Reading the resolved `theme` from `useTheme()` to pick a colour is a bug.**
  It is for choosing a non-CSS asset: a logo file, a syntax highlighting theme.
- Opacity modifiers (`bg-surface-accent/50`) are fine; they compose with tokens.

The exception is a **user-chosen colour from the API** — label colours arrive as
arbitrary CSS colours and are applied inline via `--status-color` /
`badge-custom`. Those are data, not design decisions, and high contrast ignores
them.

## Type

Six sizes, each with line-height and letter-spacing baked in. Use the `Text`
atom and its `variant`; never a raw `text-*` size class.

| Variant | Size | Default weight | For |
|---|---|---|---|
| `display` | 2rem | semibold | One per page at most. Focus/marketing surfaces. |
| `title` | 1.5rem | semibold | Page heading. |
| `heading` | 1.125rem | medium | Section heading. |
| `body` | 0.9375rem | normal | Default. Everything unmarked. |
| `caption` | 0.8125rem | normal | Secondary/supporting text, table meta. |
| `label` | 0.6875rem | medium | Eyebrows, modeline, chip text. Tracked +0.02em. |

Colour is a separate axis, and it defaults to `inherit` so text takes its
surroundings' colour unless told otherwise. Set it explicitly for `primary`
(main reading), `secondary` (supporting), `tertiary` (de-emphasised —
timestamps, IDs, metadata), `accent` or `positive`. **Do not encode hierarchy
with colour alone**; size and weight carry it first.

Two families: `sans` (Inter) and `mono` (Hack). Mono is for **machine-generated
or copyable strings** — IDs, versions, regions, code, log lines, diff bodies.
Not for emphasis.

## Spacing and radius

Spacing uses a tight subset of Tailwind's scale. The measured rhythm is
`gap-1` / `gap-2` / `gap-3` / `gap-4` / `gap-6`, with `gap-1.5` and `gap-0.5`
for inside-a-control adjustments and `gap-8` / `gap-10` for page-level
separation. Stay on it — a `gap-7` is a decision nobody else made.

**Always `flex flex-col gap-*`, never `space-y-*`.** There is currently zero
`space-y-` in Lite; keep it that way. Gap composes with `flex-wrap` and does not
leak into the last child.

Radius: `rounded-md` for small controls, `rounded-lg` for buttons, inputs and
menu items, `rounded-xl` for cards and surfaces, `rounded-full` for dots,
avatars and pills. The only additions are single-side radii (`rounded-l`,
`rounded-r`) where a control is welded into a group. Do not introduce another
step.

Dividers come from the base layer — `* { border-color: var(--divider) }` — so
write bare `border`, `border-t`, `divide-y` and let the colour come for free.
**Never set a border colour.**

## Elevation

`Card` is the only elevation primitive, and it composes four independent axes
rather than offering named "kinds":

- `padding` — `none` / `sm` (12px) / `md` (16px, default) / `lg` (24px)
- `opacity` — `default` (translucent wash) / `strong` / `solid`
- `blur` — `none` / `sm` / `md` (default) / `lg`
- `shadow` — `none` / `default` / `floating`
- `interactive` — adds the hover background for the chosen opacity

Rough guide: content blocks are default opacity with `shadow="default"`.
Anything floating above the page — popovers, toasts, menus — is `strong` or
`solid` with `blur="lg"` and `shadow="floating"`. Change `padding` rather than
fighting it with `!p-*`.

The app background is a dot grid (`.shell-background`, 24px, masked top and
bottom). It is shell chrome — do not put it behind content.

## Motion

Motion is short and functional. Two durations only: **150ms** for colour and
state changes, **200ms** for size and layout changes. Both `ease-out`.

Page transitions use the View Transitions API — a 180ms opacity-plus-4px slide,
driven by `view-transition-name: lite-page`. Every navigation path must opt in
by passing `viewTransition` to the router; `guardrails/view-transitions.test.tsx`
enforces that across `Link`, `Button href`, `NavLink`, `MenuItem` and keyboard
shortcuts.

**Every animation respects `prefers-reduced-motion`.** The stylesheet zeroes the
page transition, `status-pulse` and the shimmer; components use
`motion-reduce:transition-none`. An animation without that guard is a bug.

Nothing loops except `Spinner` and `status-pulse` — and `status-pulse` only
while a resource is genuinely in progress.

## Status colour

Six themes — `success`, `error`, `warn`, `info`, `brand`, `neutral` — resolved
from an API status string by `getStatusTheme()`. **Never map a status to a colour
at a call site**; pass the status and let `Status` resolve it. That is what keeps
one status looking identical in a table, a header and a toast.

Each theme pairs a colour with a distinct icon, so **colour is never the only
carrier**. `.status-tint` mixes the theme colour at 18% for chip fills.
`.status-pulse` is a 2s halo for in-progress work only.

Six themes is the whole vocabulary. Do not invent a seventh. Which *variant* to
use where is a System guidance decision — see [DESIGN.md](./DESIGN.md).

## Loading

**Loading is a `loading` prop on the primitive, never a `*Skeleton` component.**
A hand-built skeleton is a second copy of the layout measured by eye, and it
drifts.

The skeleton must occupy exactly the box the loaded content will, and the way to
guarantee that is to make the skeleton *the same kind of thing* as the content
rather than a box sized to match it. `Text` renders its real typographic classes
around a zero-width space with a `ch` width, so it inherits font size and
line-height for free. See `Text`'s `Overview` story.

- Chrome, labels and headings render real while their values load.
- Collections use the loading state built into `Table` and `Timeline`.
- Distinguish first load (`loading` — no data yet) from revalidation
  (`fetching` — data on screen). Revalidation must not blank the page.

## Accessibility baseline

Non-negotiable, and cheaper to do first than to retrofit:

- **Focus is always visible.** `focus-visible:outline-2
  focus-visible:outline-offset-2 focus-visible:outline-focus-ring`. Never
  `outline-none` without a replacement.
- **Disabled controls use `aria-disabled`, not the native attribute**, when they
  carry a tooltip explaining why — native `disabled` swallows pointer events, so
  the explanation never shows. `Button` already does this.
- Interactive elements are real `<button>` / `<a>` elements. A `div` with
  `onClick` is a bug.
- Icon-only controls carry an accessible name.
- Live regions: urgent toasts are `role="alert"` + `aria-live="assertive"`,
  everything else `role="status"` + `polite`.
- Decorative visuals — skeletons, spinner glyphs, the dot grid — are
  `aria-hidden`.
- All three themes clear WCAG AA; high contrast clears AAA.

## Anti-slop

Things that look like design work but are actually mess:

- A new colour, radius, duration or spacing step that exists once.
- A `dark:` variant.
- Colour as the only difference between two states.
- A gradient, a glow, or a shadow that is not `Card`'s.
- Emoji anywhere in the UI.
- A second component that does what an existing one does with different padding.
- Centre-aligned body text, or text over an image without a scrim.
- An animation longer than 200ms, or one that loops without a reason.
- More than one `primary` button in a view.
- A visual treatment applied to only one instance of a repeated thing.
