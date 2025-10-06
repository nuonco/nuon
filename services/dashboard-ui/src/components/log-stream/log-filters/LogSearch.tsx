import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { useLogs } from '@/hooks/use-logs'

export const LogSearch = () => {
  const {
    filters: { filterStats, handleSearchChange, searchQuery },
  } = useLogs()

  return (
    <div className="flex items-center gap-4">
      <SearchInput
        placeholder="Search logs..."
        value={searchQuery}
        onChange={handleSearchChange}
      />
      <div className="flex items-center gap-6">
        <Text variant="subtext" theme="neutral">
          {filterStats?.selectedCount} of {filterStats?.totalCount} logs
        </Text>
      </div>
    </div>
  )
}
