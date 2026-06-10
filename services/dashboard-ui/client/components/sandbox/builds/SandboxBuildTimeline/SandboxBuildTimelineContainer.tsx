import { useSearchParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/hooks/use-sse-timeline-query'
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
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useSSETimelineQuery({
    sseUrl:
      org?.id && app?.id
        ? `/api/orgs/${org.id}/apps/${app.id}/sandbox-builds/sse?limit=${LIMIT}&offset=${offset}`
        : undefined,
    queryKey: ['sandbox-builds', org?.id, app?.id, offset],
    queryFn: () =>
      getSandboxBuilds({
        orgId: org.id,
        appId: app.id,
        limit: LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!app?.id,
    shouldPoll,
    pollInterval,
    eventName: 'sandbox-builds',
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
