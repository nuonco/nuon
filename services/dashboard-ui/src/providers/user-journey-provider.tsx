'use client'

import { useParams } from 'next/navigation'
import { useState, useEffect, createContext, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { ProductionReadyChecklistModal } from '@/components/Apps/ProductionReadyChecklistModal'
import { useAccount } from '@/hooks/use-account'
import type { TUserJourney } from '@/types'

interface UserJourneyContextValue {
  showChecklist: () => void
  hideChecklist: () => void
  isChecklistOpen: boolean
}

export const UserJourneyContext = createContext<
  UserJourneyContextValue | undefined
>(undefined)

export const UserJourneyProvider = ({ children }: { children: ReactNode }) => {
  const params = useParams()
  const { ['org-id']: orgId } = params
  const { account, refreshAccount } = useAccount()
  const [showChecklistModal, setShowChecklistModal] = useState(false)
  const [userDismissedForNavigation, setUserDismissedForNavigation] =
    useState(false)

  // Get evaluation journey from account
  const getEvaluationJourney = () => {
    const accountWithJourneys = account as any
    if (!accountWithJourneys?.user_journeys) return null

    return (accountWithJourneys.user_journeys as TUserJourney[]).find(
      (journey) => journey.name === 'evaluation'
    )
  }

  // Check if user should see the journey modal (for any incomplete steps)
  const shouldShowChecklist = () => {
    const evaluationJourney = getEvaluationJourney()
    if (!evaluationJourney) return false

    // Show modal if ANY step is incomplete - modal persists until journey complete
    const hasIncompleteSteps = evaluationJourney.steps.some(
      (step) => !step.complete
    )

    return hasIncompleteSteps
  }

  // Show journey modal based on incomplete steps and dismissal state
  useEffect(() => {
    const shouldShow = shouldShowChecklist() && !userDismissedForNavigation
    setShowChecklistModal(shouldShow)
  }, [account, userDismissedForNavigation])

  // Reset dismissal flag when journey is complete (for future journeys)
  useEffect(() => {
    const evaluationJourney = getEvaluationJourney()
    const allStepsComplete =
      evaluationJourney?.steps.every((step) => step.complete) ?? false

    if (allStepsComplete && userDismissedForNavigation) {
      setUserDismissedForNavigation(false)
    }
  }, [account, userDismissedForNavigation])

  const handleCloseChecklistModal = async () => {
    // Check if all journey steps are complete before allowing close
    const evaluationJourney = getEvaluationJourney()
    const allStepsComplete =
      evaluationJourney?.steps.every((step) => step.complete) ?? false

    if (allStepsComplete) {
      // All steps complete - allow modal to close
      setShowChecklistModal(false)
    } else {
      // Steps still incomplete - just refresh account data
      await refreshAccount()
    }
  }

  const showChecklist = () => {
    // Modal visibility is now purely based on journey completion
    // This method is kept for API compatibility but doesn't override journey logic
  }

  const hideChecklist = () => {
    // Modal visibility is now purely based on journey completion
    // This method is kept for API compatibility but doesn't override journey logic
  }

  const contextValue: UserJourneyContextValue = {
    showChecklist,
    hideChecklist,
    isChecklistOpen: showChecklistModal,
  }

  return (
    <UserJourneyContext.Provider value={contextValue}>
      {children}

      {/* Journey checklist modal - shows for all incomplete steps including org creation */}
      {showChecklistModal && typeof document !== 'undefined'
        ? createPortal(
            <ProductionReadyChecklistModal
              isOpen={showChecklistModal}
              onClose={handleCloseChecklistModal}
              account={account}
              orgId={orgId as string}
              onForceClose={() => {
                setUserDismissedForNavigation(true)
                setShowChecklistModal(false)
              }}
            />,
            document.body
          )
        : null}
    </UserJourneyContext.Provider>
  )
}
