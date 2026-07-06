import { createContext, useEffect, useMemo, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getApp, getAppLabels } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TApp } from '@/types'

type AppContextValue = {
  app: TApp
  labelColors: Record<string, string>
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

  const { data: labelsData } = useQuery({
    queryKey: ['app-labels', org.id!, appId],
    queryFn: () => getAppLabels({ orgId: org.id!, appId }),
    enabled: !!org.id && !!appId,
  })

  const labelColors = useMemo(() => {
    const colors: Record<string, string> = {}
    for (const lk of labelsData?.labels ?? []) {
      colors[lk.key] = lk.color
    }
    return colors
  }, [labelsData])

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
    <AppContext.Provider value={{ app, labelColors, refresh: refetch }}>
      {children}
    </AppContext.Provider>
  )
}
