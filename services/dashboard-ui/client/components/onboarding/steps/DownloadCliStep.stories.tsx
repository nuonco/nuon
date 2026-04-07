export default {
  title: 'Onboarding/V1 Steps/DownloadCliStep',
}

import { OnboardingJourneyContext } from '@/providers/onboarding-journey-provider'
import { DownloadCliStep } from './DownloadCliStep'

const mockProps = {
  onAdvance: () => {},
  onGoBack: () => {},
  isComplete: false,
  sharedData: {},
  setSharedData: () => {},
  nextStepTitle: 'Create app',
}

const mockJourney = {
  isLoading: false,
  orgId: 'org-1',
  isStepComplete: () => false,
  getStepMetadata: () => undefined,
}

export const Default = () => (
  <OnboardingJourneyContext.Provider value={mockJourney}>
    <DownloadCliStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)

export const CliInstalled = () => (
  <OnboardingJourneyContext.Provider value={{ ...mockJourney, isStepComplete: () => true }}>
    <DownloadCliStep {...mockProps} />
  </OnboardingJourneyContext.Provider>
)
