# Lite dashboard

`client/lite/` is a rebuild of the Nuon dashboard around a smaller surface area.
It is a self-contained application: its own routes, shell, providers, hooks,
component library, stylesheet and Ladle instance. It shares only the API client
and a handful of framework-level utilities with the production SPA.

## This file supersedes the parent AGENTS.md

`services/dashboard-ui/AGENTS.md` documents the **production** dashboard in
`client/components/` and `client/views/`. Its conventions do not apply here and
several of them directly contradict this tree — Lite has no `PageLayout`,
`ListPage`, `DetailPage`, `SectionHeader` or `common/` directory, uses a
different naming convention for boolean props, and uses a different design
token set.

**When working anywhere under `client/lite/`, follow these documents and ignore
the parent AGENTS.md.** When working anywhere else in the repo, ignore this one.

## Isolation, in both directions

- **Lite must not import from `client/components/`.** Those components assume
  the production stylesheet and use `dark:` variants that do not exist here.
- **Production code must not import from `client/lite/`.** Lite is not a shared
  library and its API is not stable.
- Both directions are enforced by `no-restricted-imports` in
  `client/.oxlintrc.json`. If a rule blocks you, the answer is to build the
  thing in Lite, not to add an exception.

The shared surface is deliberately small — see [DEV.md](./DEV.md) "Imports" for
the exact allowlist.

## Hard rules

These are the ones that get broken most often. Full reasoning in the linked docs.

1. **No `dark:` variants, no raw colour values, no stock Tailwind colour
   utilities.** Semantic tokens only (`bg-surface-default`, `text-secondary`,
   `border-divider`). Needing a `dark:` variant means a token is missing.
2. **Boolean props are unprefixed** — `loading`, `open`, `selected`, `external`.
   Never `isLoading` / `isOpen` / `hasError`.
3. **Destructure every prop the component owns before spreading the rest onto a
   DOM element.** Our prop names share a namespace with real HTML attributes.
4. **Loading is a `loading` prop on the component, never a separate `*Skeleton`
   component.**
5. **Anything that fetches, polls or subscribes is a `*Container` with a
   presentational sibling.** The presentational half must render from props
   alone.
6. **Every component under `components/` has a `.stories.tsx` whose first export
   is `Overview`**, built with `ComponentDocs`.
7. **No comments** unless they explain a non-obvious *why*. Never narrate what
   the code does.

## Where things live

| What | Where |
|---|---|
| Pages (what the router mounts) | `pages/` |
| Templates | `components/templates/` |
| Organisms | `components/organisms/` |
| Molecules | `components/molecules/` |
| Atoms | `components/atoms/` |
| Hooks | `hooks/` |
| Providers | `providers/` |
| Pure helpers | `utils/` |
| Design tokens + global CSS | `styles.css` |
| Routes | `routes.tsx` |
| App entry | `LiteApp.tsx` |

A component imports from its own tier or below, never above.

## The other documents

Read the relevant one before starting — they are not loaded automatically.

- **[DEV.md](./DEV.md)** — engineering conventions. Read before writing any
  component, hook, provider, page or test.
- **[DESIGN.md](./DESIGN.md)** — themes, tokens, spacing, UX patterns, component
  emphasis. Read before any visual work.
- **[COPY.md](./COPY.md)** — voice and copy rules for all user-facing text. Read
  before writing a label, heading, empty state, error or toast.
- **[FLOWS.md](./FLOWS.md)** — the user-facing flows the app implements. Read
  before building or changing a multi-step flow such as a setup wizard.
- Per-component rules live in that component's `Overview` story, not in these
  documents.

## Commands

```bash
bun run dev:ladle:lite   # Lite component stories, port 61001
bun test client/lite     # unit tests
bunx oxlint -c client/.oxlintrc.json client/lite
```

Do not start, stop or restart the dev stack — ask for it to be run instead.
