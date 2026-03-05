import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface LogSearchProps {
  filters: {
    filterStats: TLogFiltersProps['filterStats']
    handleSearchChange: TLogFiltersProps['handleSearchChange'] 
    searchQuery: TLogFiltersProps['searchQuery']
  }
}

export const LogSearch = ({ filters }: LogSearchProps) => {
  const { filterStats, handleSearchChange, searchQuery } = filters

  return (
    <div className="flex items-center gap-3">
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
