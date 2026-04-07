export default {
  title: 'Onboarding/V1 Steps/CreateAppStep',
}

import { OnboardingJourneyContext } from '@/providers/onboarding-journey-provider'
import { CreateAppStep } from './CreateAppStep'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {},
  setSharedData: () => {},
  nextStepTitle: 'Sync app',
}

const mockJourney = {
  isLoading: false,
  orgId: 'org-1',
  isStepComplete: () => false,
  getStepMetadata: () => undefined,
}

export const Default = () => (
  <OnboardingJourneyContext.Provider value={mockJourney}>
    <CreateAppStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)

export const StepComplete = () => (
  <OnboardingJourneyContext.Provider value={{ ...mockJourney, isStepComplete: () => true }}>
    <CreateAppStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)
