'use client'

import { useState, useEffect, createContext, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { FullScreenOnboarding } from '@/components/old/Apps/FullScreenOnboarding'
import { useAccount } from '@/hooks/use-account'
import type { TAccount } from '@/types'
import { getUserJourney } from '@/utils/user-journey-utils'

interface UserJourneyContextValue {
  isViewOpen: boolean
  openOnboarding: () => void
}

export const UserJourneyContext = createContext<
  UserJourneyContextValue | undefined
>(undefined)

// check if any journey steps are incomplete
const incompleteSteps = (account: TAccount) => {
  const evaluationJourney = getUserJourney(account, 'evaluation')
  if (!evaluationJourney) return false

  // Show view if ANY step is incomplete - view persists until journey complete
  const hasIncompleteSteps = evaluationJourney.steps.some(
    (step: any) => !step.complete
  )

  return hasIncompleteSteps
}

export const UserJourneyProvider = ({ children }: { children: ReactNode }) => {
  const { account, refreshAccount } = useAccount()
  const [showJourneyView, setShowJourneyView] = useState(false)
  const [manuallyOpened, setManuallyOpened] = useState(false)

  // Show journey view based on incomplete steps and manually opened
  useEffect(() => {
    // Skip auto logic when manually opened
    if (manuallyOpened) {
      setShowJourneyView(true)
      return
    }

    setShowJourneyView(incompleteSteps(account))
  }, [account, manuallyOpened])

  // Add method to manually open onboarding (without affecting journey state)
  const openOnboarding = () => {
    setManuallyOpened(true) // Prevent auto-close logic
    setShowJourneyView(true)
    // Note: This does NOT update journey steps, just reopens the modal
  }

  const handleCloseViewModal = async () => {
    // Reset manual flag to allow auto logic to resume
    setManuallyOpened(false)

    // Check if all journey steps are complete before allowing close
    const evaluationJourney = getUserJourney(account, 'evaluation')
    const allStepsComplete =
      evaluationJourney?.steps.every((step: any) => step.complete) ?? false

    if (allStepsComplete) {
      // All steps complete - allow view to close
      setShowJourneyView(false)
    } else {
      // Steps still incomplete - just refresh account data
      await refreshAccount()
    }
  }

  const contextValue: UserJourneyContextValue = {
    isViewOpen: showJourneyView,
    openOnboarding,
  }

  const onboarding = (
    <FullScreenOnboarding
      isOpen={showJourneyView}
      onClose={handleCloseViewModal}
      account={account}
    />
  )

  return (
    <UserJourneyContext.Provider value={contextValue}>
      {children}

      {/* Journey checklist - shows for all incomplete steps including org creation */}
      {showJourneyView && typeof document !== 'undefined'
        ? createPortal(onboarding, document.body)
        : null}
    </UserJourneyContext.Provider>
  )
}
