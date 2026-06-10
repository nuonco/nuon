import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/hooks/use-sse-timeline-query'
import { getComponentDeploys } from '@/lib'
import { DeployTimeline } from './DeployTimeline'

const LIMIT = 10

interface IDeployTimelineContainer {
  componentName: string
  componentId: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const DeployTimelineContainer = ({
  componentName,
  componentId,
  pollInterval = 20000,
  shouldPoll = false,
}: IDeployTimelineContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading, error } = useSSETimelineQuery({
    sseUrl:
      org?.id && install?.id && componentId
        ? `/api/orgs/${org.id}/installs/${install.id}/components/${componentId}/deploys/sse?limit=${LIMIT}&offset=${offset}`
        : undefined,
    queryKey: ['component-deploys', org?.id, install?.id, componentId, offset],
    queryFn: () =>
      getComponentDeploys({
        orgId: org.id,
        installId: install.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!install?.id && !!componentId,
    shouldPoll,
    pollInterval,
    eventName: 'deploys',
  })

  const deploys = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <DeployTimeline
      deploys={deploys}
      pagination={pagination}
      orgId={org?.id}
      installId={install?.id}
      componentId={componentId}
      componentName={componentName}
      isLoading={isLoading}
      error={error}
    />
  )
}
