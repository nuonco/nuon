import { createContext, useMemo, useEffect, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getSandboxBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TAppSandboxBuild } from '@/types'

type SandboxBuildContextValue = {
  build: TAppSandboxBuild
}

export const SandboxBuildContext = createContext<
  SandboxBuildContextValue | undefined
>(undefined)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function SandboxBuildProvider({
  children,
  buildId,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const queryKey = ['sandbox-build', org?.id, app?.id, buildId]

  const sseUrl = org?.id && app?.id && buildId
    ? `/api/orgs/${org.id}/apps/${app.id}/sandbox-builds/${buildId}/sse`
    : undefined

  const listeners = useMemo(() => ({
    'sandbox-build': (event: MessageEvent) => {
      try {
        const data: TAppSandboxBuild = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
      } catch {}
    },
  }), [org?.id, app?.id, buildId])

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: build, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getSandboxBuild({ orgId: org!.id, appId: app!.id, buildId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!app?.id && !!buildId,
  })

  useStatusToast({
    status: build?.status_v2?.status,
    resourceType: 'sandbox build',
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
    <SandboxBuildContext.Provider value={{ build }}>
      {children}
    </SandboxBuildContext.Provider>
  )
}
