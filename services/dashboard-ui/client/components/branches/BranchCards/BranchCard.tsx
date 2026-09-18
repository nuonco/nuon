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
    <Card className="gap-4 p-4 min-w-0">
      <div className="flex flex-col gap-1.5">
        <div className="flex items-start justify-between gap-2">
          <span className="flex items-center gap-2 min-w-0 flex-wrap">
            <Link href={href} className="font-strong truncate max-w-56">
              {name}
            </Link>
            {latestRun?.awaitingApproval ? (
              <Badge size="sm" theme="warn">
                Awaiting approval
              </Badge>
            ) : null}
          </span>
          <span className="flex items-center shrink-0 ml-auto">{action}</span>
        </div>

        {repo || repoBranch ? (
          <span className="flex items-center gap-2 min-w-0 flex-wrap">
            <BranchVcsBadges repo={repo} branch={repoBranch} />
          </span>
        ) : null}
      </div>

      <div className="flex flex-col gap-1.5">
        <Text variant="label" theme="neutral" weight="strong">
          Latest run
        </Text>
        {latestRun ? (
          <BranchRunCommit
            status={latestRun.status}
            href={latestRun.href}
            message={latestRun.commitMessage}
            author={latestRun.author}
            avatarUrl={latestRun.avatarUrl}
            sha={latestRun.sha}
            createdAt={latestRun.createdAt}
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            No runs yet
          </Text>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <Text variant="label" theme="neutral" weight="strong">
          Deployment plan
        </Text>
        {planGroups && planGroups.length > 0 ? (
          <BranchPlanDots groups={planGroups} />
        ) : (
          <Text variant="subtext" theme="neutral">
            No deployment plan yet
          </Text>
        )}
      </div>
    </Card>
  )
}
