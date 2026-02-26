import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { getComponentDeploys } from '@/lib'
import type { TDeploy } from '@/types'

interface IDeployTimeline
  extends Omit<ITimeline<TDeploy>, 'events' | 'renderEvent'> {
  componentName: string
  componentId: string
  initDeploys: TDeploy[]
  pollInterval?: number
  shouldPoll?: boolean
}

export const DeployTimeline = ({
  componentName,
  componentId,
  initDeploys,
  pagination,
  pollInterval = 20000,
  shouldPoll = false,
}: IDeployTimeline) => {
  const { install } = useInstall()
  const { org } = useOrg()

  const queryParams = useQueryParams({
    offset: pagination.offset,
    limit: 10,
  })
  const { data: deploys } = useQuery<TDeploy[]>({
    queryKey: ['component-deploys', org?.id, install?.id, componentId, queryParams],
    queryFn: () =>
      getComponentDeploys({
        orgId: org.id,
        installId: install.id,
        componentId,
        limit: 10,
        offset: pagination?.offset,
      }).then((r) => r.data ?? []),
    initialData: initDeploys,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <Timeline<TDeploy>
      events={deploys}
      pagination={pagination}
      renderEvent={(deploy) => {
        return (
          <TimelineEvent
            key={deploy.id}
            caption={<ID>{deploy?.id}</ID>}
            createdAt={deploy?.created_at}
            status={deploy?.status}
            title={
              <span className="flex items-center gap-2">
                <Link
                  href={`/${org.id}/installs/${install.id}/components/${componentId}/deploys/${deploy.id}`}
                >
                  {componentName}{' '}
                  {deploy?.install_deploy_type === 'teardown'
                    ? 'teardown'
                    : 'deploy'}
                </Link>
                {deploy?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
              </span>
            }
            underline={
              <Text variant="label" theme="neutral">
                Deployed by: {deploy?.created_by?.email}
              </Text>
            }
          />
        )
      }}
    />
  )
}
