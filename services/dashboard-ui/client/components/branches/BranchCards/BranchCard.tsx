import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { BranchVcsBadges } from '@/components/branches/BranchVcsBadges'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import {
  MiniDeploymentView,
  type TMiniDeployInstall,
} from '@/components/branches/MiniDeploymentView'
import { BranchPlanDots, type TBranchPlanGroup } from './BranchPlanDots'

export type TBranchRunTrigger =
  | 'manual'
  | 'push'
  | 'pull_request'
  | 'tag'
  | 'github_label'

const TRIGGER_LABELS: Record<TBranchRunTrigger, string> = {
  manual: 'Manual run',
  push: 'Commit pushed',
  pull_request: 'Pull request',
  tag: 'Tag pushed',
  github_label: 'Label triggered',
}

const RunMeta = ({
  trigger,
  prNumber,
  tag,
}: {
  trigger?: TBranchRunTrigger
  prNumber?: number
  tag?: string
}) => {
  if (prNumber == null && !tag && !trigger) return null

  return (
    <span className="flex items-center gap-2 min-w-0 flex-wrap">
      {prNumber != null ? (
        <Badge size="sm" theme="info">
          PR #{prNumber}
        </Badge>
      ) : null}
      {tag ? (
        <Badge size="sm" theme="neutral" variant="code">
          {tag}
        </Badge>
      ) : null}
      {trigger && prNumber == null && !tag ? (
        <Badge size="sm" theme="neutral">
          {TRIGGER_LABELS[trigger]}
        </Badge>
      ) : null}
    </span>
  )
}

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
    trigger?: TBranchRunTrigger
    prNumber?: number
    tag?: string
  }
  planGroups?: TBranchPlanGroup[]
  latestRunInstalls?: TMiniDeployInstall[]
  action?: ReactNode
}

export const BranchCard = ({ card }: { card: TBranchCardData }) => {
  const {
    action,
    href,
    latestRun,
    latestRunInstalls,
    name,
    planGroups,
    repo,
    repoBranch,
  } = card
  const hasPlan = Boolean(planGroups && planGroups.length > 0)
  const canExpand = Boolean(
    hasPlan && latestRun && latestRunInstalls && latestRunInstalls.length > 0
  )

  return (
    <Card className="gap-4 p-5 min-w-0">
      <div className="flex flex-col gap-1">
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
          <span className="flex items-center shrink-0 ml-auto gap-2">
            {action}
            {latestRun?.href ? (
              <Button href={latestRun.href} variant="secondary" size="sm">
                View run
              </Button>
            ) : null}
          </span>
        </div>

        {repo || repoBranch ? (
          <span className="flex items-center gap-2 min-w-0 flex-wrap">
            <BranchVcsBadges repo={repo} branch={repoBranch} />
          </span>
        ) : null}
      </div>

      {latestRun ? (
        <div className="flex flex-col gap-2 min-w-0">
          <BranchRunCommit
            status={latestRun.status}
            href={latestRun.href}
            message={latestRun.commitMessage}
            author={latestRun.author}
            avatarUrl={latestRun.avatarUrl}
            sha={latestRun.sha}
            createdAt={latestRun.createdAt}
          />
          <RunMeta
            trigger={latestRun.trigger}
            prNumber={latestRun.prNumber}
            tag={latestRun.tag}
          />
        </div>
      ) : (
        <Text variant="subtext" theme="neutral">
          No runs yet
        </Text>
      )}

      {canExpand ? (
        <Expand
          id={`${card.branchId}-plan`}
          heading={<BranchPlanDots groups={planGroups!} />}
          isOpen
          isIconBeforeHeading
          headerClassName="px-1 py-1 rounded"
        >
          <div className="pl-7 pr-1 pb-2 pt-2">
            <MiniDeploymentView
              showRollout
              groups={planGroups!}
              installs={latestRunInstalls!}
            />
          </div>
        </Expand>
      ) : hasPlan ? (
        <BranchPlanDots groups={planGroups!} />
      ) : null}
    </Card>
  )
}
