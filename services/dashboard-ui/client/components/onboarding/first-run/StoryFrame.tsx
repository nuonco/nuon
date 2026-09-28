import type { ReactNode } from 'react'
import { OnboardingWizardLayout } from '@/components/onboarding/OnboardingWizard/OnboardingWizard'
import type { TFirstRunStep } from '@/hooks/use-first-run-journey'
import { OnboardingWizardProvider } from '@/providers/onboarding-wizard-provider'
import type { TCloud, TPath } from './constants'
import { buildFirstRunSteps } from './steps'

// Stories only: the real wizard chrome (stepper, title, description) around a
// step's presentational view, so a story shows the step as users see it without
// the container's API calls.
export const StoryFrame = ({
  step,
  path = 'own',
  cloud = 'aws',
  children,
}: {
  step: TFirstRunStep
  path?: TPath
  cloud?: TCloud
  children: ReactNode
}) => {
  const steps = buildFirstRunSteps(path, cloud).map((def) =>
    def.id === step ? { ...def, component: () => <>{children}</> } : def
  )
  const index = Math.max(
    0,
    steps.findIndex((def) => def.id === step)
  )
  return (
    <OnboardingWizardProvider
      steps={steps}
      initialStepIndex={index}
      initialSharedData={{}}
      onComplete={() => {}}
    >
      <OnboardingWizardLayout onSkip={() => {}} />
    </OnboardingWizardProvider>
  )
}
