import { LogFiltersDropdown } from './LogFiltersDropdown'
import { LogSearch } from './LogSearch'
import { LogSeverityDropdown } from './LogSeverityDropdown'
import { DownloadLogsButton } from '../DownloadLogs'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { useUnifiedLogData } from '@/hooks/use-logs'

interface LogFiltersProps {
  filters: TLogFiltersProps
}

export const LogFilters = ({ filters }: LogFiltersProps) => {
  const { isStreamOpen } = useUnifiedLogData()
  return (
    <div className="flex items-center justify-between gap-4 py-4 w-full">
      <LogSearch filters={filters} />

      <div className="flex items-center gap-2">
        <LogSeverityDropdown filters={filters} />
        <LogFiltersDropdown filters={filters} />
        {!isStreamOpen && <DownloadLogsButton includeSystemLogs={filters.includeSystemLogs} />}
      </div>
    </div>
  )
}
