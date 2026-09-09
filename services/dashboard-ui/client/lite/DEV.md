# Development guidelines

Engineering conventions for `client/lite/`. Visual and UX guidance lives in
[DESIGN.md](./DESIGN.md) and [STYLES.md](./STYLES.md), copy rules in
[COPY.md](./COPY.md), and per-component
rules in each component's `Overview` story.

## Directory layout

Each atomic tier is a directory, so a component has one obvious home:

| Tier | Directory |
|---|---|
| Pages | `pages/` |
| Templates | `components/templates/` |
| Organisms | `components/organisms/` |
| Molecules | `components/molecules/` |
| Atoms | `components/atoms/` |

Pages sit at the top level rather than under `components/` because they are what
the router mounts, not something other code composes.

**A component only ever imports from its own tier or below.** An atom importing
a molecule is a signal that the atom is really a molecule, or that the shared
part needs extracting downward. This is what keeps the dependency graph a DAG
and keeps atoms cheap to story and test.

This table is only where the files go. **[DESIGN.md](./DESIGN.md) owns what the
tiers mean** — what each may know about, and how to decide the tier of something
new. Read it before adding a component.

Alongside the tiers:

- `hooks/` — reusable behaviour. One hook per file, named `use-*.ts`.
- `providers/` — React context. One provider per file, named `*-provider.tsx`.
- `utils/` — pure helpers with no React import. One concern per file.
- `guardrails/` — tree-wide invariant tests. See "Tests" below.

**Do not add a new top-level directory.** The layout above is deliberate. If
something needs a shared home, put it in the directory that already owns that
concern — a query key belongs with the other query-key code in
`utils/list-query.ts`, not in a new `queries/`. If you are convinced a new home
is genuinely needed, raise it rather than creating it.

## Imports

**Within Lite, use relative paths.** Not `@/lite/*`. The `@/` alias resolves
from `client/`, so an `@/lite/...` import inside Lite reads as if it were
crossing an application boundary when it is not.

```ts
import { Text } from '../atoms/Text'
import { useOrg } from '../../../providers/org-provider'
```

**From outside Lite, the allowlist is small and closed.** These are the only
things Lite may import from the production tree:

| Import | Why it is shared |
|---|---|
| `@/lib` | The ctl-api client. There is one API surface, not two. |
| `@/types`, `@/types/ctl-api.types` | Generated API types. |
| `@/utils/*` | Pure, presentation-free helpers (`classnames`, `string-utils`, `status-utils`, `branch-utils`, `timeline-utils`). |
| `@/configs/*` | Static data tables such as cloud regions. |
| `@/providers/config-provider`, `@/hooks/use-config` | Runtime config injected by the Go BFF. Framework-level, not dashboard-level. |
| `@/lib/cookies` | Auth cookie access. |
| `@/lib/fixtures/*` | Story and test fixtures only. Never in shipped code. |

Everything else is off limits, and `no-restricted-imports` in
`client/.oxlintrc.json` enforces the two big ones: `@/components/**` and
dashboard providers/hooks. **If a lint rule blocks an import, build the thing in
Lite.** Do not add an override — every exception is a place the two apps
re-fuse, which is exactly what this tree exists to avoid.

A helper that is genuinely pure and useful to both apps belongs in `@/utils/`
and gets added to the table above. A helper that knows about Lite's components,
tokens or routes belongs in `client/lite/utils/`.

## Component anatomy

### Containers and presentational components

**Anything that fetches, polls or subscribes is a container with a
presentational sibling.** The container is `FooContainer`, the presentational
half is `Foo`, and they live in a directory together:

```
components/organisms/InstallsTable/
├── InstallsTable.tsx            presentational — props in, JSX out
├── InstallsTableContainer.tsx   queries, mutations, context
├── InstallsTable.stories.tsx
├── InstallsTable.test.tsx
└── index.ts
```

```ts
export { InstallsTableContainer as InstallsTable } from './InstallsTableContainer'
export { InstallsTable as InstallsTableComponent } from './InstallsTable'
```

The barrel exports the container under the plain name, because that is what
pages want. The presentational half is exported as `*Component` for stories and
tests.

**The split can happen at any tier.** A live-updating organism owns its own
query rather than having the page hoist it — otherwise pages become
god-components holding a dozen queries and nothing below them can be storied.

The falsifiable test: **if a component under `components/` cannot be rendered
from props alone, it needs a container split before it can have a story.**

Components with no data dependency stay as flat files (`atoms/Button.tsx`,
`molecules/Time.tsx`). Do not create a directory for a component that does not
need one.

**Never have both `Foo.tsx` and `Foo/` at the same level.** The flat file
shadows the directory's `index.ts` and import resolution silently picks the
wrong one.

### Naming

- `T` prefix for types and unions: `TButtonVariant`, `TLabelColors`.
- `I` prefix for props interfaces and config objects: `IButton`, `IInstallFilter`.
- A component's props interface is `I{ComponentName}` and is exported.

## Props

**A boolean prop is named for what it describes, with no `is` or `should`
prefix** — `loading`, `open`, `external`, `selected`, `fetching`, `expanded`,
`disabled`. Never `isLoading` / `isOpen` / `shouldPoll`.

When the boolean asserts that **something else exists**, `has` is that name:
`hasNext`, `hasMore`, `hasError`. This is not an exception to the rule — "a next
page exists" has no adjective form, so `hasNext` *is* what it is called. The
test is whether the boolean describes this component's own state (`loading`,
`open`) or the existence of something beyond it (`hasNext`).

Why unprefixed: **43 of Lite's 44 component interfaces extend a native HTML
attributes type**, so they already inherit `disabled`, `open`, `checked` and
`hidden` unprefixed — you cannot rename an inherited prop. Prefixing everything
would mean permanently exempting every native boolean across nearly every
component. Unprefixed is also the path of least resistance, because the obvious
name for a boolean is the adjective; the production dashboard has both `loading`
and `isLoading`, which is the evidence.

The exception is **destructured query results**, which are TanStack Query's
names and not ours:

```tsx
const { data, isLoading, isPlaceholderData, error } = useQuery({ ... })
return <InstallsTable loading={isLoading} fetching={isPlaceholderData} error={error} />
```

The container renames at the boundary. `isLoading` never appears in a props
interface we own.

**Props that mirror a native attribute keep the native name and native
meaning** — `disabled`, `readOnly`, `required`, `hidden`, `type` — and are
passed through, not reinvented.

**Always destructure every prop the component owns before spreading the rest
onto an element.** Unprefixed booleans share a namespace with real DOM
attributes, so a stray `loading` lands in the DOM as `loading="true"` on a
`<span>`.

```tsx
export const Text = ({ variant = 'body', loading, className, children, ...props }: IText) => (
  <span className={cn(VARIANT_CLASSES[variant], className)} {...props}>{children}</span>
)
```

**Variant props are string unions backed by a lookup record**, not conditional
class strings:

```tsx
const VARIANT_CLASSES: Record<TButtonVariant, string> = { primary: '...', ghost: '...' }
```

This makes the exhaustive set visible in one place and makes a missing variant a
type error.

## Loading

**Loading is a `loading` prop on the component, never a separate `*Skeleton`
component.** A hand-built skeleton is a second copy of the layout measured by
eye, and it drifts.

The skeleton must occupy exactly the box the loaded content will, and the way to
guarantee that is to make the skeleton the same kind of thing as the content
rather than a box sized to match it. `Text` does this by rendering its real
typographic classes around a zero-width space with a `ch` width — so the
skeleton inherits line-height and font-size for free. See `Text`'s `Overview`
story.

- Primitives take `loading` and `loadingWidth` (in `ch`).
- Collections use the loading state built into `Table` and `Timeline`.
- Chrome, labels and headings render real while their values load.
- A container distinguishes `loading` (first load, no data) from `fetching`
  (revalidating with data on screen) and passes both.
- Use `placeholderData: keepPreviousData` on list and detail queries so
  revisits skip the cold load entirely.

## Data fetching

TanStack Query, configured once in `LiteApp.tsx`: 30s `staleTime`, refetch on
window focus, and **no retry on 4xx**.

- **Query keys are arrays that start with the resource name and include every
  input that changes the result**, org id first:
  `['installs', orgId, ...list.queryKey]`.
- **`enabled` guards every query that depends on a route or context value**, and
  is the one place a non-null assertion is acceptable — `enabled: !!orgId` makes
  `orgId!` in the `queryFn` safe.
- **List state lives in the URL, not in component state.** `useListQueryState`
  owns search, offset and filters, reads and writes them as query parameters,
  and hands back a `queryKey` fragment. A list that keeps its filters in
  `useState` is not linkable and loses them on navigation.
- **Invalidate related queries in `onSuccess`** after any mutation that creates,
  updates or deletes.

### Defensive access

Treat every API response as potentially partial, whatever the generated types
claim. A single unguarded access crashes the page.

- Optional chaining on all nested API data: `install?.app_branch?.name`.
- Nullish coalescing for anything used in a comparison or arithmetic:
  `(step?.duration ?? 0) > 1000`.
- Guard before rendering children that require the data, rather than relying on
  the children to cope.

## Routing and pages

Routes are declared in `routes.tsx` as a React Router v7 object tree. Every
route has a stable `id`. `withPageTransitions` wraps each leaf element, so page
transitions are applied by the router rather than remembered per page.

**Redirect with a `loader` returning `redirect()`, never `<Navigate>`.**

**A page's job is to compose organisms and declare its own chrome.** Pages do
not fetch. A page that needs data either reads a provider or renders a
container.

Every page declares:

```tsx
usePageTitle('Installs')
useBreadcrumbs([
  { label: org?.name, href: orgId ? `/${orgId}` : undefined, loadingWidth: 16 },
  { label: 'Installs' },
])
```

Both are hooks rather than rendered components, so they work identically on
every branch of a page that returns early.

**Build URLs with the helpers in `utils/hrefs.ts`**, not template literals at
the call site. A route shape then changes in one file.

**Layout routes own providers and shells.** `OrgLayout` mounts `OrgProvider`,
`BreadcrumbProvider`, `StatusBarProvider` and `SurfaceHost`, then renders
`DashboardShell` around an `<Outlet />`. Child pages render bare content.

## Providers and hooks

- **A provider exports its context and a `use*` hook; consumers use the hook,
  never `useContext` directly.**
- **The hook throws when used outside its provider.** Fail loudly at the call
  site rather than returning `undefined` and crashing three components deeper.
- Provider values may be undefined while loading. Consumers use `org?.id`, never
  `org.id`.
- A hook that only wraps a single `useState` is not a hook. Hooks earn their
  file by encapsulating behaviour (`use-list-query-state`, `use-popover`,
  `use-focus-containment`).

## Surfaces and toasts

Modals and panels go through the surface system: `useSurfaces()` for
`openModal` / `openPanel` / `closeSurface`, `SurfaceHost` mounted by the layout
that scopes them. Surfaces are addressable by URL — `usePanelHref` and
`useModalHref` build links that open one — so a surface is linkable and
back-button-correct. **Never render a modal by holding an `open` boolean in a
page.**

Toasts go through `useToast()`. See [COPY.md](./COPY.md) for what they say.

## Stories and the Overview requirement

**Every component under `components/` has a `.stories.tsx` beside it**, titled
`lite/{tier}/{Component}`, whose **first export is `Overview`**.

The `Overview` story is the component's documentation. There is no other. These
markdown files hold conventions that span components; **what a single component
is for, when not to reach for it, and what every one of its props does lives in
its `Overview` and nowhere else** — so the rules render next to the thing they
describe, in the theme they describe it in, and go stale visibly rather than
quietly.

A component without a complete `Overview` is not finished. Treat it the way you
would treat a component that does not compile.

### The shape

`Overview` renders `ComponentDocs`
(`components/__stories__/ComponentDocs.tsx`). Every field below is required —
`ComponentDocs` types `use`, `avoid`, `rules` and `props` as optional so that
`sections` stays free-form, but a Lite component fills all of them.

```tsx
export default { title: 'lite/atoms/Badge' }

export const Overview = () => (
  <ComponentDocs
    name="Badge"
    tier="atom"
    summary="A small piece of metadata, as a single pill or a key/value label pair."
    use={[
      'Show metadata about a resource, such as a label, a count or a short classification.',
      'Pass labelKey with labelValue for a key/value label, which renders as one joined pill.',
    ]}
    avoid={[
      'Do not use a badge for resource state. That is Status.',
      'Do not make a badge clickable. Removal is the one exception, and it renders a real button inside.',
    ]}
    rules={[
      'Pass labelKey and labelValue rather than composing two Badges.',
      'A user-chosen colour applies to the value half only, and is ignored in the high contrast theme.',
    ]}
    props={[
      { name: 'tone', type: "'neutral' | 'accent'", default: "'neutral'", description: 'Semantic tone.' },
      { name: 'onRemove', type: '() => void', description: 'Renders a remove button inside the badge.' },
    ]}
  />
)
```

`Badge`, `Text`, `Button` and `Status` are the reference implementations. Read
one before writing a new one.

### The bar for each field

- **`summary`** — one sentence naming what the thing *is*, not what it renders.
  "A small piece of metadata, as a single pill or a key/value label pair," not
  "Renders a span with rounded corners."
- **`use`** — the cases this component is the right answer for, including its
  non-obvious modes. If a prop combination unlocks a distinct use, say so here.
- **`avoid`** — **the highest-value field, and the one that actually prevents
  mistakes.** Each entry names the wrong reach and points at the right
  component: "Do not use a badge for resource state. That is Status." An
  `avoid` list that only says "don't overuse it" is not doing the job. If you
  cannot think of a way to misuse the component, you have not looked at the
  components it sits next to.
- **`rules`** — the constraints a correct call site must respect: what must be
  passed together, what is ignored in which theme, what truncates, what the
  component does to its children.
- **`props`** — **every prop in the exported interface, with no exceptions.**
  Including `className`, `children` and pass-throughs; if a prop is a plain
  pass-through, that is a one-line description, not a reason to omit it. Give
  the real type as written in the interface, and a `default` wherever the
  component defaults it.
- **`sections`** — optional, free-form `ReactNode`. Use it when a component has
  a concept that needs showing rather than listing — Button's emphasis ladder,
  Text's loading measurement, Diff's rendering modes.

### Keeping it true

**Changing a component's props means changing its `Overview` in the same
change.** This is where the documentation rots: the component grows a prop, the
props table does not, and a month later the table is a lie that costs more than
no table at all. If you add, rename, retype or re-default a prop, the `Overview`
edit is part of that work, not follow-up.

The same applies to behaviour. Widening what a component accepts usually means a
new `use` entry; narrowing it usually means a new `avoid` entry.

### What is enforced

`components/__stories__/overview-docs.test.ts` checks all of this on every
`bun test` run. It parses each component's exported `I{Component}` interface and
each `Overview`'s `props={[...]}` and fails when they disagree. Specifically:

- every component exporting a props interface is covered by an `Overview`,
- `Overview` is the first export and renders `ComponentDocs`,
- `summary`, `use`, `avoid` and `rules` are all non-empty,
- every prop in the interface has a `props` entry,
- no `props` entry names a prop that no longer exists.

The stale-prop check is skipped for components whose props come from `extends`
or an imported type, because that is not resolvable without a type checker.
Internal parts that are not standalone components — `SurfaceOverlay`,
`SurfaceTransition`, `ToastStack` — are listed in `INTERNAL_COMPONENTS`.

**`KNOWN_GAPS` is the existing debt, and it only shrinks.** It lists props that
predate this test. Documenting one and leaving it in the list fails, and so does
listing a prop the component no longer has — so an entry cannot go stale and the
debt cannot quietly reopen. **Nothing new goes in it.** A prop added today is
documented today; if you find yourself adding a `KNOWN_GAPS` entry, write the
props entry instead.

### The remaining stories

Everything after `Overview` is a plain function component demonstrating a real
state, named for the state it shows — `Tones`, `Variants`, `Loading`, `Empty`,
`Error`, `Truncation`. Between them they should cover every visual branch the
component has, because these are what a reviewer scrolls to check a change in
all three themes. A component with a `loading` prop and no loading story is
under-storied.

Rules:

- **Ladle v5 format — plain function exports only.** `StoryObj` with `render:`
  is Storybook syntax and fails with "got: object".
- **Stories render the presentational half**, never the container.
- **Ladle provides a `MemoryRouter` globally.** Wrapping a story in another
  router throws.
- Pages are not in Ladle. A template is already the presentational page, so
  storying a page would mean splitting it just to have something to render.
- Story infrastructure in `__stories__/` and fixtures in `__fixtures__/` are
  outside the glob.

Lite has its own Ladle instance (`bun run dev:ladle:lite`, port 61001, config in
`.ladle-lite/`) because it must load `client/lite/styles.css` and *not* the
production stylesheet — a story rendered inside the old global CSS would not
match what ships. Ladle's light/dark/auto control is wired to the Lite theme
preference; high contrast is reached through a `ThemeSwitcher` rendered in the
story.

## Tests

`bun test` with `@testing-library/react`. Lite has two kinds of test, and which
kind it is decides where it goes.

### Unit tests — colocated

The default. `Foo.test.tsx` sits next to `Foo.tsx`, `foo.test.ts` next to
`foo.ts`. They test one subject through its own public surface.

`routes.test.tsx` is colocated too — its subject is `routes.tsx`, which lives at
the root.

### Guardrails — `guardrails/`

A guardrail asserts a **rule holding across the whole tree** rather than a unit
behaving. It has no single subject, so there is nothing to sit beside:

| File | What it holds |
|---|---|
| `comments.test.ts` | No source file carries a comment. |
| `overview-docs.test.ts` | Every component has a complete `Overview` story. |
| `view-transitions.test.tsx` | Every navigation path opts into `viewTransition`. |

Two of these are really lint rules that run in `bun test` because there is no
lint rule that can express them, which is the other reason to keep them out of
the unit tests — a failure means "you broke a convention", not "you broke the
code", and the message should read that way.

Add a guardrail when a convention in these documents is one an agent keeps
breaking and a script can check it. A convention nothing checks is a suggestion.

### Conventions

```tsx
import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
```

- **Test the presentational component with fixture props**, not the container.
- **Query by role and accessible name.** A test that reaches for a class name is
  testing the stylesheet.
- `cleanup()` in `afterEach`.
- Pure logic — diff parsing, query-parameter codecs, href builders, session
  storage — is tested directly in `utils/`, where it is cheapest.

What is worth a test: parsing and formatting, query-state round-trips,
keyboard interaction, conditional rendering with real consequences. Not: that a
component renders its children.

## Comments

**Lite source files contain no comments. Zero.** Not in components, hooks,
providers, utils, tests, stories or type files. Not narrative comments, not
explanatory ones, not JSDoc, not section banners, and not "why" comments.

The only exception is a tool directive: `eslint-*`, `oxlint-*`, `@ts-*`,
`/// <reference>`, `prettier-ignore`.

This is not a style preference to weigh against other concerns. It is absolute,
and `comments.test.ts` fails the suite on any comment it finds — reporting the
file, line and text.

### What to do instead

A comment is almost always a naming failure wearing a disguise. When you feel
the urge to write one:

- **Rename.** `const d = ...` needing `// duration in ms` wants to be
  `durationMs`. `if (x > 3)` needing `// max retries` wants `MAX_RETRIES`.
- **Extract a named function.** A block that needs a comment to say what it does
  wants to be a function whose name says it. `skipTemplate`, `filterControl`,
  `resolveTimeout` — each replaces a paragraph.
- **Extract a named constant.** A magic value that needs explaining wants a name.
- **Put it in the `Overview` story.** Guidance about how a component is used, its
  constraints, its gotchas — that is exactly what `use`, `avoid` and `rules` are
  for, and it renders where people will actually read it.
- **Put it in these documents.** A convention that spans components belongs in
  DEV.md, DESIGN.md or STYLES.md, not in a comment in one file.

The reason the rule is absolute rather than "no *bad* comments" is that the
carve-out is what kills it. Every comment its author writes feels like the
justified exception, "why" is trivially claimable for anything, and nothing can
check the difference — so a soft rule decays into the narrated code it was meant
to prevent. A rule a test can enforce is worth more than a better rule it
cannot.

## Commands

```bash
bun run dev:ladle:lite                             # stories, port 61001
bun test client/lite                               # unit tests
bunx oxlint -c client/.oxlintrc.json client/lite   # lint
bunx tsc --noEmit --project client/tsconfig.json   # type check
```

Do not run production builds and do not start, stop or restart the dev stack.
