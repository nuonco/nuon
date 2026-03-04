import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallActionRun } from '@/lib'
import { Loading } from '@/components/common/Loading'
import type { TInstallActionRun } from '@/types'

type InstallActionRunContextValue = {
  installActionRun: TInstallActionRun
  refresh: () => void
}

export const InstallActionRunContext = createContext<
  InstallActionRunContextValue | undefined
>(undefined)

export function InstallActionRunProvider({
  children,
  runId,
  pollInterval = 3000,
  shouldPoll = false,
}: {
  children: ReactNode
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: installActionRun, isLoading, refetch } = useQuery({
    queryKey: ['install-action-run', org.id!, install.id!, runId],
    queryFn: () => getInstallActionRun({ orgId: org.id!, installId: install.id!, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!install.id && !!runId,
  })

  if (isLoading || !installActionRun) return <Loading />

  return (
    <InstallActionRunContext.Provider value={{ installActionRun, refresh: refetch }}>
      {children}
    </InstallActionRunContext.Provider>
  )
}
