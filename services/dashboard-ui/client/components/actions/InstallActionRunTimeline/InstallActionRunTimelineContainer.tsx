import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/hooks/use-sse-timeline-query'
import { getInstallAction } from '@/lib'
import { InstallActionRunTimeline } from './InstallActionRunTimeline'

const LIMIT = 10

interface IInstallActionRunTimelineContainer {
  actionId: string
  actionName: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const InstallActionRunTimelineContainer = ({
  actionId,
  actionName,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallActionRunTimelineContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: action } = useSSETimelineQuery({
    sseUrl:
      org?.id && install?.id && actionId
        ? `/api/orgs/${org.id}/installs/${install.id}/actions/${actionId}/runs/sse?limit=${LIMIT}&offset=${offset}`
        : undefined,
    queryKey: ['install-action', org?.id, install?.id, actionId, offset],
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId,
        limit: LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!install?.id && !!actionId,
    shouldPoll,
    pollInterval,
    eventName: 'action-runs',
    transform: (data) => data?.data,
  })

  const runs = action?.runs ?? []
  const basePath = `/${org.id}/installs/${install.id}`

  return (
    <InstallActionRunTimeline
      actionId={actionId}
      actionName={actionName}
      runs={runs}
      basePath={basePath}
      pagination={{ hasNext: runs.length >= LIMIT, offset, limit: LIMIT }}
    />
  )
}
