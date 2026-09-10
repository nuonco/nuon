import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { BranchRunComparisonCard } from '@/components/branches/BranchRunComparisonCard'
import { cn } from '@/utils/classnames'
import type { TBranchRunComparisonRunSummary } from '@/lib/ctl-api/apps/branches/get-branch-run-comparison'

export interface IBranchRunComparisonRuns {
  orgId: string
  appId: string
  branchId: string
  baseRun?: TBranchRunComparisonRunSummary | null
  headRun?: TBranchRunComparisonRunSummary | null
  repoSlug?: string
  currentGithubHref?: string
}

const runDetailHref = (
  orgId: string,
  appId: string,
  branchId: string,
  workflowId?: string
) => {
  if (!workflowId) {
    return undefined
  }
  return `/${orgId}/apps/${appId}/branches/${branchId}/runs/${workflowId}`
}

export const BranchRunComparisonRuns = ({
  orgId,
  appId,
  branchId,
  baseRun,
  headRun,
  repoSlug,
  currentGithubHref,
}: IBranchRunComparisonRuns) => {
  const hasBaseline = !!baseRun?.id

  return (
    <div className="flex flex-col gap-2">
      {!hasBaseline ? (
        <Text variant="subtext" theme="neutral">
          First run on this branch — no previous baseline to compare against.
        </Text>
      ) : null}

      <div
        className={cn('grid grid-cols-1 gap-3 items-stretch', {
          'md:grid-cols-[1fr_auto_1fr]': hasBaseline,
        })}
      >
        {hasBaseline ? (
          <>
            <BranchRunComparisonCard
              label="Previous run"
              run={baseRun}
              runHref={runDetailHref(orgId, appId, branchId, baseRun?.workflow_id)}
              repoSlug={repoSlug}
            />
            <div className="hidden md:flex items-center justify-center px-1">
              <Icon variant="ArrowRightIcon" size={20} className="text-cool-grey-400" />
            </div>
          </>
        ) : null}

        <BranchRunComparisonCard
          label="Current run"
          run={headRun}
          repoSlug={repoSlug}
          githubHeaderHref={currentGithubHref}
        />
      </div>
    </div>
  )
}
