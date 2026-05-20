import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getSandboxBuilds } from '@/lib'
import { SandboxBuildTimeline } from './SandboxBuildTimeline'

const LIMIT = 10

interface ISandboxBuildTimelineContainer {
  pollInterval?: number
  shouldPoll?: boolean
}

export const SandboxBuildTimelineContainer = ({
  pollInterval = 10000,
  shouldPoll = false,
}: ISandboxBuildTimelineContainer) => {
  const { app } = useApp()
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['sandbox-builds', org?.id, app?.id, offset],
    [org?.id, app?.id, offset]
  )

  const sseUrl =
    org?.id && app?.id
      ? `/api/orgs/${org.id}/apps/${app.id}/sandbox-builds/sse?limit=${LIMIT}&offset=${offset}`
      : undefined

  const listeners = useMemo(
    () => ({
      'sandbox-builds': (event: MessageEvent) => {
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
      getSandboxBuilds({
        orgId: org.id,
        appId: app.id,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  const builds = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <SandboxBuildTimeline
      builds={builds}
      pagination={pagination}
      orgId={org?.id}
      appId={app?.id}
      isEmpty={builds.length === 0 && offset === 0}
    />
  )
}
