export default {
  title: 'LogStream/SSELogs',
}

import { SSELogs, LogsSkeleton } from './SSELogs'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import { LogStreamContext } from '@/providers/log-stream-provider'
import type { TOTELLog, TLogStream } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const noop = () => {}

const mockLogs: TOTELLog[] = [
  {
    id: 'log-1',
    timestamp: '2024-01-15T10:30:00.000Z',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Deploying application to cluster...',
    service_name: 'runner',
    scope_name: 'oteljob',
  },
  {
    id: 'log-2',
    timestamp: '2024-01-15T10:30:01.000Z',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Pulling image: 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:latest',
    service_name: 'runner',
    scope_name: 'oteljob',
  },
  {
    id: 'log-3',
    timestamp: '2024-01-15T10:30:02.500Z',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'Retrying connection attempt (2/3)',
    service_name: 'api',
    scope_name: 'system',
  },
  {
    id: 'log-4',
    timestamp: '2024-01-15T10:30:04.000Z',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Helm release updated successfully',
    service_name: 'runner',
    scope_name: 'oteljob',
  },
  {
    id: 'log-5',
    timestamp: '2024-01-15T10:30:05.000Z',
    severity_number: 17,
    severity_text: 'Error',
    body: 'Failed to connect to database: connection timeout after 30s',
    service_name: 'api',
    scope_name: 'system',
  },
] as TOTELLog[]

const mockFilters: TLogFiltersProps = {
  selectedSeverities: new Set(['Info', 'Warn', 'Error', 'Fatal']),
  handleSeverityInputToggle: noop,
  handleSeverityButtonClick: noop,
  handleSeverityReset: noop,
  selectedServices: new Set(['api', 'runner']),
  availableServices: new Set(['api', 'runner']),
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
  serviceStats: { selectedCount: 2, totalCount: 2, isAllSelected: false },
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

export const Default = () => (
  <Providers>
    <SSELogs
      filteredLogs={mockLogs}
      filters={mockFilters}
      activeLog={undefined}
      handleActiveLog={noop}
      loadMore={noop}
      hasMore={false}
      isLoading={false}
      isStreamOpen={false}
    />
  </Providers>
)

export const WithActiveLog = () => (
  <Providers>
    <SSELogs
      filteredLogs={mockLogs}
      filters={mockFilters}
      activeLog={mockLogs[0]}
      handleActiveLog={noop}
      loadMore={noop}
      hasMore={false}
      isLoading={false}
      isStreamOpen={false}
    />
  </Providers>
)

export const WithLoadMore = () => (
  <Providers>
    <SSELogs
      filteredLogs={mockLogs}
      filters={mockFilters}
      activeLog={undefined}
      handleActiveLog={noop}
      loadMore={noop}
      hasMore={true}
      isLoading={false}
      isStreamOpen={false}
    />
  </Providers>
)

export const Loading = () => (
  <Providers>
    <SSELogs
      filteredLogs={[]}
      filters={mockFilters}
      activeLog={undefined}
      handleActiveLog={noop}
      loadMore={noop}
      hasMore={false}
      isLoading={true}
      isStreamOpen={false}
    />
  </Providers>
)

export const Skeleton = () => <LogsSkeleton />
