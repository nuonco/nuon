import { useConfig } from '@/hooks/use-config'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import type { TOnboarding } from '@/types/ctl-api.types'
import { WizardNav } from './WizardNav'

export const WizardNavContainer = ({ isScrolled = false }: { isScrolled?: boolean }) => {
  const { steps, currentStepIndex, completedSteps, goToStep, sharedData } = useOnboardingWizard()
  const { onboardingV2 } = useConfig()

  const orgId = (sharedData.onboarding as TOnboarding | undefined)?.org_id
  const skipHref = orgId ? `/${orgId}` : null

  return (
    <WizardNav
      isScrolled={isScrolled}
      steps={steps}
      currentStepIndex={currentStepIndex}
      completedSteps={completedSteps}
      onboardingV2={!!onboardingV2}
      skipHref={skipHref}
      onGoToStep={goToStep}
    />
  )
}
