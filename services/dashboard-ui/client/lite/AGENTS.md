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
   is `Overview`**, built with `ComponentDocs`, filling `summary`, `use`,
   `avoid`, `rules` and a `props` entry for **every** prop in the interface.
   The `Overview` is the component's only documentation, and changing a
   component's props means updating it in the same change. A component without
   a complete `Overview` is not finished.
7. **No comments. Zero.** Not narrative, not explanatory, not JSDoc, not a
   "why" comment. The only exception is a tool directive (`eslint-*`,
   `oxlint-*`, `@ts-*`, `/// <reference>`, `prettier-ignore`). This is enforced
   by `comments.test.ts` — a comment fails the test suite.
8. **Never break a documented rule without explicit approval first.** If you
   conclude a rule in these documents has to be broken — a new UX pattern, a new
   tier, a bespoke component, a new token — **stop and do not build it.** Say
   plainly that you want to break a rule, name it, show which existing options
   you evaluated and why each fails, and offer the compliant alternative.
   "Sounds good" or silence is not approval. See DESIGN.md "Breaking these
   rules". These should be rare; reaching for one repeatedly means you are
   misreading the docs.

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
| Tree-wide invariant tests | `guardrails/` |
| Design tokens + global CSS | `styles.css` |
| Routes | `routes.tsx` |
| App entry | `LiteApp.tsx` |

A component imports from its own tier or below, never above. Unit tests are
colocated with their subject; only tree-wide guardrails live in `guardrails/`.

**Do not add a new top-level directory.** This layout is deliberate — put shared
code in the directory that already owns the concern, or raise it in review.

## The other documents

Read the relevant one before starting — they are not loaded automatically.

- **[DESIGN.md](./DESIGN.md)** — the five UX patterns and the atomic design
  method. Read before deciding **what** to build: which pattern a problem is,
  which tier a component belongs at, and which treatments already exist.
- **[STYLES.md](./STYLES.md)** — themes, tokens, type, spacing, elevation,
  motion, accessibility floor. Read before deciding **what it looks like**.
- **[DEV.md](./DEV.md)** — engineering conventions. Read before writing any
  component, hook, provider, page or test.
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
