---
name: dashboard-ui:component
description: Use when adding or using a UI component in the dashboard-ui client/ SPA
---

This skill enforces checking existing components before creating new ones, using the container/component pattern, and including Ladle stories.

## Steps

1. Run `ls client/components/common/` and the relevant domain directory (e.g., `client/components/actions/`, `client/components/installs/`). Read the filenames.
2. If an existing component meets your needs, use it. Do NOT create a new component that duplicates an existing one.
3. Before writing JSX that uses a component, **read its `.stories.tsx` file first** — stories are the primary reference for correct prop usage, patterns, and edge cases. Then read the `interface I*` props definition in the source file. Do not guess prop names.
4. For Modal or Panel: always use `Modal` or `Panel` from `client/components/surfaces/`. Never use `ModalBase` or `PanelBase` directly.

## Container / Component Pattern

Feature components that fetch data or use context hooks must use this pattern:

```
client/components/[domain]/MyComponent/
├── MyComponent.tsx              ← Pure presentational (props in, JSX out)
├── MyComponentContainer.tsx     ← Data-fetching wrapper (hooks, queries, mutations)
├── MyComponent.stories.tsx      ← Ladle stories (required)
├── index.ts                     ← Barrel export
```

**`MyComponent.tsx`** — No `useQuery`, `useMutation`, or context hooks that require providers. All data comes via props.

**`MyComponentContainer.tsx`** — Calls hooks (`useOrg()`, `useQuery()`, etc.) and passes resolved data to the presentational component.

**`index.ts`** — Exports the container as the default/primary name:
```typescript
export { MyComponentContainer as MyComponent } from './MyComponentContainer'
export { MyComponent as MyComponentComponent } from './MyComponent'
```

**Simple presentational components** (no data-fetching) stay as flat files in `client/components/common/MyComponent.tsx` or `client/components/[domain]/MyComponent.tsx`.

**Never have both a flat file `MyComponent.tsx` and a directory `MyComponent/` at the same level** — the flat file shadows the directory's `index.ts` and causes import resolution bugs.

## Ladle Stories (Required)

Every component directory must include a `.stories.tsx` file. Stories use **Ladle v5** format — NOT Storybook.

```tsx
// ✅ Correct
export default { title: 'Domain/MyComponent' }
import { MyComponent } from './MyComponent'
export const Default = () => <MyComponent items={mockItems} />
```

```tsx
// ❌ Wrong — breaks Ladle with "got: object" error
import type { StoryObj } from '@ladle/react'
export const Default: StoryObj = { render: () => <MyComponent /> }
```

**Stories render the presentational component** with mock props — not the container.

**When a child component needs a context provider**, mock the context:
```tsx
import { SomeContext } from '@/providers/some-provider'
export const Default = () => (
  <SomeContext.Provider value={mockValue}>
    <MyComponent />
  </SomeContext.Provider>
)
```

**Modal stories** use the `ModalStory` helper:
```tsx
import { ModalStory } from '@/components/__stories__/helpers'
export const Default = () => <ModalStory><MyModal data={mock} /></ModalStory>
```

**Timeline stories** — mock items must have unique `created_at` on different calendar days.

**Do not** wrap stories in `MemoryRouter` — Ladle provides one globally.

## Text & Copy Style

All user-facing text follows `services/dashboard-ui/COPY_STYLE.md` — read it before writing copy. Quick rules: sentence case everywhere (never title case), buttons are verb + object, "[thing] failed" error headings, no exclamation marks / "please" / "successfully".

- ✅ "Create your org" / "Connect a cloud account"
- ❌ "Create Your Org" / "Connect A Cloud Account"

For visual decisions (tokens, spacing, borders, dark mode), follow `services/dashboard-ui/DESIGN.md` — no raw hex colors, off-scale spacing, or custom border colors/widths.

For `Tabs`: keys are rendered via `toSentenceCase(camelToWords(key))` which lowercases everything after the first character — always write keys all-lowercase (`'create your own app'`, not `'Create Your Own App'`).

## Rendered strings (casing, font, chip)

Every string you render is exactly one of three classes; the class fixes its casing, font, and (as a chip) its component. Classify with these tests, then apply the treatment. Full record: `DESIGN.md` §1 / UXDR 019.

1. **Identifier** — *Could a user copy this exact string into a terminal, config, or search?* (names, IDs, k8s kinds, terraform/cloud resource types, image tags, paths, versions, label keys/values) → **verbatim, never re-cased, mono** (`ID`, `Code`, `Badge variant="code"`, `LabelBadge`, `Text family="mono"`).
2. **API vocabulary** — *Is the full value set enumerable from our Go code?* (statuses, trigger/workflow/plan/step types, ops) → **`humanize()` from `@/utils/string-utils` (sentence case, acronyms preserved), sans**. `Status` humanizes for you; never re-case vocabulary by hand at a call site.
3. **UI copy** — neither of the above → `COPY_STYLE.md`.

**Chip pick:** value changes on its own while you watch (lifecycle) → `Status`; static classification (type/kind/count/version) → `Badge` (`variant="code"` iff identifier content, default sans variant iff vocabulary); user key:value label → `LabelBadge`. **mono ⇔ identifier.** Never `humanize()` a string that contains a user identifier (mixed strings, e.g. a step name) or a free-text API sentence (`status_human_description`) — render those verbatim.

## Icons

Use the `Icon` component from `@/components/common/Icon` for ALL icons. Always use the `Icon` suffix for variant names (e.g., `HouseIcon` not `House`).

If you need a Phosphor icon that isn't already available, add it to `client/components/common/Icon.tsx`:
1. Add the named import: `import { NewIconNameIcon } from '@phosphor-icons/react'`
2. Add it to the `phosphorIcons` object: `NewIconNameIcon,`

Never import directly from `@phosphor-icons/react`, `lucide-react`, or `heroicons` in component files.

## Links

Import `Link` from `@/components/common/Link` (never from `react-router`; it uses `href`, not `to`). Every content link is one of three classes (full taxonomy in `DESIGN.md` §5 "Links" / `COPY_STYLE.md#links`):

- **Entity link** — the resource's own name is the link text; the name navigates. No verb, no icon.
- **View link** — a standalone `View {resource}` link (`View plan`, `View logs`, `View all runs`; `View details` only when no better noun). Wrap in `<Text variant="subtext">` for sizing.
- **External link** — set `isExternal`; the new-tab icon renders automatically. Never hand-place `ArrowSquareOutIcon`.

A `Link` never carries a text-size class, a trailing `CaretRightIcon`/`ArrowRightIcon`, or a manual external icon. **Row navigation is the entity link, not an icon-only `<Button href><Icon/></Button>`** (deprecated — icon-only buttons are for non-nav chrome like modal/panel close only). Leading *content* icons (a `GitBranchIcon` before a branch name) are fine.

## Button tooltips (disabled reasons & nudges)

The `Button` owns its tooltip via `tooltipProps` (`Omit<ITooltip, 'children'>`). **Never hand-wrap `<Tooltip>` around a `Button`, and never put `title=` on a button.** With `disabled` + `tooltipProps`, Button renders `aria-disabled` (not native `disabled`, which swallows pointer events) so the reason shows on hover and keyboard focus.

```tsx
<Button disabled tooltipProps={{ tipContent: 'Sync the app config first' }}>Trigger run</Button>
```

- **Every disabled button whose reason isn't obvious from context gets a `tooltipProps` reason.** Copy follows `COPY_STYLE.md`: sentence case, explains the unmet condition, fragment (no period). A plain-string `tipContent` auto-wraps in `Text subtext` — pass a string, don't wrap it.
- "Obvious from context" (no tooltip needed): label already changes for async ops, form fields show their own errors, a type-to-confirm input is right above, or pagination/positional convention.
- **Nudge** (controlled tooltip opened by app state): `useNudge(trigger)` → `{ isOpen, close }` + `tooltipProps={{ isOpen, disableHover: true, tipContent }}`. Don't re-implement the timer.
- Tooltips on **non-Button** elements (text, icons, badges, toggles) keep the hand-wrapped `<Tooltip>` — `tooltipProps` is Button-only.

## Loading states

**Never hand-build a `*Skeleton` component** — a hand-measured skeleton is a second copy of the layout that drifts from the real component. Skeletons are derived from the real components:

- **Primitive `loading` prop**: `Text`, `LabeledValue`, `LabeledStatus`, `ID`, `Time`, `Duration`, `Badge`, `Status`, `Code` all take `loading?: boolean` (+ `loadingWidth?: number` in `ch`). The container passes `isLoading` down; the presentational component fans it into `loading` props.
- **Chrome and labels always render real** — only unknown values shimmer.
- **Collections**: `<Table isLoading>` / the `Timeline` loading state own their skeleton — never a bespoke row skeleton.
- **Spinner** (`<Loading variant="large" />`) only for genuinely unknown-shape content (plan diffs, dynamic-field forms, unknown outputs). Known shape → loading primitives.
- Direct use of `common/Skeleton` in a feature component is a review smell (it's the low-level block the primitives use internally). See `DESIGN.md` §5 "Loading states".

## Anti-Patterns

- **Do not** hand-wrap `<Tooltip>` around a `Button` or put `title=` on a button — use `Button`'s `tooltipProps`
- **Do not** leave a disabled button unexplained when its reason isn't obvious from context — add a `tooltipProps` reason
- **Do not** create a component that duplicates an existing one — always check existing components first
- **Do not** pass props to a component without reading its interface — wrong props cause runtime errors
- **Do not** use `ModalBase` or `PanelBase` directly — always use the `Modal`/`Panel` wrappers from `surfaces/`
- **Do not** put a domain-specific component (e.g., `InstallCard`) into `client/components/common/`
- **Do not** leave both a flat file and directory with the same name — delete the flat file after migrating to the directory pattern
- **Do not** skip the `.stories.tsx` file — every component directory must have one
- **Do not** use `StoryObj` or `render:` in stories — Ladle v5 requires plain function exports
- **Do not** import icons directly from `@phosphor-icons/react` — always use the `Icon` component
- **Do not** put a text-size class, a trailing `CaretRightIcon`/`ArrowRightIcon`, or a manual `ArrowSquareOutIcon` on a `Link`; standalone view links wrap in `<Text variant="subtext">` and external links get their new-tab icon from `isExternal`
- **Do not** use an icon-only `<Button href><Icon/></Button>` for row navigation (the entity link — the resource name — is the navigation), or a non-"View" link verb ("See"/"Open"/"Go to" all become "View")
- **Do not** hand-build a `*Skeleton` component — use the primitive `loading` prop, `<Table isLoading>`, or a spinner for unknown shape
- **Do not** render a raw API enum (`{role.type}`, `{step.type}`) or re-case vocabulary at a call site (`toSentenceCase`/`toTitleCase`/hand-rolled) — route vocabulary through `humanize()`, render identifiers verbatim + mono (see "Rendered strings")
- **Do not** add a `Record<string, string>` display map whose entries just equal `humanize(key)`, put a lifecycle status in a `Badge` (use `Status`), or make a chip mono for vocabulary / sans for an identifier
