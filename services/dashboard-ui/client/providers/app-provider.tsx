import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getApp } from '@/lib'
import { Loading } from '@/components/common/Loading'
import type { TApp } from '@/types'

type AppContextValue = {
  app: TApp
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
  const { data: app, isLoading, refetch } = useQuery({
    queryKey: ['app', org.id!, appId],
    queryFn: () => getApp({ orgId: org.id!, appId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!appId,
  })

  if (isLoading || !app) return <Loading />

  return (
    <AppContext.Provider value={{ app, refresh: refetch }}>
      {children}
    </AppContext.Provider>
  )
}
