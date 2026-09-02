import { type ReactNode, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TStepChangePlanType, TStepChangeSummary } from '@/types'
import {
  PLAN_TYPE_META,
  formatAggregate,
  hasChanges,
  sumCounts,
} from './change-summary-utils'
import { WorkflowChangeRow } from './WorkflowChangeRow'

export interface IWorkflowChangesSummaryLoadingStep {
  stepName: string
  componentName?: string
  planType: TStepChangePlanType
}

export interface IWorkflowChangesSummary {
  summaries: TStepChangeSummary[]
  isLoading?: boolean
  loadingSteps?: IWorkflowChangesSummaryLoadingStep[]
  renderDetail: (summary: TStepChangeSummary) => ReactNode
}

const isActionable = (summary: TStepChangeSummary): boolean =>
  hasChanges(summary.counts) ||
  summary.status === 'generating' ||
  summary.status === 'error'

const ChangesHeader = ({ summaries }: { summaries: TStepChangeSummary[] }) => {
  const totals = sumCounts(summaries)
  const changedSteps = summaries.filter((s) => hasChanges(s.counts))

  if (changedSteps.length === 0) {
    return (
      <div className="flex flex-col gap-1">
        <Text variant="base" weight="strong">
          No changes
        </Text>
        <Text variant="subtext" theme="neutral">
          {summaries.length} {summaries.length === 1 ? 'step' : 'steps'},
          nothing to apply
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-1">
      <Text variant="base" weight="strong">
        {formatAggregate(totals, changedSteps.length)}
      </Text>
      <Text variant="subtext" theme="neutral">
        {summaries.length} {summaries.length === 1 ? 'step' : 'steps'} in this
        workflow
      </Text>
    </div>
  )
}

const LoadingRow = ({ step }: { step: IWorkflowChangesSummaryLoadingStep }) => (
  <div className="flex items-center justify-between gap-4 px-4 py-3 border-b last:border-b-0">
    <div className="flex items-center gap-3 min-w-0 flex-1">
      <span className="w-4 shrink-0" />
      <Icon
        variant={PLAN_TYPE_META[step.planType].icon}
        size={18}
        className="shrink-0"
      />
      <div className="flex flex-col min-w-0">
        <Text variant="base" weight="strong" nowrap className="truncate">
          {step.stepName}
        </Text>
        {step.componentName ? (
          <Text variant="subtext" theme="neutral" nowrap className="truncate">
            {step.componentName}
          </Text>
        ) : null}
      </div>
    </div>
    <div className="flex items-center gap-4 shrink-0">
      <Text variant="subtext" loading loadingWidth={10} />
      <Badge loading loadingWidth={8} size="sm" />
    </div>
  </div>
)

const NoopGroup = ({
  summaries,
  renderDetail,
}: {
  summaries: TStepChangeSummary[]
  renderDetail: (summary: TStepChangeSummary) => ReactNode
}) => {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <div className="border rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setIsExpanded((v) => !v)}
        aria-expanded={isExpanded}
        className={cn(
          'flex items-center gap-3 px-4 py-3 w-full text-left outline-none transition-all',
          'hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10'
        )}
      >
        <Icon
          variant={isExpanded ? 'CaretDownIcon' : 'CaretRightIcon'}
          size={16}
          className="shrink-0"
        />
        <Text variant="base" weight="strong" theme="neutral">
          {summaries.length} {summaries.length === 1 ? 'step' : 'steps'} with no
          changes
        </Text>
      </button>

      {isExpanded ? (
        <div className="border-t">
          {summaries.map((summary) => (
            <WorkflowChangeRow
              key={summary.stepId}
              summary={summary}
              renderDetail={renderDetail}
            />
          ))}
        </div>
      ) : null}
    </div>
  )
}

export const WorkflowChangesSummary = ({
  summaries,
  isLoading,
  loadingSteps,
  renderDetail,
}: IWorkflowChangesSummary) => {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong" loading loadingWidth={30} />
        <div className="border rounded-lg overflow-hidden">
          {(loadingSteps ?? []).map((step, i) => (
            <LoadingRow key={`${step.stepName}-${i}`} step={step} />
          ))}
        </div>
      </div>
    )
  }

  if (summaries.length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan steps"
        emptyMessage="This workflow has no steps with plans to review."
      />
    )
  }

  const actionable = summaries.filter(isActionable)
  const noops = summaries.filter((s) => !isActionable(s))

  return (
    <div className="flex flex-col gap-4">
      <ChangesHeader summaries={summaries} />

      {actionable.length > 0 ? (
        <div className="border rounded-lg overflow-hidden">
          {actionable.map((summary) => (
            <WorkflowChangeRow
              key={summary.stepId}
              summary={summary}
              renderDetail={renderDetail}
            />
          ))}
        </div>
      ) : null}

      {noops.length > 0 ? (
        <NoopGroup summaries={noops} renderDetail={renderDetail} />
      ) : null}
    </div>
  )
}
