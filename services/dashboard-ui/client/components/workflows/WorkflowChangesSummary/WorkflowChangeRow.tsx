import { type ReactNode, useState } from 'react'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TStepChangeSummary } from '@/types'
import { cn } from '@/utils/classnames'
import { PLAN_TYPE_META, STATUS_META, hasChanges } from './change-summary-utils'

interface IWorkflowChangeRow {
  summary: TStepChangeSummary
  renderDetail: (summary: TStepChangeSummary) => ReactNode
}

const RowCounts = ({ summary }: { summary: TStepChangeSummary }) => {
  if (summary.status === 'generating') {
    return <Text variant="subtext" loading loadingWidth={10} />
  }
  if (summary.status === 'error') {
    return (
      <Text variant="subtext" theme="error">
        Failed to load
      </Text>
    )
  }
  if (!summary.hasDetail && !hasChanges(summary.counts)) {
    return (
      <Text variant="subtext" theme="neutral">
        No detail available
      </Text>
    )
  }
  return (
    <ChangeCountSummary
      added={summary.counts.create}
      updated={summary.counts.update}
      removed={summary.counts.delete}
      replaced={summary.counts.replace}
    />
  )
}

export const WorkflowChangeRow = ({
  summary,
  renderDetail,
}: IWorkflowChangeRow) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const planMeta = PLAN_TYPE_META[summary.planType]
  const statusMeta = STATUS_META[summary.status]
  const canExpand = summary.hasDetail && summary.status !== 'generating'

  const header = (
    <div className="flex items-center gap-3 min-w-0 flex-1">
      {canExpand ? (
        <Icon
          variant={isExpanded ? 'CaretDownIcon' : 'CaretRightIcon'}
          size={16}
          className="shrink-0"
        />
      ) : (
        <span className="w-4 shrink-0" />
      )}
      <Icon variant={planMeta.icon} size={18} className="shrink-0" />
      <div className="flex flex-col min-w-0">
        <Text variant="base" weight="strong" nowrap className="truncate">
          {summary.stepName}
        </Text>
        {summary.componentName ? (
          <Text variant="subtext" theme="neutral" nowrap className="truncate">
            {summary.componentName}
          </Text>
        ) : null}
      </div>
    </div>
  )

  const right = (
    <div className="flex items-center gap-4 shrink-0">
      <RowCounts summary={summary} />
      <Badge theme={statusMeta.theme} size="sm">
        {statusMeta.label}
      </Badge>
    </div>
  )

  return (
    <div className="flex flex-col border-b last:border-b-0">
      {canExpand ? (
        <button
          type="button"
          onClick={() => setIsExpanded((v) => !v)}
          aria-expanded={isExpanded}
          className={cn(
            'flex items-center justify-between gap-4 px-4 py-3 w-full text-left outline-none transition-all',
            'hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10'
          )}
        >
          {header}
          {right}
        </button>
      ) : (
        <div className="flex items-center justify-between gap-4 px-4 py-3 w-full">
          {header}
          {right}
        </div>
      )}

      {canExpand && isExpanded ? (
        <div className="px-4 pb-4 pt-1 bg-cool-grey-50 dark:bg-dark-grey-900">
          {renderDetail(summary)}
        </div>
      ) : null}
    </div>
  )
}
