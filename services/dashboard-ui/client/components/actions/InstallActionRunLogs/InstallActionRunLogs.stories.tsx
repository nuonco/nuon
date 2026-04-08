export default {
  title: 'Actions/InstallActionRunLogs',
}

import type { TActionConfig, TLogStream } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { InstallActionRunLogs } from './InstallActionRunLogs'

const noop = () => {}

const mockConfig: TActionConfig = {
  steps: [
    { id: 'step-1', name: 'build', idx: 0 },
    { id: 'step-2', name: 'deploy', idx: 1 },
    { id: 'step-3', name: 'verify', idx: 2 },
  ],
} as TActionConfig

const mockLogs = Array.from({ length: 5 }, (_, i) => ({
  id: `log-${i}`,
  body: `Log line ${i + 1}: running step...`,
  timestamp: new Date(Date.now() - i * 60000).toISOString(),
  severity_number: 9,
  severity_text: 'INFO',
  service_name: 'runner',
  scope_name: 'oteljob',
  log_attributes: { workflow_step_name: 'build' },
})) as any

const mockFilters: TLogFiltersProps = {
  selectedSeverities: new Set(['Info', 'Warn', 'Error', 'Fatal']),
  handleSeverityInputToggle: noop,
  handleSeverityButtonClick: noop,
  handleSeverityReset: noop,
  selectedServices: new Set(['runner']),
  availableServices: new Set(['runner']),
  handleServiceInputToggle: noop,
  handleServiceButtonClick: noop,
  handleServiceReset: noop,
  jobOutputOnly: false,
  handleJobOutputToggle: noop,
  searchQuery: '',
  sortDirection: 'desc',
  filteredLogs: mockLogs,
  handleSearchChange: noop,
  handleSortToggle: noop,
  handleSortChange: noop,
  filterStats: { selectedCount: 5, totalCount: 5 },
  sortStats: { direction: 'desc', isNewestFirst: true, isOldestFirst: false },
  severityStats: { selectedCount: 4, totalCount: 4 },
  serviceStats: { selectedCount: 1, totalCount: 1, isAllSelected: true },
} as unknown as TLogFiltersProps

const mockLogStream: TLogStream = {
  id: 'log-stream-1',
  org_id: 'org-mock-001',
  open: false,
} as TLogStream

const mockUnifiedContext = {
  logs: mockLogs,
  isLoading: false,
  error: null,
  connectionState: 'disconnected' as const,
  loadMore: noop,
  hasMore: false,
  isStreamOpen: false,
}

const Providers = ({ children }: { children: React.ReactNode }) => (
  <LogStreamContext.Provider value={{ logStream: mockLogStream, refresh: noop }}>
    <UnifiedLogsContext.Provider value={mockUnifiedContext}>
      {children}
    </UnifiedLogsContext.Provider>
  </LogStreamContext.Provider>
)

export const Vertical = () => (
  <Providers>
    <InstallActionRunLogs
      actionConfig={mockConfig}
      layout="vertical"
      filteredLogs={mockLogs}
      loadMore={noop}
      hasMore={false}
      isLoading={false}
      isStreamOpen={false}
      activeLog={undefined}
      handleActiveLog={noop}
      filters={mockFilters}
    />
  </Providers>
)

export const Horizontal = () => (
  <Providers>
    <InstallActionRunLogs
      actionConfig={mockConfig}
      layout="horizontal"
      filteredLogs={mockLogs}
      loadMore={noop}
      hasMore={true}
      isLoading={false}
      isStreamOpen={false}
      activeLog={undefined}
      handleActiveLog={noop}
      filters={mockFilters}
    />
  </Providers>
)

export const Loading = () => (
  <Providers>
    <InstallActionRunLogs
      actionConfig={mockConfig}
      layout="vertical"
      filteredLogs={[]}
      loadMore={noop}
      hasMore={false}
      isLoading={true}
      isStreamOpen={false}
      activeLog={undefined}
      handleActiveLog={noop}
      filters={{ ...mockFilters, filteredLogs: [] }}
    />
  </Providers>
)
