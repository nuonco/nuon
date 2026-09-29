import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import {
  OnboardingWizardProvider,
  type IOnboardingWizardProps,
} from '@/providers/onboarding-wizard-provider'
import { OnboardingWizardLayout } from './OnboardingWizard'
import { useWizardHistory } from './use-wizard-history'

function FirstRunSkipLayout({ onSkip }: { onSkip: (stepId: string) => void }) {
  const { steps, currentStepIndex } = useOnboardingWizard()
  const stepId = steps[currentStepIndex]?.id

  return (
    <OnboardingWizardLayout onSkip={stepId ? () => onSkip(stepId) : undefined} />
  )
}

function ExistingSkipLayout() {
  const { orgId } = useOnboardingJourney()
  const skipHref = orgId ? `/${orgId}/apps` : null

  return <OnboardingWizardLayout skipHref={skipHref} />
}

function ConnectedWizardLayout({ onSkip }: { onSkip?: (stepId: string) => void }) {
  if (onSkip) return <FirstRunSkipLayout onSkip={onSkip} />
  return <ExistingSkipLayout />
}

function WizardHistory({ onHistoryBack }: { onHistoryBack?: () => void }) {
  const { currentStepIndex, goPrev } = useOnboardingWizard()
  useWizardHistory(currentStepIndex, goPrev, onHistoryBack)
  return null
}

export function OnboardingWizardContainer({
  onSkip,
  onHistoryBack,
  ...props
}: IOnboardingWizardProps & {
  onSkip?: (stepId: string) => void
  onHistoryBack?: () => void
}) {
  return (
    <OnboardingWizardProvider {...props}>
      <WizardHistory onHistoryBack={onHistoryBack} />
      <ConnectedWizardLayout onSkip={onSkip} />
    </OnboardingWizardProvider>
  )
}
