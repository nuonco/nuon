import { createContext, useMemo, useEffect, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useToast } from '@/hooks/use-toast'
import { getInstallActionRun } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TInstallActionRun, TWorkflow } from '@/types'

type InstallActionRunContextValue = {
  installActionRun: TInstallActionRun
  refresh: () => void
}

export const InstallActionRunContext = createContext<
  InstallActionRunContextValue | undefined
>(undefined)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

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
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const queryKey = ['install-action-run', org?.id, install?.id, runId]

  const sseUrl = org?.id && install?.id && runId
    ? `/api/orgs/${org.id}/installs/${install.id}/action-runs/${runId}/sse`
    : undefined

  const listeners = useMemo(() => ({
    'action-run': (event: MessageEvent) => {
      try {
        const data: TInstallActionRun = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
      } catch {}
    },
    workflow: (event: MessageEvent) => {
      try {
        const data: TWorkflow = JSON.parse(event.data)
        queryClient.setQueryData(['workflow', org?.id, data?.id], data)
      } catch {}
    },
  }), [org?.id, install?.id, runId])

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: installActionRun, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: () => getInstallActionRun({ orgId: org!.id, installId: install!.id, runId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!install?.id && !!runId,
  })

  useEffect(() => {
    if (error && installActionRun) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !installActionRun) return <ProviderError error={error} />
  if (isLoading || !installActionRun) return <ProviderLoading />

  return (
    <InstallActionRunContext.Provider value={{ installActionRun, refresh: refetch }}>
      {children}
    </InstallActionRunContext.Provider>
  )
}
