import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
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

  const { data: result, isLoading, error } = useQuery({
    queryKey: ['component-deploys', org?.id, install?.id, componentId, offset],
    queryFn: () =>
      getComponentDeploys({
        orgId: org.id,
        installId: install.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!componentId,
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
