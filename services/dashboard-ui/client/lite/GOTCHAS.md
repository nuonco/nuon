# Gotchas

Things that cost real debugging time in `client/lite/` and are not obvious from
reading the code. Read this when something is behaving impossibly.

This is an append-only log of hard-won facts, not a rules document — rules live
in [DEV.md](./DEV.md), [DESIGN.md](./DESIGN.md) and [STYLES.md](./STYLES.md).
When you lose an hour to something non-obvious, add it here.

## CSS and Tailwind

**Tailwind resolves conflicting utilities by stylesheet order, not by their
order in the `className` string.** Putting `border-transparent` in a base class
silently beat `border-button-secondary-border` in a variant, and secondary
buttons rendered with no border. If two utilities set the same property, only
one can survive the merge — keep the conflicting one out of the base.

**`outline-none` in a base class silently kills a `focus-visible:outline-*`
ring.** Tailwind v4 carries outline *style* in `--tw-outline-style`:
`outline-none` sets it to `none`, and `focus-visible:outline-2` resolves
`outline-style` from that same variable. The width and the colour land, the ring
is invisible, and the classes read as correct — thirteen components shipped with
no keyboard focus ring this way. Same class of trap as `border-transparent`
above. The `focus-ring` utility in `styles.css` sets `outline-style` literally
inside a `&:focus-visible` block, so it outranks the base `outline-none` on
specificity rather than on stylesheet order.

**`border-box` does not save you when the height is `auto`.** A bordered variant
came out 38px while the others were 36px. Every `Button` variant now carries a
border, transparent where it is not visible, so geometry is identical across
variants.

**`inline-flex` does not stop a flex-column parent from stretching a child.**
Pills filled their container until `w-fit` was added — `Status` and `Badge` both
carry it for this reason. Do not remove it.

**Tailwind v4 Preflight no longer sets `cursor: pointer` on buttons.** It has to
be explicit, which is why `Button` sets `cursor-pointer`.

**`light-dark()` only accepts colours.** You cannot vary a percentage or a
length per theme with it. That is why the `color-mix` ratios in `styles.css` are
fixed values rather than themed.

## Layout and geometry

**A zero-width grid cell still gets the grid `gap`.** `Button`'s spinner column
is always in the DOM — collapsed to `0fr` so it can animate — which added 6px of
gap on the left only. Spacing lives on the icon and spinner as margins now, not
as a grid gap.

**Inline text takes its line box from the container's strut.** `Text` renders a
`span`, so as the only child of a padded box it inherits the container's
line-height and the padding goes lopsided. Pass `as="p"` or set the type on the
container. This bit both the tooltip body and the loading skeleton.

## Icons

**Phosphor glyphs sit in a 256 viewBox with roughly 12.5% transparent padding
per side**, so a 16px icon inks about 11.5px. That is why `Button`'s icon has
`-ml-0.5`, and why `Spinner` is a hand-drawn SVG rather than a Phosphor glyph.

**`Button` always renders its spinner `<svg>` first**, so a test reading
`button svg path` gets the spinner, not the icon. Take the last path.

## The code and diff renderer

`CodeBlock` and `Diff` wrap `@pierre/diffs`, which brings a set of traps.

**The renderer lives in a shadow root** (`<diffs-container>`). Nothing inside it
is reachable from `document.querySelector` — go through `host.shadowRoot`, or
use a Playwright locator, which pierces shadow DOM.

**That shadow host forces `color-scheme: dark`**, so every `light-dark()` token
inherited into it resolved to its dark branch and light-theme code rendered
white on white. Outer-document rules beat `:host` rules, so `styles.css` sets
`diffs-container { color-scheme: inherit }`. **Any third-party shadow component
we adopt needs the same check.**

**`CodeView` owns its own scroll root and must not be wrapped in the library's
`Virtualizer`.** Wrapping it silently disables virtualization — roughly 2000 DOM
spans instead of ~100 for the same document. No error, just a slow page.

**`CodeView` needs `overflow` as well as a max height.** Setting only
`maxHeight` left its content at full height, and the wrapper's `overflow-hidden`
clipped thousands of lines of JSON with no way to scroll to them.

**`data-line-index` in the shadow root is zero-based, while `scrollTo`'s
`lineNumber` is one-based.** Both indexing schemes meet inside the search-match
highlighting, which is painted by injecting a rule through the `unsafeCSS`
option (it lands in an `@layer unsafe`). This was wrong once.

**Bun tree-shakes a side-effect-only re-export.** A wrapper module around the
renderer's worker built to a 0-byte file. The build points at
`node_modules/@pierre/diffs/dist/worker/worker.js` directly, as a second
entrypoint — see the `build:js` script.

**Real payloads are enormous.** An install state is around 13,000 lines with a
single line of ~112,000 characters, because the stack template is embedded as a
JSON string. Virtualization is not an optimisation here, it is the only way these
views render — which is what all of the `CodeView` rules above are protecting.

## Tooling

**oxlint overrides are last-match-wins, and `files` globs resolve relative to
the config's directory.** From `client/.oxlintrc.json` the Lite glob is
`**/lite/**`, not `client/lite/**` — and `**/lite/**` also catches
`components/playground/lite/`, so the playground exemption has to sit *after*
the Lite boundary to win. `excludedFiles` is not supported. **A mis-scoped glob
fails silently as zero errors**, so verify any change with a throwaway probe
file that should error.

**`HTMLAttributes` collides with our slot props, and oxlint will not catch it.**
`title` is a native attribute typed `string`, so `title: ReactNode` on a
component extending `HTMLAttributes` is a type error — `Disclosure` needed
`Omit<…, 'title'>`, exactly as `Tooltip` needed it for `content`. The deprecated
native `color?: string` bites the same way when spreading onto `Text`. **Only
`tsc` finds these**, so run it before calling a component done.

**TypeScript 7 is the Go port and exposes no JS compiler API.** `require('typescript')`
returns only `{ version, versionMajorMinor }` — there is no `createScanner`, no
`ScriptTarget`. Anything that needs to parse TS (such as
`guardrails/comments.test.ts`) has to hand-roll it.

## Tests

**`createBrowserRouter` needs `window.location.origin`, and the full suite
destroys it.** `client/lib/api.test.ts` replaces `window.location` with a
spread object plus a mocked `reload`. Location's `origin` and `href` are
getters, so the spread copies neither, and every later `createBrowserRouter`
throws `No window.location.(origin|href) available to create URL`. `happyDOM`
`setURL` cannot put them back — the property is already a plain object.
`bun test client/lite` does not load that file, so the failure only appears
under `bun run test`. Lite router tests use `createMemoryRouter` and
`initialEntries` instead; it does not read `window.location`.

## Ladle

**There are two Ladle setups, and they must not collide:**

| Instance | Config | Port | Preview |
|---|---|---|---|
| Dashboard | `.ladle/` | 61000 | 61001 |
| Lite | `.ladle-lite/` | 62002 | 62003 |

**Ports must stay pinned in both configs.** Ladle passes Vite a fallback chain —
`[config.port, 61001, 62002, 62003, 62004, 62005]` — so an occupied port makes
the server silently move. Lite was configured at 61001, collided with the
dashboard's preview port, and came up on 62002 instead, which is why the two are
now explicitly separated by range. **The dashboard's 61000 is depended on by
systems outside this repo — do not change it.**

**Lite's Ladle binds to `::1`**, so `127.0.0.1` is refused. Use `localhost`.

**Story ids join every title segment with a double dash** —
`lite--molecules--disclosure--group`, not `lite-molecules-disclosure--group`.

**Ladle's theme control owns the Lite theme preference.**
`.ladle-lite/components.tsx` syncs `globalState.theme` into `setPreference`, so
writing `data-theme` before mount gets overwritten. Use `&theme=light|dark` in
the URL. There is no high-contrast option, so set `data-theme` *after* mount,
where the sync effect will not fire again.

## Scratch scripts and verification

**A scratch script must live in the repo's `tmp/`, not `/tmp`.** From `/tmp`,
Bun resolves `playwright` out of its global cache at a different version than
the project's, and Chromium refuses to launch.

**`navigator.clipboard` only exists in a secure context.** `page.setContent` in
Playwright is not one, so the copy verifier serves its page from a throwaway Bun
server on `127.0.0.1`. `useCopy`'s hidden-textarea fallback exists for the same
reason — it is a real code path, not defensive padding.

**The `verify-lite-*.mjs` scripts in `services/dashboard-ui/tmp/` measure real
computed styles in Chromium** rather than asserting on class names, which is how
most of the bugs on this page were found. They are gitignored and kept as a
lightweight regression suite. Several assert contrast ratios, so run them after
touching tokens in `styles.css`.
