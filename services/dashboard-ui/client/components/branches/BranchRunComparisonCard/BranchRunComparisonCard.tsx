import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import {
  githubCommitUrl,
  resolvePrLink,
} from '@/components/branches/shared/pr-link'
import type { TBranchRunComparisonRunSummary } from '@/lib/ctl-api/apps/branches/get-branch-run-comparison'

export interface IBranchRunComparisonCard {
  label: string
  run?: TBranchRunComparisonRunSummary | null
  runHref?: string
  repoSlug?: string
  githubHeaderHref?: string
}

export const BranchRunComparisonCard = ({
  label,
  run,
  runHref,
  repoSlug,
  githubHeaderHref,
}: IBranchRunComparisonCard) => {
  const commit = run?.vcs_connection_commit
  const message = commit?.message?.split('\n')[0]?.trim()
  const sha = commit?.sha
  const commitUrl = githubCommitUrl(repoSlug, sha)
  const prLink = resolvePrLink({
    repoSlug,
    prNumber: run?.pr_number,
    commitMessage: commit?.message,
  })

  return (
    <Card className="!p-4 !gap-3 min-w-0">
      <div className="flex items-center justify-between gap-2 min-w-0">
        <Text variant="label" theme="neutral" weight="strong">
          {label}
        </Text>
        <div className="flex items-center gap-2 shrink-0">
          {githubHeaderHref ? (
            <Link href={githubHeaderHref} isExternal className="text-[13px] inline-flex items-center gap-1">
              <Icon variant="GitHub" size={14} />
              View in GitHub
            </Link>
          ) : null}
          {runHref ? (
            <Link href={runHref} variant="inline" className="text-[13px]">
              View run
            </Link>
          ) : null}
        </div>
      </div>

      {run ? (
        <>
          {(prLink || run.base_branch || run.event_type) && (
            <div className="flex items-center gap-2 flex-wrap">
              {prLink ? (
                <Link href={prLink.url} isExternal>
                  <Badge size="sm" theme="info">
                    PR #{prLink.number}
                  </Badge>
                </Link>
              ) : null}
              {run.base_branch ? (
                <Text variant="subtext" theme="neutral">
                  into {run.base_branch}
                </Text>
              ) : null}
              {run.event_type ? (
                <Badge size="sm" theme="neutral">
                  {run.event_type.replace(/_/g, ' ')}
                </Badge>
              ) : null}
            </div>
          )}

          <BranchRunCommit
            status={run.status}
            href={runHref ?? commitUrl}
            isExternal={!runHref && !!commitUrl}
            message={message}
            author={commit?.author_name}
            avatarUrl={commit?.author_avatar_url}
            sha={sha}
            createdAt={run.created_at}
          />
        </>
      ) : (
        <Text variant="subtext" theme="neutral">
          No run data
        </Text>
      )}
    </Card>
  )
}
