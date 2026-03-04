import { useContext } from 'react'
import { WizardContext, type IWizardContext } from '@/providers/onboarding-wizard-provider'

export function useOnboardingWizard(): IWizardContext {
  const ctx = useContext(WizardContext)
  if (!ctx) {
    throw new Error('useOnboardingWizard must be used within an OnboardingWizardProvider')
  }
  return ctx
}
