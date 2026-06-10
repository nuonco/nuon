import { useSearchParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/hooks/use-sse-timeline-query'
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
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useSSETimelineQuery({
    sseUrl:
      org?.id && componentId
        ? `/api/orgs/${org.id}/components/${componentId}/builds/sse?limit=${LIMIT}&offset=${offset}&appId=${app?.id ?? ''}`
        : undefined,
    queryKey: ['component-builds', org?.id, componentId, offset],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!componentId,
    shouldPoll,
    pollInterval,
    eventName: 'builds',
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
