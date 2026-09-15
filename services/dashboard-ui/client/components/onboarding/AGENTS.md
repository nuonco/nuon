# Onboarding Components

The signup wizard: a full-screen, step-based flow driven by a declarative array of step definitions. Read
[README.md](./README.md) for the full API.

Service-wide rules still apply: [services/dashboard-ui/AGENTS.md](../../../AGENTS.md).

## Layout

| Path | Role |
|------|------|
| `OnboardingWizard/` | Full-screen layout (nav + step view). `OnboardingWizardLayout` is the presentational shell. |
| `WizardNav/` | Stepper: dots, labels, animated progress, header with Docs/Skip. |
| `WizardStepView.tsx` | Renders the current step's title, description, and component; owns the transition animation. |
| `steps/` | Live v1 step components. |
| `steps/v2/` | Live v2 step components (current production flow). |
| `PlaygroundFlow.stories.tsx` | Fake, API-free flows for prototyping. Not imported by production code. |

State lives in `@/providers/onboarding-wizard-provider` (`steps`, `currentStepIndex`, `completedSteps`, `sharedData`).
Read it with `useOnboardingWizard()`. The live flows (`STEPS`, `STEPS_V2`) are defined in `@/views/Onboarding.tsx` and
switched by the `onboardingV2` runtime config flag.

## Prototyping a new onboarding experience

**Default to `PlaygroundFlow.stories.tsx`.** When someone asks to try, design, mock up, explore, or "vibe code" an
onboarding flow, screen, or step, build it there. Do not touch `views/Onboarding.tsx`, `steps/`, or `steps/v2/` unless
the request is explicitly to change what real users see.

Rules for playground work:

- No network. No `api()`, no `@/lib` imports, no `useQuery`/`useMutation`, no mocking libraries. Fake progress with
  `setTimeout` and local state.
- Every story ends with `.meta = { fullBleed: true }` or the wizard renders inside Ladle's padded canvas.
- A flow is an `IWizardStepDef[]` passed to the local `Playground` harness. Compose from the existing step consts
  (`HERO_STEP`, `NAME_STEP`, `STACK_STEP`, `CLOUD_STEP`, `STATUS_STEP`, `SUMMARY_STEP`) before writing new ones.
- Reuse the local helpers in that file: `NextButton`, `ChoiceCard`, `makeChoiceStep(sharedKey, choices)`.
- Keep everything in that one file. Do not add directories, config, or dependencies to prototype a flow.

A step component takes `IWizardStepComponentProps`:

```tsx
const MyStep = ({ sharedData, setSharedData, onAdvance, nextStepTitle }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-6">
    <Text variant="body" theme="neutral">Whatever the step asks for.</Text>
    <NextButton label={nextStepTitle} onClick={onAdvance} />
  </div>
)
```

`onAdvance()` marks the step complete and moves forward; on the last step it finishes the flow. `setSharedData(key, val)`
passes values to later steps. `onGoBack` is wired by the wizard, not the step.

## UI building blocks

Use components from `@/components/common/` — `Button`, `Card`, `Text`, `Badge`, `Icon`, `Banner`, `Divider`, `Expand`,
`CodeBlock`, `ClickToCopy`, `Tabs`, and form controls under `common/form/`. Read a component's `.stories.tsx` for its
props before using it.

- Icons: only `Icon` with a `variant` from the map in `common/Icon.tsx`. A variant not in that map renders nothing.
- Layout: `flex flex-col gap-*`, never `space-y-*`. Bare `border` / `divide-y` — no grey border color classes.
- Text sizing and color come from `Text` `variant` / `theme` props, not ad-hoc Tailwind.
- Copy style: [COPY_STYLE.md](../../../COPY_STYLE.md).

## Viewing and checking work

```bash
bun run dev:ladle                                             # http://localhost:61000 → Onboarding/Playground
bunx oxlint -c client/.oxlintrc.json client/components/onboarding/<file>
```

Ask the user to start Ladle rather than starting or restarting it yourself. Ladle's global providers (router, query
client, config, org, install, toasts, surfaces) come from `.ladle/components.tsx` — stories do not add their own.

## Changing the live flow

Production changes go through `views/Onboarding.tsx` (step arrays) plus the step components in `steps/` and `steps/v2/`.
Those steps call the real API and depend on the `onboarding` object in `sharedData`.

`AppProfileStep`, `CloudSetupStep`, `ProvisioningStep`, and `NextStepsStep` keep all their UI in the Container file
alongside their queries, so they have no presentational component to story in isolation.
