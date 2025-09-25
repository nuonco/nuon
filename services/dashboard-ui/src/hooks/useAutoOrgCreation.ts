'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAccount } from '@/components/AccountProvider'
import { createTrialOrganization } from '@/components/org-actions'
import type { TUserJourney } from '@/types'

export const useAutoOrgCreation = () => {
  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { account, refreshAccount } = useAccount()
  const router = useRouter()

  // Check if user needs org created automatically
  const shouldAutoCreate = () => {
    const accountWithJourneys = account as any
    if (!accountWithJourneys?.user_journeys) return false

    const evaluationJourney = (
      accountWithJourneys.user_journeys as TUserJourney[]
    ).find((journey) => journey.name === 'evaluation')

    if (!evaluationJourney) return false

    const orgStep = evaluationJourney.steps.find(step => step.name === 'org_created')
    return orgStep && !orgStep.complete && !isCreating
  }

  // Handle automatic org creation
  const createOrgAutomatically = async () => {
    if (isCreating) return

    setIsCreating(true)
    setError(null)

    try {
      const { data: newOrg, error: createError } = await createTrialOrganization()

      if (createError !== null) {
        setError(createError?.error || 'Failed to create organization')
        setIsCreating(false)
      } else {
        // Success - refresh account to get updated journey
        await refreshAccount()
        setIsCreating(false)

        // Navigate to the new org
        if (newOrg?.id) {
          router.push(`/${newOrg.id}/apps`)
        }
      }
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

  // Auto-trigger creation when conditions are met
  useEffect(() => {
    if (shouldAutoCreate()) {
      createOrgAutomatically()
    }
  }, [account])

  return {
    isCreating,
    error,
    shouldAutoCreate: shouldAutoCreate(),
    createOrgAutomatically,
    retry
  }
}