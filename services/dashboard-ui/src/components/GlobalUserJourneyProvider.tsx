'use client'

import React, { type FC } from 'react'
import { UserJourneyProvider } from './UserJourneyProvider'
import { useCurrentOrgId } from '@/hooks/useCurrentOrgId'

interface GlobalUserJourneyProviderProps {
  children: React.ReactNode
}

/**
 * Global wrapper for UserJourneyProvider that dynamically determines orgId
 * from the current route. This allows the journey modal to persist across
 * navigation while still having access to the correct org context.
 */
export const GlobalUserJourneyProvider: FC<GlobalUserJourneyProviderProps> = ({
  children,
}) => {
  const orgId = useCurrentOrgId()

  return (
    <UserJourneyProvider orgId={orgId}>
      {children}
    </UserJourneyProvider>
  )
}