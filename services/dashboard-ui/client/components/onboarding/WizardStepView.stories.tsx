export default {
  title: 'Onboarding/WizardStepView',
}

import { WizardContext } from '@/providers/onboarding-wizard-provider'
import { OnboardingJourneyContext } from '@/providers/onboarding-journey-provider'
import { WizardStepView } from './WizardStepView'
import { Text } from '@/components/common/Text'

const MockStepComponent = ({ onAdvance }: any) => (
  <div className="flex flex-col gap-4">
    <Text variant="body">This is a mock step component.</Text>
    <button onClick={onAdvance} className="btn">Next</button>
  </div>
)

const mockSteps = [
  {
    id: 'step-1',
    title: 'Download the CLI',
    description: 'Install the Nuon CLI to get started.',
    component: MockStepComponent,
  },
  {
    id: 'step-2',
    title: 'Create your app',
    description: 'Set up your first application.',
    component: MockStepComponent,
  },
]

const mockWizard = {
  steps: mockSteps,
  currentStepIndex: 0,
  completedSteps: new Set<string>(),
  sharedData: {},
  canClose: true,
  markComplete: () => {},
  setSharedData: () => {},
  goToStep: () => {},
  goNext: () => {},
  goPrev: () => {},
  onComplete: () => {},
}

const mockJourney = {
  isLoading: false,
  orgId: 'org-1',
  isStepComplete: () => false,
  getStepMetadata: () => undefined,
}

export const FirstStep = () => (
  <WizardContext.Provider value={mockWizard}>
    <OnboardingJourneyContext.Provider value={mockJourney}>
      <div className="max-w-xl p-8">
        <WizardStepView />
      </div>
    </OnboardingJourneyContext.Provider>
  </WizardContext.Provider>
)

export const SecondStep = () => (
  <WizardContext.Provider value={{ ...mockWizard, currentStepIndex: 1 }}>
    <OnboardingJourneyContext.Provider value={mockJourney}>
      <div className="max-w-xl p-8">
        <WizardStepView />
      </div>
    </OnboardingJourneyContext.Provider>
  </WizardContext.Provider>
)
