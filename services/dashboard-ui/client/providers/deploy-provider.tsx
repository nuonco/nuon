import { createContext, useMemo, useCallback, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery, isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { useStatusToast } from '@/hooks/use-status-toast'
import { getDeploy } from '@/lib'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TComponent, TDeploy, TWorkflow } from '@/types'

type DeployContextValue = {
  deploy: TDeploy
}

export const DeployContext = createContext<DeployContextValue | undefined>(
  undefined
)

export function DeployProvider({
  children,
  deployId,
  installId,
  shouldPoll = true,
}: {
  children: ReactNode
  deployId: string
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const queryClient = useQueryClient()

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
    queryClient.invalidateQueries({ queryKey: ['install-component-outputs', org?.id, installId] })
    queryClient.invalidateQueries({ queryKey: ['install-component', org?.id, installId] })
  }, [queryClient, org?.id, installId])

  const extraListeners = useMemo(() => ({
    component: createSSEQueryListener<TComponent>(
      queryClient,
      (data) => ['component', org?.id, data?.id]
    ),
    workflow: createSSEQueryListener<TWorkflow>(
      queryClient,
      (data) => ['workflow', org?.id, data?.id]
    ),
  }), [queryClient, org?.id])

  const { data: deploy, isLoading, error } = useSSEResourceQuery<TDeploy>({
    sseUrl: org?.id && installId && deployId
      ? `/api/orgs/${org.id}/installs/${installId}/deploys/${deployId}/sse`
      : undefined,
    queryKey: ['deploy', org?.id, installId, deployId],
    queryFn: () => getDeploy({ orgId: org!.id, installId, deployId }),
    enabled: !!org?.id && !!installId && !!deployId,
    shouldPoll,
    eventName: 'deploy',
    onPrimaryEvent: invalidateTabQueries,
    extraListeners,
    isFinished: isTerminalStatusV2,
  })

  useStatusToast({
    status: deploy?.status_v2?.status,
    label: deploy?.component_name,
    resourceType: 'deploy',
  })

  if (error && !deploy) return <ProviderError error={error} />
  if (isLoading || !deploy) return <ProviderLoading />

  return (
    <DeployContext.Provider value={{ deploy }}>
      {children}
    </DeployContext.Provider>
  )
}
