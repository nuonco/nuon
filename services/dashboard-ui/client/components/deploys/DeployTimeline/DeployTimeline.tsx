import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { Text } from '@/components/common/Text'
import type { TDeploy } from '@/types'

interface IDeployTimeline {
  deploys: TDeploy[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  installId: string
  componentId: string
  componentName: string
  isLoading: boolean
  error: unknown
}

export const DeployTimeline = ({
  deploys,
  pagination,
  orgId,
  installId,
  componentId,
  componentName,
  isLoading,
  error,
}: IDeployTimeline) => {
  if (isLoading) {
    return <TimelineSkeleton />
  }

  if (error || deploys.length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No deploys"
        emptyMessage="This component has not been deployed yet."
      />
    )
  }

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
                  href={`/${orgId}/installs/${installId}/components/${componentId}/deploys/${deploy.id}`}
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
