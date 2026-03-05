import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useAccount } from '@/hooks/use-account'
import type { TUserJourney } from '@/types'

export const useAutoOrgCreation = ({
  sfData,
  skipNavigation = false,
}: {
  sfData: Record<string, string>
  skipNavigation?: boolean
}) => {
  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { account } = useAccount()
  const navigate = useNavigate()

  // Check if user needs org created automatically
  const shouldAutoCreate = () => {
    const accountWithJourneys = account as any
    if (!accountWithJourneys?.user_journeys) return false

    const evaluationJourney = (
      accountWithJourneys.user_journeys as TUserJourney[]
    ).find((journey) => journey.name === 'evaluation')

    if (!evaluationJourney) return false

    const orgStep = evaluationJourney.steps.find(
      (step) => step.name === 'org_created'
    )
    return orgStep && !orgStep.complete && !isCreating
  }

  // Handle automatic org creation
  const createOrgAutomatically = async () => {
    if (isCreating) return

    setIsCreating(true)
    setError(null)

    try {
      // TODO: replace with direct API call via @/lib once org creation is wired up
      setError('Org creation not yet supported in SPA mode')
      setIsCreating(false)
    } catch (err) {
      setError('An unexpected error occurred')
      setIsCreating(false)
    }
  }

  // Retry org creation after error
  const retry = () => {
    setError(null)
    createOrgAutomatically()
  }

  return {
    isCreating,
    error,
    shouldAutoCreate: shouldAutoCreate(),
    createOrgAutomatically,
    retry,
  }
}
