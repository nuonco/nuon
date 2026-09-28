# Onboarding

Everyone who reaches `/onboarding` gets the first-run flow: an intro, then a wizard that creates an app template and a
first test install in the user's own cloud account. The flow lives in [`first-run/`](./first-run) and runs on the
wizard shell in this directory.

## Files

| Path | Purpose |
|------|---------|
| `first-run/IntroScreen.tsx` | Full-screen intro before the stepper. No network calls. |
| `first-run/StartStep.tsx` | Picks the path. Own app: GitHub, name, test cloud, then app + branch + branch config. Example: Kitchen Sink. |
| `first-run/ConnectStep.tsx` | Own path only. Agent prompt from `nuon.co/loop.md`, push detection, fine print. |
| `first-run/DeployStep.tsx` | Region and auto-approve, waits for an active app config, creates the install. |
| `first-run/StackStep.tsx` | Polls the install stack: AWS quick-create link, GCP Terraform, Azure `az` commands. |
| `first-run/ProvisionStep.tsx` | Static preview of what the install builds. Finish opens the deploy workflow. |
| `first-run/constants.ts` | Clouds, regions, sandbox repos, config stubs, stack copy. |
| `first-run/api.ts` | Multi-call helpers: set up the app and branch, find the active config, create the install. |
| `first-run/resume.ts` | Resolves the org (creating a trial org for a new sign-up) and where to resume. |
| `first-run/steps.tsx` | Step definitions and `buildFirstRunSteps(path, cloud)`. |
| `first-run/*.stories.tsx` | One story file per step, built on the presentational `*View` components. |
| `OnboardingWizard/` | Full-screen layout (nav + step view). `OnboardingWizardLayout` is the presentational shell. |
| `WizardNav/` | Stepper and header (Docs, Skip). Also used by `RunRunbookForm`. |
| `WizardStepView.tsx` | Renders the current step's title, description, and component; owns the transition. |

The route view is `@/views/Onboarding.tsx`. The dashboard server sends new sign-ups (no org) and anyone with an
unfinished `first_run` journey from `/` to `/onboarding`, and returns GitHub App callbacks there when the `state` is
`<orgID>:onboarding` (`server/internal/handlers/root.go`, `connect.go`).

## Paths

| Path | Steps |
|------|-------|
| Own app | Start, Connect, Deploy, Stack, Provision |
| Example app | Start, Deploy, Stack, Provision |

The stepper follows the path: `choosePath(path, cloud)` from `useFirstRun()` swaps the step array. Step IDs are the
journey's step names, so completion and resume map one to one.

## The `first_run` journey

Progress is stored on the account as the `first_run` user journey (`@/hooks/use-first-run-journey`). The view creates
it on the first visit (a 409 means another tab already did) with the steps `start`, `connect`, `deploy`, `stack`,
`provision`. The example path marks `connect` complete along with `start`.

It is created before the org on purpose: ctl-api's journey create saves the whole account and fails once the account
holds a role. When it cannot be created (an account that already belongs to an org), the flow still runs and nothing
is saved.

Step metadata:

| Key | Written by | Value |
|-----|------------|-------|
| `path` | start | `own` or `example` |
| `cloud` | start | `aws`, `gcp`, or `azure` |
| `app_name` | start | App template name as typed |
| `app_id` | start | App ID |
| `app_branch_id` | start | The `default` branch ID |
| `repo` | start | `owner/repo` the branch tracks |
| `region` | deploy | Region or Azure location |
| `install_id` | deploy | Install ID |
| `workflow_id` | deploy | The install's provision workflow ID |
| `skipped` | any | `"true"` on the step the user skipped from |

Metadata saves resend the step's current `complete` value, because the step PATCH sets `complete` to whatever it is
sent. Resume opens the first incomplete step with the saved data, and starts over at Start if a saved app or install
returns 404.

- **Skip** saves `skipped` on the current step, completes the journey, and goes to `/<orgID>`.
- **Finish** completes the journey and opens `/<orgID>/installs/<install_id>/history/<workflow_id>`.
- **Re-open onboarding** (user menu) resets the journey, clears the metadata keys, and opens `/onboarding?reopen=1`.

## What each step calls

| Step | Calls |
|------|-------|
| View | `GET /v1/account`, `POST /v1/account/user-journeys`, `GET /v1/orgs`, `/api/random-name` + `POST /v1/orgs` (409 retried up to 5 times) |
| Start, own | `GET /v1/vcs/connections`, `GET /v1/vcs/connections/:id/repos`, `POST /v1/apps`, `PATCH /v1/apps/:id` (`config_repo`, `config_directory`), `POST /v1/apps/:id/branches`, `POST .../branches/:id/configs` |
| Start, example | `GET /v1/apps?q=`, `POST /v1/apps`, the same branch calls with `public_git_vcs_config` |
| Connect | `GET https://nuon.co/loop.md`, `GET .../branches/:id/runs` every 5 seconds |
| Deploy | `GET /v1/apps/:id/configs`, `GET /v1/apps/:id/configs/:config_id?recurse=true`, `POST /v1/apps/:id/installs` |
| Stack | `GET /v1/installs/:id/stack` every 3 seconds |
| Provision | None until Finish |

Kitchen Sink's repo root is AWS-only, so the example path uses `kitchen-sink` (directory `.`) for AWS and
`kitchen-sink-gcp` (directory `gcp`) for GCP.

## Wizard API

```tsx
<OnboardingWizard
  steps={steps}                 // IWizardStepDef[]
  initialStepIndex={0}          // required
  initialSharedData={{}}        // required
  onComplete={() => {}}         // advancing past the last step
  onSkip={(stepId) => {}}       // optional; shows Skip in the header
/>
```

The provider never reads or writes browser storage; callers pass where to start.

A step component receives `IWizardStepComponentProps`:

```ts
interface IWizardStepComponentProps {
  isComplete: boolean
  sharedData: Record<string, unknown>
  setSharedData: (key: string, val: unknown) => void
  onAdvance: () => void        // marks the step complete and moves forward
  onGoBack?: () => void        // set on every step after the first
  nextStepTitle?: string
}
```

`useOnboardingWizard()` exposes `steps`, `currentStepIndex`, `completedSteps`, `sharedData`, `markComplete`,
`setSharedData`, `goToStep`, `goNext`, and `goPrev` for steps that need more than `onAdvance`.

## Viewing it

```bash
bun run dev:ladle   # http://localhost:61000 → Onboarding / First run
```

To run the live flow locally, see [services/dashboard-ui/AGENTS.md](../../../AGENTS.md#first-run-onboarding-locally).
