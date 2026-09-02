import { Text } from '@/components/common/Text'
import type { TStepChangeSummary } from '@/types'
import { cn } from '@/utils/classnames'
import { formatAggregate, hasChanges, sumCounts } from './change-summary-utils'

interface IChangesAggregate {
  summaries: TStepChangeSummary[]
  loading?: boolean
  className?: string
}

export const ChangesAggregate = ({
  summaries,
  loading,
  className,
}: IChangesAggregate) => {
  if (loading) {
    return (
      <Text variant="subtext" loading loadingWidth={40} className={className} />
    )
  }

  const totals = sumCounts(summaries)
  const changedSteps = summaries.filter((s) => hasChanges(s.counts))

  if (changedSteps.length === 0) {
    return (
      <Text
        variant="subtext"
        theme="success"
        weight="strong"
        className={cn('flex items-center gap-1', className)}
      >
        No changes across {summaries.length}{' '}
        {summaries.length === 1 ? 'step' : 'steps'}
      </Text>
    )
  }

  return (
    <Text variant="subtext" family="mono" className={className}>
      {formatAggregate(totals, changedSteps.length)}
    </Text>
  )
}
