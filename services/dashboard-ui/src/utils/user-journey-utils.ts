import type { TAccount, TUserJourney, TUserJourneyStep } from '@/types'

// Get evaluation journey from account
export const getUserJourney = (account: TAccount, journeyName: string) => {
  const accountWithJourneys = account as any
  if (!accountWithJourneys?.user_journeys) return null

  return (accountWithJourneys.user_journeys as TUserJourney[]).find(
    (journey) => journey.name === journeyName
  )
}

export const getUserJourneyStep = (
  account: TAccount,
  journeyName: string,
  stepName: string
) => {
  return getUserJourney(account, journeyName)?.steps?.find(
    (s: any) => s.name === stepName
  )
}

export const getUserJourneyStepMetadata = (
  account: TAccount,
  journeyName: string,
  stepName: string,
  metadataKey: string
) => {
  return getUserJourneyStep(account, journeyName, stepName)?.metadata?.[
    metadataKey
  ]
}

export const getCurrentStep = (
  steps: TUserJourneyStep[]
): TUserJourneyStep | null => {
  return steps.find((step) => !step.complete) || null
}
