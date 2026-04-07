export default {
  title: 'Onboarding/V1 Steps/SyncAppStep',
}

import { OnboardingJourneyContext } from '@/providers/onboarding-journey-provider'
import { SyncAppStep } from './SyncAppStep'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {},
  setSharedData: () => {},
  nextStepTitle: 'Create install',
}

const mockJourney = {
  isLoading: false,
  orgId: 'org-1',
  isStepComplete: () => false,
  getStepMetadata: () => undefined,
}

export const Default = () => (
  <OnboardingJourneyContext.Provider value={mockJourney}>
    <SyncAppStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)

export const AppSynced = () => (
  <OnboardingJourneyContext.Provider value={{ ...mockJourney, isStepComplete: () => true }}>
    <SyncAppStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)
