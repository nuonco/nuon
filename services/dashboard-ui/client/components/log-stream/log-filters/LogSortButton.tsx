import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface ILogSortButton {
  filters: TLogFiltersProps
}

export const LogSortButton = ({ filters }: ILogSortButton) => {
  const { sortStats, handleSortToggle } = filters
  const newestFirst = sortStats.isNewestFirst

  return (
    <Button
      aria-label={
        newestFirst
          ? 'Sorted latest first. Activate to sort oldest first.'
          : 'Sorted oldest first. Activate to sort latest first.'
      }
      title={newestFirst ? 'Latest first' : 'Oldest first'}
      onClick={handleSortToggle}
    >
      <Icon
        variant={newestFirst ? 'ArrowDownIcon' : 'ArrowUpIcon'}
        size="16"
      />
      <span className="@max-[64rem]:hidden">
        {newestFirst ? 'Latest' : 'Oldest'}
      </span>
    </Button>
  )
}
