import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getApp } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TApp } from '@/types'

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
  const { addToast } = useToast()
  const { data: app, isLoading, error, refetch } = useQuery({
    queryKey: ['app', org.id!, appId],
    queryFn: () => getApp({ orgId: org.id!, appId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!appId,
  })

  useEffect(() => {
    if (error && app) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])

  if (error && !app) return <ProviderError error={error} />

  if (isLoading || !app) return <ProviderLoading />

  return (
    <AppContext.Provider value={{ app, refresh: refetch }}>
      {children}
    </AppContext.Provider>
  )
}
