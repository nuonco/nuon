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
  branchId?: string
  excludeBuildId?: string
}

export const SandboxBuildTimeline = ({
  builds,
  pagination,
  orgId,
  appId,
  isEmpty,
  branchId,
  excludeBuildId,
}: ISandboxBuildTimeline) => {
  const filtered = builds.filter(
    (b) =>
      b.id !== excludeBuildId && (!branchId || b.app_branch_id === branchId)
  )

  const isFiltered = !!excludeBuildId || !!branchId
  const showEmpty =
    filtered.length === 0 && (isEmpty || (isFiltered && !pagination.hasNext))

  if (showEmpty) {
    return (
      <EmptyState
        emptyTitle="No previous builds"
        emptyMessage="Previous sandbox builds will appear here once triggered."
        variant="history"
      />
    )
  }

  return (
    <Timeline<TAppSandboxBuild>
      events={filtered}
      pagination={pagination}
      renderEvent={(build) => {
        const href = branchId
          ? `/${orgId}/apps/${appId}/branches/${branchId}/sandbox/builds/${build.id}`
          : `/${orgId}/apps/${appId}/sandbox/builds/${build.id}`
        return (
          <TimelineEvent
            key={build.id}
            caption={<ID>{build?.id}</ID>}
            createdAt={build?.created_at}
            status={build?.status}
            title={
              <span className="flex items-center gap-2">
                <Link href={href} variant="inline">
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
