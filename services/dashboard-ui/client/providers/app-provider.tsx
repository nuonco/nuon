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
  appId,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  appId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: app, error, isLoading, refetch } = useQuery({
    queryKey: ['app', org?.id, appId],
    queryFn: () => getApp({ orgId: org.id, appId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!appId,
  })

  return (
    <AppContext.Provider
      value={{
        app: app ?? null,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </AppContext.Provider>
  )
}
