import { useLogs } from '@/hooks/use-logs'
import { LogSearch } from './LogSearch'
import { LogServiceFilter } from './LogServiceFilter'
import { LogSeverityFilter } from './LogSeverityFilter'
import { LogSort } from './LogSort'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

// Overload: with filters prop (new unified system)
interface LogFiltersWithFilters {
  filters: TLogFiltersProps
}

// Overload: without filters prop (legacy system)
interface LogFiltersWithoutFilters {
  filters?: never
}

type LogFiltersProps = LogFiltersWithFilters | LogFiltersWithoutFilters

export const LogFilters = ({ filters: externalFilters }: LogFiltersProps) => {
  // Only use legacy hook if no filters provided
  // eslint-disable-next-line
  const legacyFilters = externalFilters ? null : useLogs()
  const filters = externalFilters || legacyFilters!.filters

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 py-4">
      <LogSearch filters={filters} />

      <div className="flex items-center gap-4">
        <LogSort filters={filters} />
        <LogServiceFilter title="service" filters={filters} />
        <LogSeverityFilter title="severity" filters={filters} />
      </div>
    </div>
  )
}
