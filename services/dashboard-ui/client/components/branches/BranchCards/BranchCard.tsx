import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'

export type TBranchCardData = {
  branchId: string
  name: string
  href: string
  managedBy?: string | null
  repo?: string
  repoBranch?: string
  latestRun?: {
    href: string
    status: string
    commitMessage?: string
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
    <Card className="gap-3 p-4">
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

      {repo ? (
        <span className="flex items-center gap-1.5 min-w-0">
          <Icon variant="GitHub" size={14} />
          <Text variant="subtext" theme="neutral" className="truncate">
            {repo}
            {repoBranch ? `:${repoBranch}` : ''}
          </Text>
        </span>
      ) : null}

      {latestRun ? (
        <div className="flex items-center gap-2 min-w-0">
          <Status status={latestRun.status} variant="badge" />
          <Link href={latestRun.href} className="truncate text-sm min-w-0 flex-1">
            {latestRun.commitMessage || 'View run'}
          </Link>
          {latestRun.createdAt ? (
            <Time
              variant="subtext"
              time={latestRun.createdAt}
              format="relative"
            />
          ) : null}
        </div>
      ) : (
        <Text variant="subtext" theme="neutral">
          No runs yet
        </Text>
      )}

      {planSummary && planSummary.groups > 0 ? (
        <Text variant="subtext" theme="neutral">
          Deployment plan: {planSummary.groups}{' '}
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
    </Card>
  )
}
