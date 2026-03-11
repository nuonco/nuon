import { OnboardingWizardProvider, type IOnboardingWizardProps } from '@/providers/onboarding-wizard-provider'
import { WizardNav } from './WizardNav'
import { WizardStepView } from './WizardStepView'

export function OnboardingWizard(props: IOnboardingWizardProps) {
  return (
    <OnboardingWizardProvider {...props}>
      <div className="h-screen flex flex-col bg-background">
        <WizardNav />
        <div className="flex-1 overflow-y-auto px-8 pt-14 pb-8">
          <div className="max-w-2xl mx-auto">
            <WizardStepView />
          </div>
        </div>
      </div>
    </OnboardingWizardProvider>
  )
}
