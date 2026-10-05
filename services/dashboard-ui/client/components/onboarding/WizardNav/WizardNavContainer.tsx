import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { WizardNav } from './WizardNav'

export const WizardNavContainer = ({
  isScrolled = false,
  skipHref,
  onSkip,
}: {
  isScrolled?: boolean
  skipHref?: string | null
  onSkip?: () => void
}) => {
  const { steps, currentStepIndex, completedSteps, goToStep } = useOnboardingWizard()

  return (
    <WizardNav
      isScrolled={isScrolled}
      steps={steps}
      currentStepIndex={currentStepIndex}
      completedSteps={completedSteps}
      showHeader
      skipHref={skipHref}
      onSkip={onSkip}
      onGoToStep={goToStep}
    />
  )
}
