import { createContext, useContext } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAccount } from '@/lib'

interface IOnboardingJourneyContext {
  isLoading: boolean
  orgId: string | undefined
  isStepComplete: (stepName: string) => boolean
  getStepMetadata: (stepName: string, key: string) => unknown
}

export const OnboardingJourneyContext = createContext<IOnboardingJourneyContext | undefined>(undefined)

export function OnboardingJourneyProvider({ children }: { children: React.ReactNode }) {
  const { data: account, isLoading } = useQuery({
    queryKey: ['onboarding-journey-account'],
    queryFn: getAccount,
    refetchInterval: 5000,
  })

  const journey = account?.user_journeys?.find((j) => j.name === 'evaluation')
  const orgId = account?.org_ids?.[0]

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
