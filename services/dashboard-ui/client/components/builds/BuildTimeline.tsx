import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { getComponentBuilds } from '@/lib'
import type { TBuild } from '@/types'

interface IBuildTimeline
  extends Omit<ITimeline<TBuild>, 'events' | 'renderEvent'> {
  componentName: string
  componentId: string
  initBuilds?: TBuild[]
  pollInterval?: number
  shouldPoll?: boolean
}

export const BuildTimeline = ({
  componentName,
  componentId,
  initBuilds = [],
  pagination,
  pollInterval = 20000,
  shouldPoll = false,
}: IBuildTimeline) => {
  const { app } = useApp()
  const { org } = useOrg()

  const queryParams = useQueryParams({
    offset: pagination.offset,
    limit: 10,
  })
  const { data: builds } = useQuery<TBuild[]>({
    queryKey: ['component-builds', org?.id, componentId, queryParams],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: 10,
        offset: pagination?.offset,
      }).then((r) => r.data ?? []),
    initialData: initBuilds,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!componentId,
  })

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
            status={build?.status}
            title={
              <span className="flex items-center gap-2">
                <Link
                  href={`/${org.id}/apps/${app.id}/components/${componentId}/builds/${build.id}`}
                >
                  {componentName} build
                </Link>
                {build?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
              </span>
            }
            underline={
              <span className="flex flex-col mt-2">
                <Text variant="label" theme="neutral">
                  Built by: {build?.created_by?.email}
                </Text>

                {build?.vcs_connection_commit?.message &&
                build?.vcs_connection_commit?.sha ? (
                  <span className="">
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
                ) : null}
              </span>
            }
          />
        )
      }}
    />
  )
}
