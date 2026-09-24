import { Badge } from '@/components/common/Badge'
import { Card, type ICard } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TAppBranchRun } from '@/types'
import { humanize } from '@/utils/string-utils'

export interface ICurrentAppBranchRun extends ICard {
  run?: TAppBranchRun
  orgId?: string
  appId?: string
  branchName?: string
  isLoading?: boolean
}

export const CurrentAppBranchRun = ({
  run,
  orgId,
  appId,
  branchName,
  isLoading,
  ...props
}: ICurrentAppBranchRun) => {
  const commit = run?.vcs_connection_commit
  const sha = commit?.sha || run?.head_sha

  return (
    <Card {...props}>
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <Text variant="h3">App branch</Text>
          {run ? (
            <Text variant="subtext" theme="neutral">
              The latest branch run applied to this install.
            </Text>
          ) : null}
        </div>
        {run?.status ? <Status variant="badge" status={run.status} /> : null}
      </div>

      {isLoading && !run ? (
        <div className="flex flex-col gap-3">
          <Text variant="subtext" loading loadingWidth={24} />
          <Text variant="subtext" loading loadingWidth={40} />
        </div>
      ) : run ? (
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-2">
            {run.app_branch?.name ? (
              <Badge size="sm" variant="code">
                {run.app_branch.name}
              </Badge>
            ) : null}
            {run.preview?.mode ? (
              <Badge size="sm" theme="info">
                {humanize(run.preview.mode)}
              </Badge>
            ) : null}
            {sha ? <ID>{sha.slice(0, 7)}</ID> : null}
            {run.pr_number ? (
              <Badge size="sm" theme="neutral">
                PR #{run.pr_number}
              </Badge>
            ) : null}
          </div>
          {commit?.message ? (
            <Text variant="subtext">{commit.message.split('\n')[0]}</Text>
          ) : null}
          {orgId && appId && run.app_branch?.id && run.id ? (
            <Link
              href={`/${orgId}/apps/${appId}/branches/${run.app_branch.id}/runs/${run.id}`}
            >
              View branch run
            </Link>
          ) : null}
        </div>
      ) : branchName ? (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle="No branch run applied yet"
          emptyMessage={`Runs on ${branchName} appear here once one applies to this install.`}
        />
      ) : (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle="No app branch connected"
          emptyMessage="Connect this install to an app branch to track branch runs here."
        />
      )}
    </Card>
  )
}
