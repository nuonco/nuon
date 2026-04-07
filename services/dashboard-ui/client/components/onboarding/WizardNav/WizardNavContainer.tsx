import { useConfig } from '@/hooks/use-config'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { WizardNav } from './WizardNav'

export const WizardNavContainer = ({ isScrolled = false }: { isScrolled?: boolean }) => {
  const { steps, currentStepIndex, completedSteps, goToStep } = useOnboardingWizard()
  const { onboardingV2 } = useConfig()

  return (
    <WizardNav
      isScrolled={isScrolled}
      steps={steps}
      currentStepIndex={currentStepIndex}
      completedSteps={completedSteps}
      onboardingV2={!!onboardingV2}
      onGoToStep={goToStep}
    />
  )
}
