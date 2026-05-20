import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getComponentBuilds } from '@/lib'
import { BuildTimeline } from './BuildTimeline'

const LIMIT = 10

interface IBuildTimelineContainer {
  componentName: string
  componentId: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const BuildTimelineContainer = ({
  componentName,
  componentId,
  pollInterval = 10000,
  shouldPoll = false,
}: IBuildTimelineContainer) => {
  const { app } = useApp()
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['component-builds', org?.id, componentId, offset],
    [org?.id, componentId, offset]
  )

  const sseUrl =
    org?.id && componentId
      ? `/api/orgs/${org.id}/components/${componentId}/builds/sse?limit=${LIMIT}&offset=${offset}&appId=${app?.id ?? ''}`
      : undefined

  const listeners = useMemo(
    () => ({
      builds: (event: MessageEvent) => {
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
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!componentId,
  })

  const builds = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <BuildTimeline
      builds={builds}
      pagination={pagination}
      orgId={org?.id}
      appId={app?.id}
      componentId={componentId}
      componentName={componentName}
    />
  )
}
