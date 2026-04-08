import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import type { TAppSandboxBuild } from '@/types'

interface ISandboxBuildTimeline {
  builds: TAppSandboxBuild[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  appId: string
  isEmpty: boolean
}

export const SandboxBuildTimeline = ({
  builds,
  pagination,
  orgId,
  appId,
  isEmpty,
}: ISandboxBuildTimeline) => {
  if (isEmpty) {
    return (
      <EmptyState
        emptyTitle="No sandbox builds"
        emptyMessage="Sandbox builds will appear here once triggered."
        variant="history"
      />
    )
  }

  return (
    <Timeline<TAppSandboxBuild>
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
                  href={`/${orgId}/apps/${appId}/sandbox/builds/${build.id}`}
                >
                  Sandbox build
                </Link>
                {build?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
                {build?.status_v2?.metadata?.duplicate_build ? (
                  <Badge variant="code" size="sm" theme="warn">
                    duplicate build
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
                  <span>
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
