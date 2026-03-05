import { LogSearch } from './LogSearch'
import { LogServiceFilter } from './LogServiceFilter'
import { LogSeverityFilter } from './LogSeverityFilter'
import { LogSort } from './LogSort'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface LogFiltersProps {
  filters: TLogFiltersProps
}

export const LogFilters = ({ filters }: LogFiltersProps) => {
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
