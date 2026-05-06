import { LogJobOutputFilter } from './LogJobOutputFilter'
import { LogScopeFilter } from './LogScopeFilter'
import { LogSearch } from './LogSearch'
import { LogServiceFilter } from './LogServiceFilter'
import { LogSeverityFilter } from './LogSeverityFilter'
import { LogSort } from './LogSort'
import { LogToolFilter } from './LogToolFilter'
import { DownloadLogsButton } from '../DownloadLogs'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { useUnifiedLogData } from '@/hooks/use-logs'

interface LogFiltersProps {
  filters: TLogFiltersProps
}

export const LogFilters = ({ filters }: LogFiltersProps) => {
  const { isStreamOpen } = useUnifiedLogData()
  return (
    <div className="flex flex-wrap items-center justify-between gap-4 py-4 w-full">
      <LogSearch filters={filters} />

      <div className="flex items-center justify-end gap-4">
        <LogJobOutputFilter filters={filters} />
        <LogSort filters={filters} />
        <LogToolFilter filters={filters} />
        <LogScopeFilter filters={filters} />
        <LogServiceFilter title="Service" filters={filters} />
        <LogSeverityFilter title="Severity" filters={filters} />
        {!isStreamOpen && <DownloadLogsButton />}
      </div>
    </div>
  )
}
