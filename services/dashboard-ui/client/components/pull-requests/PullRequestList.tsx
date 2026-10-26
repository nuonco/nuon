import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TPreviewSourcePR } from '@/types'

export interface IPullRequestList {
  pullRequests: TPreviewSourcePR[]
  basePath: string
  isLoading?: boolean
}

const SKELETON_ROWS = [0, 1, 2]

export const PullRequestList = ({
  pullRequests,
  basePath,
  isLoading = false,
}: IPullRequestList) => {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-2">
        {SKELETON_ROWS.map((row) => (
          <Card key={row}>
            <div className="flex flex-col gap-1">
              <span className="flex flex-wrap items-baseline gap-2">
                <Text variant="body" loading loadingWidth={5} />
                <Text variant="body" loading loadingWidth={36} />
              </span>
              <span className="flex flex-wrap items-center gap-2">
                <Badge size="xs" loading loadingWidth={18} />
                <Text variant="subtext" loading loadingWidth={7} />
              </span>
            </div>
          </Card>
        ))}
      </div>
    )
  }

  if (pullRequests.length === 0) {
    return (
      <EmptyState
        emptyTitle="No open pull requests"
        emptyMessage="Pull requests targeting this branch show up here once GitHub reports them."
      />
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {pullRequests.map((pullRequest) => (
        <Card key={pullRequest.pr_number}>
          <Link
            href={`${basePath}/pull-requests/${pullRequest.pr_number}`}
            className="flex flex-col gap-1"
          >
            <span className="flex flex-wrap items-baseline gap-2">
              <Text variant="body" family="mono" weight="strong">
                #{pullRequest.pr_number}
              </Text>
              <Text variant="body">{pullRequest.title}</Text>
            </span>
            <span className="flex flex-wrap items-center gap-2">
              {pullRequest.head_ref ? (
                <Badge size="xs" variant="code" theme="neutral">
                  {pullRequest.head_ref}
                </Badge>
              ) : null}
              {pullRequest.head_sha ? (
                <Text variant="subtext" family="mono" theme="neutral">
                  {pullRequest.head_sha.slice(0, 7)}
                </Text>
              ) : null}
            </span>
          </Link>
        </Card>
      ))}
    </div>
  )
}
