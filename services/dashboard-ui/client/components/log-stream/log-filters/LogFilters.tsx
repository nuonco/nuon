import { LogJobOutputFilter } from './LogJobOutputFilter'
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
    <div className="flex flex-wrap items-center justify-between gap-4 py-4 w-full">
      <LogSearch filters={filters} />

      <div className="flex items-center justify-end gap-4">
        <LogSort filters={filters} />
        <LogJobOutputFilter filters={filters} />
        <LogServiceFilter title="service" filters={filters} />
        <LogSeverityFilter title="severity" filters={filters} />
      </div>
    </div>
  )
}
