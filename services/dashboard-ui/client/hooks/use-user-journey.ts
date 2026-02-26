'use client'

import { useContext } from 'react'
import { UserJourneyContext } from '@/providers/user-journey-provider'

export const useUserJourney = () => {
  const context = useContext(UserJourneyContext)
  if (context === undefined) {
    throw new Error('useUserJourney must be used within a UserJourneyProvider')
  }
  return context
}
