'use client'

import { useMemo } from 'react'
import type { TAccount, TUserJourney } from '@/types'

/**
 * Hook to determine the appropriate polling interval based on user journey state
 * Returns shorter interval (5s) during onboarding, normal interval (20s) otherwise
 */
export const useJourneyPollingInterval = (account: TAccount | null) => {
  return useMemo(() => {
    if (!account) {
      // No account data yet - use fast polling to get initial data
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.log('📊 Account polling: 5s (no account data)')
      }
      return 5000
    }

    const accountWithJourneys = account as any
    const evaluationJourney = accountWithJourneys?.user_journeys?.find(
      (journey: TUserJourney) => journey.name === 'evaluation'
    )

    if (!evaluationJourney) {
      // No evaluation journey - use normal polling
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.log('📊 Account polling: 20s (no evaluation journey)')
      }
      return 20000
    }

    // Check if any steps are incomplete
    const hasIncompleteSteps = evaluationJourney.steps.some(
      (step: any) => !step.complete
    )

    if (hasIncompleteSteps) {
      // Journey is active - use fast polling for responsiveness
      if (process.env.NODE_ENV === 'development') {
        const incompleteSteps = evaluationJourney.steps
          .filter((step: any) => !step.complete)
          .map((step: any) => step.name)
        // eslint-disable-next-line no-console
        console.log(`📊 Account polling: 5s (onboarding active - incomplete: ${incompleteSteps.join(', ')})`)
      }
      return 5000
    }

    // Journey is complete - use normal polling
    if (process.env.NODE_ENV === 'development') {
      // eslint-disable-next-line no-console
      console.log('📊 Account polling: 20s (onboarding complete)')
    }
    return 20000
  }, [account])
}