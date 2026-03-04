import { OnboardingWizardProvider, type IOnboardingWizardProps } from '@/providers/onboarding-wizard-provider'
import { WizardNav } from './WizardNav'
import { WizardStepView } from './WizardStepView'

export function OnboardingWizard(props: IOnboardingWizardProps) {
  return (
    <OnboardingWizardProvider {...props}>
      <div className="h-screen flex flex-col bg-background">
        <WizardNav />
        <div className="flex-1 overflow-y-auto p-8">
          <div className="max-w-4xl mx-auto">
            <WizardStepView />
          </div>
        </div>
      </div>
    </OnboardingWizardProvider>
  )
}
