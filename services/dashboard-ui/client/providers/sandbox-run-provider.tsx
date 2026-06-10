import { createContext, useMemo, useCallback, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery, isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { useStatusToast } from '@/hooks/use-status-toast'
import { getInstallSandboxRun } from '@/lib'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TSandboxRun, TWorkflow } from '@/types'

type SandboxRunContextValue = {
  sandboxRun: TSandboxRun
}

export const SandboxRunContext = createContext<SandboxRunContextValue | undefined>(
  undefined
)

export function SandboxRunProvider({
  children,
  installId,
  runId,
  shouldPoll = true,
}: {
  children: ReactNode
  installId?: string
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const queryClient = useQueryClient()

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
  }, [queryClient])

  const extraListeners = useMemo(() => ({
    workflow: createSSEQueryListener<TWorkflow>(
      queryClient,
      (data) => ['workflow', org?.id, data?.id]
    ),
  }), [queryClient, org?.id])

  const { data: sandboxRun, isLoading, error } = useSSEResourceQuery<TSandboxRun>({
    sseUrl: org?.id && installId && runId
      ? `/api/orgs/${org.id}/installs/${installId}/sandbox-runs/${runId}/sse`
      : undefined,
    queryKey: ['sandbox-run', org?.id, runId],
    queryFn: () => getInstallSandboxRun({ orgId: org!.id, runId }),
    enabled: !!org?.id && !!runId,
    shouldPoll,
    eventName: 'sandbox-run',
    onPrimaryEvent: invalidateTabQueries,
    extraListeners,
    isFinished: isTerminalStatusV2,
  })

  useStatusToast({
    status: sandboxRun?.status_v2?.status,
    resourceType: 'sandbox run',
  })

  if (error && !sandboxRun) return <ProviderError error={error} />
  if (isLoading || !sandboxRun) return <ProviderLoading />

  return (
    <SandboxRunContext.Provider value={{ sandboxRun }}>
      {children}
    </SandboxRunContext.Provider>
  )
}
