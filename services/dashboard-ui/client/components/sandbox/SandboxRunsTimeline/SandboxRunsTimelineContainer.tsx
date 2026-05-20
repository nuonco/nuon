import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getInstallSandboxRuns } from '@/lib'
import { SandboxRunsTimeline } from './SandboxRunsTimeline'

const LIMIT = 10

interface ISandboxRunsTimelineContainer {
  pollInterval?: number
  shouldPoll?: boolean
}

export const SandboxRunsTimelineContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
}: ISandboxRunsTimelineContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['install-sandbox-runs', org?.id, install?.id, offset],
    [org?.id, install?.id, offset]
  )

  const sseUrl =
    org?.id && install?.id
      ? `/api/orgs/${org.id}/installs/${install.id}/sandbox-runs/sse?limit=${LIMIT}&offset=${offset}`
      : undefined

  const listeners = useMemo(
    () => ({
      'sandbox-runs': (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          queryClient.setQueryData(queryKey, data)
        } catch {}
      },
    }),
    [queryKey, queryClient]
  )

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: result } = useQuery({
    queryKey,
    queryFn: () =>
      getInstallSandboxRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const runs = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <SandboxRunsTimeline
      runs={runs}
      pagination={pagination}
      orgId={org?.id}
      installId={install?.id}
    />
  )
}
