# Onboarding Wizard

A full-screen, step-based wizard with a top nav bar, animated step transitions, and shared state across steps. The flow is driven by a declarative array of step definitions, making it easy to swap, extend, or A/B test flows.

## Files

| File | Purpose |
|------|---------|
| `OnboardingWizard.tsx` | Main entry point — renders the full-screen layout |
| `WizardNav.tsx` | Top nav bar with step dots, prev/next arrows, and close button |
| `WizardStepView.tsx` | Renders the current step's heading and component |
| `../../providers/onboarding-wizard-provider.tsx` | Context, provider, and all shared types |
| `../../hooks/use-onboarding-wizard.ts` | `useOnboardingWizard()` hook for accessing wizard state |

---

## Basic Usage

```tsx
import { OnboardingWizard } from '@/components/onboarding/OnboardingWizard'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const WelcomeStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <div>
    <p>Welcome! Let's get started.</p>
    <button onClick={onAdvance}>Continue</button>
  </div>
)

const STEPS = [
  {
    id: 'welcome',
    title: 'Welcome',
    description: 'A quick intro.',
    component: WelcomeStep,
  },
]

export function MyFlow() {
  return (
    <OnboardingWizard
      steps={STEPS}
      onComplete={() => { window.location.href = '/dashboard' }}
    />
  )
}
```

---

## `OnboardingWizard` Props

```ts
interface IOnboardingWizardProps {
  steps: IWizardStepDef[]   // Ordered list of step definitions
  onComplete: () => void     // Called when the final step is advanced past
  canClose?: boolean         // Show a "Close" button in the nav (default: false)
  onClose?: () => void       // Called when Close is clicked
}
```

---

## Defining Steps

Each step is a plain object:

```ts
interface IWizardStepDef {
  id: string                                        // Unique identifier
  title: string                                     // Shown as heading and in dot tooltip
  description?: string                              // Shown below the heading
  component: ComponentType<IWizardStepComponentProps>
  data?: unknown                                    // Optional static config passed to the component
}
```

---

## Step Component Contract

Every step component receives these props:

```ts
interface IWizardStepComponentProps {
  isComplete: boolean                               // Whether this step has been marked complete
  sharedData: Record<string, unknown>               // Cross-step shared state
  setSharedData: (key: string, val: unknown) => void
  onAdvance: () => void                             // Marks step complete and advances to next
}
```

### Simple step

```tsx
const MyStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <button onClick={onAdvance}>Done</button>
)
```

### Async step (API call before advancing)

```tsx
const CreateOrgStep = ({ onAdvance, setSharedData }: IWizardStepComponentProps) => {
  const handleSubmit = async () => {
    const org = await createOrg({ name: 'Acme' })
    setSharedData('orgId', org.id)  // available to all later steps
    onAdvance()
  }

  return <button onClick={handleSubmit}>Create Org</button>
}
```

### Reading shared data in a later step

```tsx
const InstallStep = ({ sharedData, onAdvance }: IWizardStepComponentProps) => {
  const orgId = sharedData.orgId as string

  return <div>Creating install for org {orgId}…</div>
}
```

---

## Navigation Rules

- **Next arrow** is disabled until the current step calls `onAdvance()`.
- **Back arrow** is always enabled (except on step 1).
- **Step dots** are clickable for the current step and any already-completed steps.
- Completed dots show a check icon; the active dot is highlighted in primary blue.

---

## Accessing Wizard State Inside a Step

If you need more control than `onAdvance` provides (e.g. mark complete without immediately advancing, or jump to a specific step), call `useOnboardingWizard()` from within the step component:

```tsx
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'

const PollingStep = ({ isComplete }: IWizardStepComponentProps) => {
  const { markComplete, goNext, steps, currentStepIndex } = useOnboardingWizard()
  const stepId = steps[currentStepIndex].id

  useEffect(() => {
    if (deploymentReady) {
      markComplete(stepId)
      // Let the user click next manually, or call goNext() here
    }
  }, [deploymentReady])

  return isComplete ? <p>Ready!</p> : <p>Waiting…</p>
}
```

### Available from `useOnboardingWizard()`

| Value | Type | Description |
|-------|------|-------------|
| `steps` | `IWizardStepDef[]` | All step definitions |
| `currentStepIndex` | `number` | Index of the active step |
| `completedSteps` | `Set<string>` | IDs of completed steps |
| `sharedData` | `Record<string, unknown>` | Cross-step shared state |
| `canClose` | `boolean` | Whether the close button is shown |
| `markComplete` | `(id: string) => void` | Mark a step complete without advancing |
| `setSharedData` | `(key, val) => void` | Write to shared state |
| `goToStep` | `(index: number) => void` | Jump to a specific step |
| `goNext` | `() => void` | Advance one step (or call `onComplete` on the last) |
| `goPrev` | `() => void` | Go back one step |
| `onComplete` | `() => void` | The completion callback passed to `OnboardingWizard` |
| `onClose` | `() => void \| undefined` | The close callback passed to `OnboardingWizard` |
