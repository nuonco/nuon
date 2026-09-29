# Onboarding Components

`/onboarding` serves the existing wizard in `steps/` (welcome, org, CLI, app, sync, install). Set
`NUON_ONBOARDING_FIRST_RUN=true` on the dashboard server to serve the first-run flow in `first-run/` instead.
Read [README.md](./README.md).

Service-wide rules still apply: [services/dashboard-ui/AGENTS.md](../../../AGENTS.md).

## Layout

| Path | Role |
|------|------|
| `steps/` | The default onboarding steps. |
| `first-run/` | The first-run flow, served when `NUON_ONBOARDING_FIRST_RUN` is set. |
| `OnboardingWizard/` | Full-screen layout (nav + step view). `OnboardingWizardLayout` is the presentational shell. |
| `WizardNav/` | Stepper: dots, labels, animated progress, header with Docs and Skip. |
| `WizardStepView.tsx` | Renders the current step's title, description, and component; owns the transition animation. |

Wizard state lives in `@/providers/onboarding-wizard-provider`; read it with `useOnboardingWizard()`. First-run
context (org ID, journey helpers, `choosePath`, `backToIntro`) lives in `@/providers/first-run-provider`; read it with
`useFirstRun()`. The route view is `@/views/Onboarding.tsx`.

## Changing a first-run step

- Each step file exports a presentational `*StepView` (props only, no queries) and a `*Step` container that the wizard
  renders. Stories use the view with mock props inside `StoryFrame`, which draws the real wizard chrome.
- Anything a later step or a reload needs goes in both `sharedData` and the journey (`journey.saveStep`). Add new keys
  to `FIRST_RUN_METADATA_KEYS` so Re-open onboarding clears them.
- Create calls must be safe to repeat: Back then Next, a reload, and a second tab all rerun them. Look for the
  existing resource before creating one (see `ensureDefaultBranch`, `setUpKitchenSink`).
- Cloud tables, copy that varies by cloud, and config stubs belong in `constants.ts`, not in step files.

## UI building blocks

Use components from `@/components/common/`: `Button`, `Card`, `Text`, `Badge`, `Icon`, `Banner`, `CodeBlock`,
`Link`, and form controls under `common/form/`. Read a component's `.stories.tsx` for its props before using it.

- Icons: only `Icon` with a `variant` from the map in `common/Icon.tsx`. A variant not in that map renders nothing.
- Layout: `flex flex-col gap-*`, never `space-y-*`. Bare `border` / `divide-y`; no grey border color classes.
- Text sizing and color come from `Text` `variant` / `theme` props, not ad-hoc Tailwind.
- Copy: [COPY_STYLE.md](../../../COPY_STYLE.md). In this flow the account is always a "test" account, the thing the
  user creates is an "app template", and no copy uses an em dash.

## Viewing and checking work

```bash
bun run dev:ladle                                             # http://localhost:61000 → Onboarding/First run
bunx oxlint -c client/.oxlintrc.json client/components/onboarding/<file>
bun test client/components/onboarding                         # config stub snapshots
```

Ask the user to start Ladle rather than starting or restarting it yourself. Ladle's global providers (router, query
client, config, org, install, toasts, surfaces) come from `.ladle/components.tsx`; stories do not add their own.

End-to-end coverage is in `e2e/specs/onboarding/` ([flow](../../../e2e/flows/onboarding.flow.md)).
