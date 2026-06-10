import { createContext, useMemo, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery, isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { getInstallActionRun } from '@/lib'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TInstallActionRun, TWorkflow } from '@/types'

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
  shouldPoll = false,
}: {
  children: ReactNode
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()

  const extraListeners = useMemo(() => ({
    workflow: createSSEQueryListener<TWorkflow>(
      queryClient,
      (data) => ['workflow', org?.id, data?.id]
    ),
  }), [queryClient, org?.id])

  const { data: installActionRun, isLoading, error, refetch } = useSSEResourceQuery<TInstallActionRun>({
    sseUrl: org?.id && install?.id && runId
      ? `/api/orgs/${org.id}/installs/${install.id}/action-runs/${runId}/sse`
      : undefined,
    queryKey: ['install-action-run', org?.id, install?.id, runId],
    queryFn: () => getInstallActionRun({ orgId: org!.id, installId: install!.id, runId }),
    enabled: !!org?.id && !!install?.id && !!runId,
    shouldPoll,
    eventName: 'action-run',
    extraListeners,
    isFinished: isTerminalStatusV2,
  })

  if (error && !installActionRun) return <ProviderError error={error} />
  if (isLoading || !installActionRun) return <ProviderLoading />

  return (
    <InstallActionRunContext.Provider value={{ installActionRun, refresh: refetch }}>
      {children}
    </InstallActionRunContext.Provider>
  )
}
