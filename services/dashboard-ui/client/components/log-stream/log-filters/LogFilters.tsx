import { LogFiltersDropdown } from './LogFiltersDropdown'
import { LogSearch } from './LogSearch'
import { LogSeverityDropdown } from './LogSeverityDropdown'
import { DownloadLogsButton } from '../DownloadLogs'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

interface LogFiltersProps {
  filters: TLogFiltersProps
}

export const LogFilters = ({ filters }: LogFiltersProps) => {
  return (
    <div className="flex items-center justify-between gap-4 py-4 w-full">
      <LogSearch filters={filters} />

      <div className="flex items-center gap-2">
        <LogSeverityDropdown filters={filters} />
        <LogFiltersDropdown filters={filters} />
        <DownloadLogsButton includeSystemLogs={filters.includeSystemLogs} />
      </div>
    </div>
  )
}
