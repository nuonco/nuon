import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getApp } from '@/lib'
import type { TApp } from '@/types'

type AppContextValue = {
  app: TApp | null
  isLoading: boolean
  error: any
  refresh: () => void
}

export const AppContext = createContext<AppContextValue | undefined>(undefined)

export function AppProvider({
  children,
  initApp,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  initApp: TApp
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: app = initApp, error, isLoading, refetch } = useQuery({
    queryKey: ['app', org.id, initApp.id],
    queryFn: () => getApp({ orgId: org.id, appId: initApp.id }),
    initialData: initApp,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <AppContext.Provider
      value={{
        app,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </AppContext.Provider>
  )
}
