import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
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

  const { data: result } = useQuery({
    queryKey: ['component-builds', org?.id, componentId, offset],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
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
