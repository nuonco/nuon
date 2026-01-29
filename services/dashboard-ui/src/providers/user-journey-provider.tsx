'use client'

import { createContext, type ReactNode } from 'react'

interface UserJourneyContextValue {
  isBYOC?: boolean
}

export const UserJourneyContext = createContext<
  UserJourneyContextValue | undefined
>(undefined)

export const UserJourneyProvider = ({
  children,
  isBYOC = false,
}: {
  children: ReactNode
  isBYOC?: boolean
}) => {
  return (
    <UserJourneyContext.Provider value={{ isBYOC }}>
      {children}
    </UserJourneyContext.Provider>
  )
}
