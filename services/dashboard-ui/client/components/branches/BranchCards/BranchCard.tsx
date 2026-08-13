import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { BranchVcsBadges } from '@/components/branches/BranchVcsBadges'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'

export type TBranchCardData = {
  branchId: string
  name: string
  href: string
  managedBy?: string | null
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
  planSummary?: {
    groups: number
    installs: number
    hasSelector: boolean
  }
  action?: ReactNode
}

export const BranchCard = ({ card }: { card: TBranchCardData }) => {
  const { action, href, latestRun, managedBy, name, planSummary, repo, repoBranch } =
    card

  return (
    <Card className="gap-4 p-4 min-w-0">
      <div className="flex flex-col gap-1.5">
        <div className="flex items-start justify-between gap-2">
          <span className="flex items-center gap-2 min-w-0 flex-wrap">
            <Icon variant="GitBranchIcon" size={16} />
            <Link href={href} className="font-strong truncate max-w-56">
              {name}
            </Link>
            {managedBy ? (
              <Badge size="sm" theme={managedBy === 'config' ? 'brand' : 'default'}>
                {managedBy}
              </Badge>
            ) : null}
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
        {planSummary && planSummary.groups > 0 ? (
          <Text variant="subtext" theme="neutral">
            {planSummary.groups}{' '}
            {planSummary.groups === 1 ? 'group' : 'groups'}
            {planSummary.hasSelector
              ? ', installs matched by label'
              : ` · ${planSummary.installs} ${
                  planSummary.installs === 1 ? 'install' : 'installs'
                }`}
          </Text>
        ) : (
          <Text variant="subtext" theme="neutral">
            No deployment plan yet
          </Text>
        )}
      </div>
    </Card>
  )
}
