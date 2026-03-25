import { LogJobOutputFilter } from './LogJobOutputFilter'
import { LogSearch } from './LogSearch'
import { LogServiceFilter } from './LogServiceFilter'
import { LogSeverityFilter } from './LogSeverityFilter'
import { LogSort } from './LogSort'
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
        <LogSort filters={filters} />
        <LogJobOutputFilter filters={filters} />
        <LogServiceFilter title="service" filters={filters} />
        <LogSeverityFilter title="severity" filters={filters} />
        {!isStreamOpen && <DownloadLogsButton />}
      </div>
    </div>
  )
}
