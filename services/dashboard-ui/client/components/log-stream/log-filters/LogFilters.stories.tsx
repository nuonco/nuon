export default {
  title: 'LogStream/LogFilters',
}

import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { LogFilters } from './LogFilters'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import type { TLogStream } from '@/types'

const noop = () => {}

const mockFilters: TLogFiltersProps = {
  selectedSeverities: new Set(['Info', 'Warn', 'Error', 'Fatal']),
  availableSeverities: new Set(['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']),
  handleSeverityInputToggle: noop,
  handleSeverityButtonClick: noop,
  handleSeverityReset: noop,
  selectedServices: new Set(['api', 'runner']),
  availableServices: new Set(['api', 'runner']),
  handleServiceInputToggle: noop,
  handleServiceButtonClick: noop,
  handleServiceReset: noop,
  selectedScopes: new Set<string>(),
  availableScopes: new Set<string>(['oteljob', 'system']),
  handleScopeInputToggle: noop,
  handleScopeButtonClick: noop,
  handleScopeReset: noop,
  availableTools: new Set<string>(['helm', 'terraform', 'kubernetes_manifest']),
  tool: '',
  setTool: noop,
  helmReleaseName: '',
  setHelmReleaseName: noop,
  helmOperation: '',
  setHelmOperation: noop,
  tfWorkspaceID: '',
  setTfWorkspaceID: noop,
  tfOperation: '',
  setTfOperation: noop,
  k8sKind: '',
  setK8sKind: noop,
  k8sNamespace: '',
  setK8sNamespace: noop,
  k8sName: '',
  setK8sName: noop,
  jobOutputOnly: false,
  handleJobOutputToggle: noop,
  searchQuery: '',
  sortDirection: 'desc',
  filteredLogs: [],
  handleSearchChange: noop,
  handleSortToggle: noop,
  handleSortChange: noop,
  filterStats: { selectedCount: 0, totalCount: 0 },
  sortStats: { direction: 'desc', isNewestFirst: true, isOldestFirst: false },
  severityStats: { selectedCount: 4, totalCount: 6, isDefault: true },
  serviceStats: { selectedCount: 2, totalCount: 2, isAllSelected: false },
  scopeStats: { selectedCount: 0, totalCount: 2, isAllSelected: true },
  isFiltered: false,
  handleResetAll: noop,
  serverFilters: {},
}

const mockContextLive = {
  logs: [],
  isLoading: false,
  error: null,
  connectionState: 'connected' as const,
  loadMore: noop,
  hasMore: false,
  isStreamOpen: true,
}

const mockLogStream: TLogStream = {
  id: 'log-stream-1',
  org_id: 'org-mock-001',
  open: false,
} as TLogStream

const mockLogStreamContext = {
  logStream: mockLogStream,
  refresh: () => {},
}

const mockContextStatic = {
  ...mockContextLive,
  connectionState: 'disconnected' as const,
  isStreamOpen: false,
  hasMore: true,
}

export const LiveStream = () => (
  <UnifiedLogsContext.Provider value={mockContextLive}>
    <LogFilters filters={mockFilters} />
  </UnifiedLogsContext.Provider>
)

export const StaticLogs = () => (
  <LogStreamContext.Provider value={mockLogStreamContext}>
    <UnifiedLogsContext.Provider value={mockContextStatic}>
      <LogFilters filters={mockFilters} />
    </UnifiedLogsContext.Provider>
  </LogStreamContext.Provider>
)

export const JobOutputOnly = () => (
  <LogStreamContext.Provider value={mockLogStreamContext}>
    <UnifiedLogsContext.Provider value={mockContextStatic}>
      <LogFilters filters={{ ...mockFilters, jobOutputOnly: true }} />
    </UnifiedLogsContext.Provider>
  </LogStreamContext.Provider>
)
