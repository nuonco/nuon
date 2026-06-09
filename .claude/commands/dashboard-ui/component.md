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

## Icons

Use the `Icon` component from `@/components/common/Icon` for ALL icons. Always use the `Icon` suffix for variant names (e.g., `HouseIcon` not `House`).

If you need a Phosphor icon that isn't already available, add it to `client/components/common/Icon.tsx`:
1. Add the named import: `import { NewIconNameIcon } from '@phosphor-icons/react'`
2. Add it to the `phosphorIcons` object: `NewIconNameIcon,`

Never import directly from `@phosphor-icons/react`, `lucide-react`, or `heroicons` in component files.

## Anti-Patterns

- **Do not** create a component that duplicates an existing one — always check existing components first
- **Do not** pass props to a component without reading its interface — wrong props cause runtime errors
- **Do not** use `ModalBase` or `PanelBase` directly — always use the `Modal`/`Panel` wrappers from `surfaces/`
- **Do not** put a domain-specific component (e.g., `InstallCard`) into `client/components/common/`
- **Do not** leave both a flat file and directory with the same name — delete the flat file after migrating to the directory pattern
- **Do not** skip the `.stories.tsx` file — every component directory must have one
- **Do not** use `StoryObj` or `render:` in stories — Ladle v5 requires plain function exports
- **Do not** import icons directly from `@phosphor-icons/react` — always use the `Icon` component
