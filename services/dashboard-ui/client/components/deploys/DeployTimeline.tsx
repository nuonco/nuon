import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getComponentDeploys } from '@/lib'
import type { TDeploy } from '@/types'

const LIMIT = 10

interface IDeployTimeline {
  componentName: string
  componentId: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const DeployTimeline = ({
  componentName,
  componentId,
  pollInterval = 20000,
  shouldPoll = false,
}: IDeployTimeline) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useQuery({
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
