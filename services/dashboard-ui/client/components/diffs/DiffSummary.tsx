import { cn } from '@/utils/classnames'
import {
  DIFF_OPERATIONS,
  type IPlanDiffSummary,
  type TDiffOperation,
} from '@/lib/diffs'
import { Text } from '@/components/common/Text'

const LABELS: Record<TDiffOperation, string> = {
  create: 'to create',
  update: 'to update',
  replace: 'to replace',
  delete: 'to delete',
  read: 'to read',
  'no-op': 'unchanged',
}

const COUNT_CLASSES: Record<TDiffOperation, string> = {
  create: 'text-diff-add',
  update: 'text-diff-change',
  replace: 'text-primary-600 dark:text-primary-400',
  delete: 'text-diff-remove',
  read: 'text-cool-grey-800 dark:text-white/70',
  'no-op': 'text-cool-grey-600 dark:text-white/70',
}

export interface IDiffSummary {
  summary: IPlanDiffSummary
  operations?: readonly TDiffOperation[]
  className?: string
}

export const DiffSummary = ({
  summary,
  operations = DIFF_OPERATIONS,
  className,
}: IDiffSummary) => (
  <div className={cn('flex flex-wrap items-center gap-x-5 gap-y-2', className)}>
    {operations.map((operation) => (
      <span key={operation} className="flex items-baseline gap-1.5">
        <Text
          variant="body"
          weight="stronger"
          className={COUNT_CLASSES[operation]}
        >
          {summary[operation]}
        </Text>
        <Text variant="subtext" theme="neutral">
          {LABELS[operation]}
        </Text>
      </span>
    ))}
  </div>
)
