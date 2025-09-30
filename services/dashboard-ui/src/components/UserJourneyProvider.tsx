'use client'

import React, { type FC, useState, useEffect, createContext, useContext } from 'react'
import { createPortal } from 'react-dom'
import { ProductionReadyChecklistModal } from './Apps/ProductionReadyChecklistModal'
import { useAccount } from './AccountProvider'
import type { TAccount, TUserJourney } from '@/types'

interface UserJourneyContextType {
  showChecklist: () => void
  hideChecklist: () => void
  isChecklistOpen: boolean
}

const UserJourneyContext = createContext<UserJourneyContextType | undefined>(undefined)

export const useUserJourney = () => {
  const context = useContext(UserJourneyContext)
  if (context === undefined) {
    throw new Error('useUserJourney must be used within a UserJourneyProvider')
  }
  return context
}

interface UserJourneyProviderProps {
  children: React.ReactNode
  orgId: string // Can be empty string for users without orgs
}

export const UserJourneyProvider: FC<UserJourneyProviderProps> = ({
  children,
  orgId,
}) => {
  const { account, refreshAccount } = useAccount()
  const [showChecklistModal, setShowChecklistModal] = useState(false)
  const [userDismissedForNavigation, setUserDismissedForNavigation] = useState(false)

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
      step => !step.complete
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
    const allStepsComplete = evaluationJourney?.steps.every(step => step.complete) ?? false

    if (allStepsComplete && userDismissedForNavigation) {
      setUserDismissedForNavigation(false)
    }
  }, [account, userDismissedForNavigation])

  const handleCloseChecklistModal = async () => {
    // Check if all journey steps are complete before allowing close
    const evaluationJourney = getEvaluationJourney()
    const allStepsComplete = evaluationJourney?.steps.every(step => step.complete) ?? false

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

  const contextValue: UserJourneyContextType = {
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
              orgId={orgId}
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