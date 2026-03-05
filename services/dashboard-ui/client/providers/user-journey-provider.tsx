import { createContext, type ReactNode } from 'react'

interface UserJourneyContextValue {
  isBYOC?: boolean
  isCustomerPortalEnabled?: boolean
}

export const UserJourneyContext = createContext<
  UserJourneyContextValue | undefined
>(undefined)

export const UserJourneyProvider = ({
  children,
  isBYOC = false,
  isCustomerPortalEnabled = false,
}: {
  children: ReactNode
  isBYOC?: boolean
  isCustomerPortalEnabled?: boolean
}) => {
  return (
    <UserJourneyContext.Provider value={{ isBYOC, isCustomerPortalEnabled }}>
      {children}
    </UserJourneyContext.Provider>
  )
}
