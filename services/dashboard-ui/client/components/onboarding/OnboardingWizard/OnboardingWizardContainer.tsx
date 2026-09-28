import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import {
  OnboardingWizardProvider,
  type IOnboardingWizardProps,
} from '@/providers/onboarding-wizard-provider'
import { OnboardingWizardLayout } from './OnboardingWizard'

function ConnectedWizardLayout({ onSkip }: { onSkip?: (stepId: string) => void }) {
  const { steps, currentStepIndex } = useOnboardingWizard()
  const stepId = steps[currentStepIndex]?.id

  return (
    <OnboardingWizardLayout onSkip={onSkip && stepId ? () => onSkip(stepId) : undefined} />
  )
}

export function OnboardingWizardContainer({
  onSkip,
  ...props
}: IOnboardingWizardProps & { onSkip?: (stepId: string) => void }) {
  return (
    <OnboardingWizardProvider {...props}>
      <ConnectedWizardLayout onSkip={onSkip} />
    </OnboardingWizardProvider>
  )
}
