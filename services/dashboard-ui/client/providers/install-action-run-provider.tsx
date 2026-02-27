import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallActionRun } from '@/lib'
import type { TInstallActionRun } from '@/types'

type InstallActionRunContextValue = {
  installActionRun: TInstallActionRun | null
  isLoading: boolean
  error: any
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
  const {
    data: installActionRun,
    error,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['install-action-run', org?.id, install?.id, runId],
    queryFn: () => getInstallActionRun({ orgId: org.id, installId: install.id, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!runId,
  })

  return (
    <InstallActionRunContext.Provider
      value={{
        installActionRun: installActionRun ?? null,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </InstallActionRunContext.Provider>
  )
}
