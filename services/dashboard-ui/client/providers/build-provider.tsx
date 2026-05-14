import { createContext, useMemo, useEffect, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getComponentBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TBuild } from '@/types'

type BuildContextValue = {
  build: TBuild
}

export const BuildContext = createContext<BuildContextValue | undefined>(
  undefined
)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function BuildProvider({
  children,
  buildId,
  componentId,
  componentName,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  componentId: string
  componentName?: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const queryKey = ['build', org?.id, componentId, buildId]

  const sseUrl = org?.id && componentId && buildId
    ? `/api/orgs/${org.id}/components/${componentId}/builds/${buildId}/sse`
    : undefined

  const listeners = useMemo(() => ({
    build: (event: MessageEvent) => {
      try {
        const data: TBuild = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
      } catch {}
    },
  }), [org?.id, componentId, buildId])

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: build, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getComponentBuild({ orgId: org!.id, componentId, buildId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!componentId && !!buildId,
  })

  useStatusToast({
    status: build?.status_v2?.status,
    label: componentName ?? build?.component_name,
    resourceType: 'build',
  })

  useEffect(() => {
    if (error && build) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !build) return <ProviderError error={error} />
  if (isLoading || !build) return <ProviderLoading />

  return (
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
