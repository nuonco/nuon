# Dashboard UI Design Guide

This guide defines the visual and design conventions for the Nuon dashboard `client/` SPA: Stratus
tokens, spacing rhythm, anti-slop rules, interaction patterns, and accessibility. The goal is UI
that is indistinguishable from the rest of the product — same components, same tokens, same spacing
rhythm, same interaction patterns. **Consistency first, cleverness never.**

This document defers to two others — do not restate or reinterpret their rules here:

- **Code conventions** (architecture, container/component pattern, stories, icons, dates, modals):
  [AGENTS.md](./AGENTS.md)
- **All user-facing copy** (voice, capitalization, buttons, modals, toasts, empty states, errors):
  [COPY_STYLE.md](./COPY_STYLE.md)

The **visual** source of truth is the Stratus Figma file (see below).

### Stratus design system (Figma)
The complete component index — every component, variant, token, and where it lives — is in the
Stratus Figma file:

> **Stratus Design System (Figma):**
> https://www.figma.com/design/K3IcokRSyYRkH2tsr1KMhV/Stratus-Design-System?node-id=11190-79898

When a user asks for a design pattern, component, or "where does X live", point them to this file so
they can see the link index and locate the exact node. If a connected Figma MCP is available, pull
the node context directly from this file instead of eyeballing.

---

## 0. Operating principles (read first)

1. **Reuse before you build.** Almost everything already exists in `components/common/` and
   `components/surfaces/`. Find the existing component/pattern and use it. Do not hand-roll a button,
   modal, table, badge, dropdown, or input.
2. **Match the nearest precedent.** Before designing a screen, find the most similar existing screen
   and mirror its structure, spacing, and component choices. Consistency with what ships beats a
   locally "nicer" one-off.
3. **Use tokens, never raw values.** Colors, spacing, radius, type, motion all come from Stratus
   tokens (see §2). Hardcoded hex, arbitrary `px`, or off-scale spacing is a defect.
4. **Less chrome.** Prefer the lightest structure that communicates hierarchy. Extra borders,
   nested cards, and boxes are usually wrong (see §4, anti-slop).
5. **Ask only when truly blocked.** Otherwise pick the option most consistent with existing
   patterns and proceed, noting the decision.

---

## 1. Component library (use these)

Surfaces: `Modal`, `Panel`, `Toast` (`components/surfaces/`).
Common (`components/common/`), non-exhaustive: `Button`, `Card`, `Badge`, `Banner`, `Divider`,
`Dropdown`, `Menu`, `Expand`, `Icon`, `Link`, `Text`, `Time`, `Duration`, `Cron`, `Status`,
`LabeledStatus`, `LabeledValue`, `KeyValueList`, `PropertyGrid`, `ID`, `Hash`, `ClickToCopy`, `Code`,
`CodeBlock`, `JSONViewer`, `Loading`, `EmptyState`, `Avatar`, `Group`, `HeadingGroup`, `BackLink`.

Rules:
- Prefer these over native elements. Native `<button>`, ad-hoc `<input>`, or a bespoke table is a
  smell unless the design system genuinely lacks the primitive.
- Use semantic helpers for data: `Time` for timestamps (never manual date formatting), `Duration`,
  `ID`/`Hash`/`ClickToCopy` for identifiers, `Status`/`Badge` for state, `Code`/`JSONViewer` for
  payloads, `LabeledValue`/`PropertyGrid`/`KeyValueList` for key→value display.
- Anything using `Modal`/`Panel`/`Toast` (or `useSurfaces`) must be mounted under `SurfacesProvider`
  — including Ladle stories, or it throws. In stories, use the `ModalStory`/`PanelStory` helpers
  from `components/__stories__/helpers.tsx`.
- Code conventions for these components (container/component split, stories format, icon imports,
  date handling) live in [AGENTS.md](./AGENTS.md) — follow it rather than restating rules here.

---

## 2. Stratus tokens (single source of truth)

Defined in `styles.css` via `@theme inline` and `:root` (light) + `prefers-color-scheme: dark`.
**Never introduce raw hex or off-scale values — reference these.**

### Color scales (50→950)
`primary` (purple, brand — 600 = `#8040BF`), `blue`, `green`, `orange`, `red`,
`cool-grey` (light-mode neutrals), `dark-grey` (dark-mode neutrals).

### Semantic / theme-aware tokens (preferred over picking a raw scale step)
- Page background: `--background` · neutral surface: `--background-neutral`
- Text: `--foreground`
- Border: `--border-color` (light `#dee3e7` / dark `#27252a`)
- Code surface: `--bg-code`
- Brand gradient (text): `.text-gradient`

### Surface convention (light ↔ dark must always be paired)
- Default surface: page `--background`.
- Secondary surface (e.g. table headers, subtle fills): `bg-cool-grey-100 dark:bg-dark-grey-700`.
- Borders/separators: rely on the border token (see §4) — light `cool-grey`/dark `dark-grey` family.
- Accent / interactive text: `text-primary-600 dark:text-primary-400`.
- Focus ring (purple): `rgba(128,64,191,0.64)`.

### Type
- Sans: **Inter** (`font-sans`), Mono: **Hack** (`font-mono`).
- Weights are tokens: `normal` 400, `strong` 500, `stronger` 600. Use `Text` `weight="strong"`
  etc.; do not reach for arbitrary `font-bold`/`font-semibold`.
- Use `Text` `variant` (`h1 | h2 | h3 | base | body | subtext | label`, default `body`) and `theme`
  (`default | neutral | info | warn | error | success | brand`) instead of raw size/color. For muted
  secondary text use `theme="neutral"` (there is no `tertiary` theme).

### Motion
- Easing: `ease-cubic`. Durations: `--duration-fastest` 150ms, `--duration-fast` 250ms.
- Keep transitions subtle and fast; animate opacity/transform, not layout.

**Every light-mode token you add must have its dark-mode pair.** A color that only looks right in one
theme is a defect.

---

## 3. Spacing, layout & rhythm

- Use the Tailwind spacing scale (multiples of 4px). No arbitrary values like `gap-[13px]`.
- **`Card`** defaults to `p-6` / `gap-6` (24px). If you override one, override both so padding and
  child spacing read intentionally:
  ```tsx
  <Card className="!p-4 !gap-4">…</Card>   {/* a consistent 16px rhythm */}
  ```
  Never mix `!p-4` padding with the default `gap-6` — mismatched rhythm is a slop. (`PageSection`
  is the exception: it has a `flush` prop — use it instead of `!p-0 !gap-0`, per AGENTS.md.)
- **Full-bleed dividers:** a divider inside a padded card must reach the card edges. Offset by the
  exact padding:
  ```tsx
  <Card className="!p-4 …"><hr className="-mx-4" /></Card>   {/* -mx must equal the padding */}
  ```
- Establish hierarchy with **spacing and type weight**, not with borders and boxes.

---

## 4. Anti-slop rules (do not do these)

These are the specific inconsistencies that make the UI look "off". Treat each as a bug.

1. **Don't restyle borders.** Stratus globally pins every `border*` to `--border-color` at a 1px
   width. Do not set custom border colors (`border-red-300`, `border-cool-grey-400`, …) or custom
   widths (`border-2`, `border-l-4`) for layout chrome. Borders are uniform, thin, and tokenized.
   A heavier or differently-colored side border is a slop. (Color is fine only where it is *semantic*
   state via an existing component, e.g. a `Banner`/`Status` variant.)
2. **No deviating colors.** No raw hex, no off-palette shades, no "close enough" greys. Pull the exact
   token. If you think you need a new color, you almost certainly don't — flag it for the user.
3. **No unwanted bounding boxes.** Don't wrap content in an extra `Card`/`border rounded` just to
   "group" it. Cards nested in cards, or a box around a single value, add visual noise. Use spacing,
   a `Divider`, or a `LabeledValue` instead. Add a container only when it represents a real, separable
   surface.
4. **No off-scale spacing/radius.** Stick to the spacing scale and the system's radius tokens
   (`rounded-md`, `rounded-lg`). No `rounded-[7px]`.
5. **No font ad-hocery.** No arbitrary sizes/weights; use `Text` props and the weight tokens.
6. **No layout-thrash animation / flicker.** See §7 (container queries) — labels must not flip
   between states on interaction.
7. **No orphaned dark mode.** Every surface/border/text color must be correct in both themes.

---

## 5. Core interaction patterns (match these exactly)

### Collapsible header ("Show more / Show less")
Make the whole header row the control; keep it keyboard accessible. A `<div role="button">` is
acceptable here (and only in cases like this) because the control is a full row containing other
content — `Button` cannot represent it. For ordinary actions, always use `Button`.
```tsx
<div
  role="button" tabIndex={0} aria-expanded={expanded}
  onClick={() => setExpanded(p => !p)}
  onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setExpanded(p => !p) } }}
  className="flex items-center justify-between gap-3 cursor-pointer select-none focus:outline-none"
>
  <div className="flex items-center gap-1.5 min-w-0">…</div>
  <Text
    as="span"
    variant="subtext"
    weight="strong"
    className="flex items-center gap-1 shrink-0 text-primary-600 dark:text-primary-400"
  >
    {expanded ? 'Show less' : 'Show more'}
    <Icon variant={expanded ? 'MinusIcon' : 'PlusIcon'} size={14} />
  </Text>
</div>
{expanded && <>…</>}
```

### Modal
- Sizes `sm | default | lg | xl | full` (`default` when unset). Use `xl` for dense content (build
  lists, metadata, multi-field forms).
- Constrain tall content (`className="h-[80vh]"`) so it scrolls internally.
- The standard pattern is **two components** — a Modal component plus a Button component that opens
  it via `useSurfaces().addModal` (see "Modal and Panel Components" in [AGENTS.md](./AGENTS.md)).
  The `triggerButton` prop is a shortcut for trivial, single-trigger cases only.

### Select
- Use native slots: `labelProps`, `helperTextProps` — don't hand-build label/helper text.
- Helper text defaults to muted `subtext` styling — pass copy through `helperText`, don't restyle it.
- Label and helper copy follow [COPY_STYLE.md](./COPY_STYLE.md) (sentence case, "(optional)"
  suffix on the label, no repeating the label in help text).
- Represent "default / none" as the **first option in the dropdown**, not a separate reset button.

### Checkbox with description
Whole label block (title + description) must be clickable. `labelText`/`labelTextProps` are nested
inside `labelProps`:
```tsx
<CheckboxInput
  labelProps={{
    labelText: (
      <>
        <Text weight="strong">Deploy dependents</Text>
        <Text variant="subtext" theme="neutral">…</Text>
      </>
    ),
    labelTextProps: { as: 'div' },
  }}
  className="items-start"
/>
```

### Buttons
- Variants: `primary | secondary | ghost | danger | tab | icon`.
- `icon` = square, icon-only (`aspect-square rounded-md !p-0`, ghost-style hover/focus). Use for
  caret/arrow affordances next to a row instead of a bordered button.
- One **primary** action per surface; everything else `secondary`/`ghost`. `danger` only for
  destructive actions.

### Tables
Prefer the Stratus `Table`. If you must hand-roll (e.g. inside a modal), mirror its styling:
- Header row on the secondary surface: `bg-cool-grey-100 dark:bg-dark-grey-700`.
- Header cells `py-3 px-4 font-normal font-sans` (not bold), rounded top corners on the ends.
- Body rows separated by `border-t` (token color, never a custom border), `align-top`,
  `whitespace-nowrap` for fixed cols.
- Use an inline table over an accordion when each row's content is short and scannable.

### Destructive confirmation
Follow the three severity tiers in [COPY_STYLE.md](./COPY_STYLE.md) (simple confirm → warning
banner → type-to-confirm). For tier 3, require typing the **exact identifier** (e.g. the org name),
not a generic phrase:
```tsx
const orgConfirmText = org?.name || orgId
<AdminActionCard inputText={orgConfirmText} … />
```

### Links — internal first
Prefer in-app navigation; only fall back to external URLs. Drive `isExternal` off the internal href:
```tsx
<Link href={href ?? `https://github.com/${name}`} isExternal={!href}>…</Link>
```

### Resizable split panel
`flex` layout, width via a CSS var; draggable `role="separator"` with `tabIndex={0}`,
`aria-orientation`, `aria-label`, and arrow-key nudging. Stack/hide divider on small screens
(`hidden lg:flex`).

---

## 6. Accessibility (baseline, not optional)

- Every interactive element is keyboard operable: `role`, `tabIndex={0}`, `aria-*`
  (`aria-expanded`, `aria-label`, `aria-orientation`), and key handlers (Enter/Space; arrows for
  sliders/separators).
- Visible focus state on all controls (use the system focus ring, don't remove outlines without a
  replacement).
- Respect color-contrast in both themes; never encode meaning in color alone — pair with text/icon.
- Icon-only controls need an `aria-label`.
- Don't trap or steal focus; modals/panels manage focus via the surfaces system.

---

## 7. Container queries — avoid label flicker

Responsive show/hide (e.g. collapsing a button label to an icon) must key off a **stable** parent
width. Putting `@container` on a wrapper whose own width depends on its dynamic content creates a
feedback loop (label toggles → width changes → query flips → label toggles): the symptom is labels
flickering between text and icon-only on click.

Fix: declare `@container` on the outer, content-independent row, and gate child labels with width
breakpoints:
```tsx
<div className="@container flex items-center justify-between">   {/* stable */}
  <Button variant="ghost"><Icon … /><span className="@max-[30rem]:hidden">User</span></Button>
</div>
```

---

## 8. UX principles (general best practice)

Apply widely-accepted interaction-design heuristics (stated generically; do not cite vendors/brands):

- **Consistency & standards.** Same action looks and behaves the same everywhere. Match existing
  patterns over inventing new ones.
- **Visual hierarchy.** Guide the eye with type weight, size, and spacing — not borders/boxes. One
  clear primary action per view.
- **Progressive disclosure.** Show the essential first; tuck advanced/secondary detail behind
  expanders, modals, or panels (e.g. metadata in a modal, "Show more" on detail headers).
- **Recognition over recall.** Surface options and current state; don't make users remember values.
  Labeled fields, sensible defaults, inline helper text.
- **Feedback & system status.** Reflect loading, success, and error states (`Loading`, `Toast`,
  `Banner`, `Status`). Never leave an action without acknowledgement.
- **Error prevention.** Guard destructive actions (typed confirmation), validate inline, prefer safe
  defaults.
- **Forgiveness.** Make actions reversible where possible; confirm only the irreversible.
- **Minimize cognitive load.** Reduce choices per step; group related fields; keep copy terse and
  styled per [COPY_STYLE.md](./COPY_STYLE.md).
- **Aesthetic-functional integrity.** Clean, uniform spacing and tokens read as "trustworthy".
  Inconsistency reads as "broken".
- **Empty, loading, and error states are part of the design** — design all three, don't bolt them on.

---

## 9. Definition of done (self-check before handing off)

- [ ] Reused existing components/patterns; nothing hand-rolled that the library provides.
- [ ] Every color/space/radius/type value is a Stratus token; no raw hex or off-scale values.
- [ ] Borders are the default 1px token — no custom border colors or widths; no extra bounding boxes.
- [ ] Light **and** dark mode both verified.
- [ ] Spacing rhythm is consistent (padding and gaps agree).
- [ ] Keyboard + focus + `aria-*` covered for every interactive element.
- [ ] Loading / empty / error states handled.
- [ ] All user-facing copy follows [COPY_STYLE.md](./COPY_STYLE.md).
- [ ] Matches the provided Figma/diagram, or the closest existing precedent if none was provided.
- [ ] If a genuinely new pattern was introduced, it was flagged and confirmed with the user.
