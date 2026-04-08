import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
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

  const { data: action } = useQuery({
    queryKey: ['install-action', org?.id, install?.id, actionId, offset],
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!actionId,
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
