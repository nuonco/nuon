import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { BranchVcsBadges } from '@/components/branches/BranchVcsBadges'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { BranchPlanDots, type TBranchPlanGroup } from './BranchPlanDots'

export type TBranchCardData = {
  branchId: string
  name: string
  href: string
  repo?: string
  repoBranch?: string
  latestRun?: {
    href?: string
    status: string
    commitMessage?: string
    author?: string
    avatarUrl?: string
    sha?: string
    createdAt?: string
    awaitingApproval?: boolean
  }
  planGroups?: TBranchPlanGroup[]
  action?: ReactNode
}

export const BranchCard = ({ card }: { card: TBranchCardData }) => {
  const { action, href, latestRun, name, planGroups, repo, repoBranch } = card

  return (
    <Card className="gap-3 p-4 min-w-0 md:grid md:grid-cols-[minmax(12rem,1fr)_minmax(18rem,2fr)_minmax(12rem,1fr)] md:items-center">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1.5">
          <span className="flex min-w-0 items-center gap-2">
            <Link href={href} className="font-strong truncate max-w-56">
              {name}
            </Link>
          </span>
          {repo || repoBranch ? (
            <span className="flex items-center gap-2 min-w-0 flex-wrap">
              <BranchVcsBadges repo={repo} branch={repoBranch} />
            </span>
          ) : null}
        </div>
        <span className="flex shrink-0 items-center gap-2">
          {latestRun?.awaitingApproval ? (
            <Badge size="sm" theme="warn">
              Awaiting approval
            </Badge>
          ) : null}
          {action}
        </span>
      </div>

      {latestRun ? (
        <BranchRunCommit
          status={latestRun.status}
          href={latestRun.href}
          message={latestRun.commitMessage}
          author={latestRun.author}
          avatarUrl={latestRun.avatarUrl}
          sha={latestRun.sha}
          createdAt={latestRun.createdAt}
          className="border-l py-1 pl-4"
        />
      ) : (
        <Text variant="subtext" theme="neutral">
          No runs yet
        </Text>
      )}

      {planGroups && planGroups.length > 0 ? (
        <BranchPlanDots groups={planGroups} />
      ) : (
        <Text variant="subtext" theme="neutral">
          No deployment plan yet
        </Text>
      )}
    </Card>
  )
}
