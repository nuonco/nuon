export default {
  title: 'Features / Logs / Log filters',
}

import { LogStreamContext } from '@/providers/log-stream-provider'
import { LogFilters } from './LogFilters'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const noop = () => {}

const mockFilters: TLogFiltersProps = {
  selectedSeverities: new Set(['Info', 'Warn', 'Error', 'Fatal']),
  availableSeverities: new Set(['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']),
  handleSeverityInputToggle: noop,
  handleSeverityButtonClick: noop,
  handleSeverityReset: noop,
  includeSystemLogs: false,
  handleSystemLogsToggle: noop,
  availableTools: new Set<string>(['helm', 'terraform', 'kubernetes_manifest']),
  tool: '',
  setTool: noop,
  helmReleaseName: '',
  helmOperation: '',
  tfWorkspaceID: '',
  tfOperation: '',
  k8sKind: '',
  k8sNamespace: '',
  k8sName: '',
  viewMode: 'structured',
  handleViewModeChange: noop,
  searchQuery: '',
  sortDirection: 'desc',
  filteredLogs: [],
  handleSearchChange: noop,
  handleSortToggle: noop,
  filterStats: { selectedCount: 0, totalCount: 0 },
  sortStats: { direction: 'desc', isNewestFirst: true, isOldestFirst: false },
  severityStats: { selectedCount: 4, totalCount: 6, isDefault: true },
  isFiltered: false,
  handleResetAll: noop,
}

const mockLogStreamContext = {
  logs: [],
  logStreamId: 'log-stream-1',
  isLoading: false,
  isCatchingUp: false,
  error: null,
  connectionState: 'connected' as const,
}

export const LiveStream = () => (
  <LogStreamContext.Provider value={mockLogStreamContext}>
    <LogFilters filters={mockFilters} />
  </LogStreamContext.Provider>
)

export const StaticLogs = () => (
  <LogStreamContext.Provider value={{ ...mockLogStreamContext, connectionState: 'disconnected' }}>
    <LogFilters filters={mockFilters} />
  </LogStreamContext.Provider>
)

export const WithSystemLogs = () => (
  <LogStreamContext.Provider value={{ ...mockLogStreamContext, connectionState: 'disconnected' }}>
    <LogFilters filters={{ ...mockFilters, includeSystemLogs: true }} />
  </LogStreamContext.Provider>
)
