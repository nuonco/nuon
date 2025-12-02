import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface LogSortProps {
  filters: {
    handleSortToggle: TLogFiltersProps['handleSortToggle']
    sortStats: TLogFiltersProps['sortStats']
  }
}

export const LogSort = ({ filters }: LogSortProps) => {
  const { handleSortToggle, sortStats } = filters
  
  return (
    <Button onClick={handleSortToggle} variant="ghost">
      {sortStats.isNewestFirst ? 'Latest' : 'Oldest'}
      {sortStats.isNewestFirst ? (
        <Icon variant="SortDescending" />
      ) : (
        <Icon variant="SortAscending" />
      )}
    </Button>
  )
}
