import { useContext } from 'react'
import { OnboardingJourneyContext } from '@/providers/onboarding-journey-provider'

export function useOnboardingJourney() {
  const context = useContext(OnboardingJourneyContext)
  if (context === undefined) {
    throw new Error('useOnboardingJourney must be used within an OnboardingJourneyProvider')
  }
  return context
}
