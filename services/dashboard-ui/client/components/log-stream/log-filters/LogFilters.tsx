import { Icon } from '@/components/common/Icon'
import { ToggleButton } from '@/components/common/ToggleButton'
import type { TLogFiltersProps, ViewMode } from '@/hooks/use-log-filters'
import { LogFiltersDropdown } from './LogFiltersDropdown'
import { LogSearch } from './LogSearch'
import { LogSeverityDropdown } from './LogSeverityDropdown'
import { DownloadLogsButton } from '../DownloadLogs'

const viewModeOptions: { value: ViewMode; label: React.ReactNode; ariaLabel: string; title: string }[] = [
  {
    value: 'structured',
    label: (
      <>
        <Icon variant="GridNineIcon" size="16" />
        <span className="@max-[64rem]:hidden">Structured</span>
      </>
    ),
    title: 'Structured',
    ariaLabel: 'Structured log view',
  },
  {
    value: 'raw',
    label: (
      <>
        <Icon variant="TerminalWindowIcon" size="16" />
        <span className="@max-[64rem]:hidden">Raw</span>
      </>
    ),
    title: 'Raw',
    ariaLabel: 'Raw output view',
  },
]

interface LogFiltersProps {
  filters: TLogFiltersProps
}

export const LogFilters = ({ filters }: LogFiltersProps) => {
  return (
    <div className="flex items-center flex-wrap justify-between gap-4 py-4 w-full">
      <LogSearch filters={filters} />

      <div className="flex items-center gap-2">
        <LogSeverityDropdown filters={filters} />
        <LogFiltersDropdown filters={filters} />
        <ToggleButton
          size="md"
          options={viewModeOptions}
          value={filters.viewMode}
          onChange={filters.handleViewModeChange}
        />
        <DownloadLogsButton includeSystemLogs={filters.includeSystemLogs} />
      </div>
    </div>
  )
}
