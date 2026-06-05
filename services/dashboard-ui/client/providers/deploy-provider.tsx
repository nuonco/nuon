import { createContext, useMemo, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getDeploy } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TComponent, TDeploy, TWorkflow } from '@/types'

type DeployContextValue = {
  deploy: TDeploy
}

export const DeployContext = createContext<DeployContextValue | undefined>(
  undefined
)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

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
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const queryKey = ['deploy', org?.id, installId, deployId]

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
    queryClient.invalidateQueries({ queryKey: ['install-component-outputs', org?.id, installId] })
    queryClient.invalidateQueries({ queryKey: ['install-component', org?.id, installId] })
  }, [queryClient, org?.id, installId])

  const sseUrl = org?.id && installId && deployId
    ? `/api/orgs/${org.id}/installs/${installId}/deploys/${deployId}/sse`
    : undefined

  const listeners = useMemo(() => ({
    deploy: (event: MessageEvent) => {
      try {
        const data: TDeploy = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        invalidateTabQueries()
      } catch {}
    },
    component: (event: MessageEvent) => {
      try {
        const data: TComponent = JSON.parse(event.data)
        queryClient.setQueryData(['component', org?.id, data?.id], data)
      } catch {}
    },
    workflow: (event: MessageEvent) => {
      try {
        const data: TWorkflow = JSON.parse(event.data)
        queryClient.setQueryData(['workflow', org?.id, data?.id], data)
      } catch {}
    },
  }), [org?.id, installId, deployId])

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: deploy, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getDeploy({ orgId: org!.id, installId, deployId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!installId && !!deployId,
  })

  useStatusToast({
    status: deploy?.status_v2?.status,
    label: deploy?.component_name,
    resourceType: 'deploy',
  })

  useEffect(() => {
    if (error && deploy) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])

  if (error && !deploy) return <ProviderError error={error} />
  if (isLoading || !deploy) return <ProviderLoading />

  return (
    <DeployContext.Provider value={{ deploy }}>
      {children}
    </DeployContext.Provider>
  )
}
