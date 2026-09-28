import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { WizardNav } from './WizardNav'

export const WizardNavContainer = ({
  isScrolled = false,
  onSkip,
}: {
  isScrolled?: boolean
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
      onSkip={onSkip}
      onGoToStep={goToStep}
    />
  )
}
