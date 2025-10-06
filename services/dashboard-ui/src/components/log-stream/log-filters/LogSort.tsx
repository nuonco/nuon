import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useLogs } from '@/hooks/use-logs'

export const LogSort = () => {
  const {
    filters: { handleSortToggle, sortStats },
  } = useLogs()
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
