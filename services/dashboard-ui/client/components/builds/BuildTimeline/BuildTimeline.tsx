import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import type { TBuild } from '@/types'

interface IBuildTimeline {
  builds: TBuild[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  appId: string
  componentId: string
  componentName: string
}

export const BuildTimeline = ({
  builds,
  pagination,
  orgId,
  appId,
  componentId,
  componentName,
}: IBuildTimeline) => {
  return (
    <Timeline<TBuild>
      events={builds}
      pagination={pagination}
      renderEvent={(build) => {
        return (
          <TimelineEvent
            key={build.id}
            caption={<ID>{build?.id}</ID>}
            createdAt={build?.created_at}
            createdBy={build?.created_by?.email}
            status={build?.status}
            title={
              <span className="flex items-center gap-2">
                <Link
                  href={`/${orgId}/apps/${appId}/components/${componentId}/builds/${build.id}`}
                >
                  {componentName} build
                </Link>
                {build?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code">
                    drift scan
                  </Badge>
                ) : null}
                {build?.status_v2?.metadata?.duplicate_build ? (
                  <Badge variant="code" theme="warn">
                    duplicate build
                  </Badge>
                ) : null}
              </span>
            }
            underline={
              build?.vcs_connection_commit?.message &&
              build?.vcs_connection_commit?.sha ? (
                <span className="flex flex-col gap-1 mt-2">
                  <Text
                    className="truncate !flex w-full"
                    variant="label"
                    family="mono"
                  >
                    SHA: {build?.vcs_connection_commit?.sha}
                  </Text>
                  <Text
                    className="!max-w-[350px] !flex"
                    variant="label"
                    theme="neutral"
                  >
                    <span className="truncate">
                      {build?.vcs_connection_commit?.message}
                    </span>
                  </Text>
                </span>
              ) : undefined
            }
          />
        )
      }}
    />
  )
}
