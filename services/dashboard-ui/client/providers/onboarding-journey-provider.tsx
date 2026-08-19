import { createContext, useContext, useEffect, useRef } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getAccount } from '@/lib'
import { capturePostHogEvent } from '@/lib/posthog'

interface IOnboardingJourneyContext {
  isLoading: boolean
  orgId: string | undefined
  isStepComplete: (stepName: string) => boolean
  getStepMetadata: (stepName: string, key: string) => unknown
}

export const OnboardingJourneyContext = createContext<IOnboardingJourneyContext | undefined>(undefined)

export function OnboardingJourneyProvider({ children }: { children: React.ReactNode }) {
  const { data: account, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['onboarding-journey-account'],
    queryFn: getAccount,
    refetchInterval: 5000,
  })

  const journey = account?.user_journeys?.find((j) => j.name === 'evaluation')
  const orgId = account?.org_ids?.[0]

  const appCreatedStep = journey?.steps?.find((s) => s.name === 'app_created')
  const appCreatedFiredRef = useRef(false)
  const appCreatedSeenIncompleteRef = useRef(false)

  useEffect(() => {
    if (!appCreatedStep || appCreatedFiredRef.current) return
    if (!appCreatedStep.complete) {
      appCreatedSeenIncompleteRef.current = true
      return
    }
    if (!appCreatedSeenIncompleteRef.current) return
    appCreatedFiredRef.current = true
    capturePostHogEvent('app_created', {
      org_id: orgId,
      app_id: appCreatedStep.metadata?.app_id,
    })
  }, [appCreatedStep, orgId])

  const getStep = (stepName: string) =>
    journey?.steps?.find((s) => s.name === stepName)

  const isStepComplete = (stepName: string): boolean => getStep(stepName)?.complete ?? false

  const getStepMetadata = (stepName: string, key: string): unknown =>
    getStep(stepName)?.metadata?.[key]

  return (
    <OnboardingJourneyContext.Provider value={{ isLoading, orgId, isStepComplete, getStepMetadata }}>
      {children}
    </OnboardingJourneyContext.Provider>
  )
}
