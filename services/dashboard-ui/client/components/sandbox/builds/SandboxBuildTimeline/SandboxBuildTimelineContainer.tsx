import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
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

  const { data: result } = useQuery({
    queryKey: ['sandbox-builds', org?.id, app?.id, offset],
    queryFn: () =>
      getSandboxBuilds({
        orgId: org.id,
        appId: app.id,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
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
