export default {
  title: 'Log Stream/SSELogs',
}

import { SSELogs, LogsSkeleton } from './SSELogs'
import type { TOTELLog } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

const mockLog: TOTELLog = {
  id: 'log-1',
  timestamp: '2024-01-15T10:30:00Z',
  severity_number: 9,
  severity_text: 'INFO',
  body: 'Deploying application to cluster...',
  service_name: 'runner',
} as TOTELLog

const mockLog2: TOTELLog = {
  ...mockLog,
  id: 'log-2',
  timestamp: '2024-01-15T10:30:01Z',
  severity_number: 13,
  severity_text: 'WARN',
  body: 'Retrying connection attempt',
  service_name: 'system',
} as TOTELLog

const mockFilters: TLogFiltersProps = {
  search: '',
  setSearch: () => {},
  severity: [],
  setSeverity: () => {},
  service: [],
  setService: () => {},
  sortDirection: 'desc',
  setSortDirection: () => {},
  jobOutputOnly: false,
  setJobOutputOnly: () => {},
  availableServices: ['runner', 'system'],
  availableSeverities: ['INFO', 'WARN', 'ERROR'],
} as TLogFiltersProps

export const Default = () => (
  <SSELogs
    filteredLogs={[mockLog, mockLog2]}
    filters={mockFilters}
    activeLog={undefined}
    handleActiveLog={() => {}}
    loadMore={() => {}}
    hasMore={false}
    isLoading={false}
    isStreamOpen={false}
  />
)

export const Loading = () => (
  <SSELogs
    filteredLogs={[]}
    filters={mockFilters}
    activeLog={undefined}
    handleActiveLog={() => {}}
    loadMore={() => {}}
    hasMore={false}
    isLoading={true}
    isStreamOpen={false}
  />
)

export const WithActiveLog = () => (
  <SSELogs
    filteredLogs={[mockLog, mockLog2]}
    filters={mockFilters}
    activeLog={mockLog}
    handleActiveLog={() => {}}
    loadMore={() => {}}
    hasMore={false}
    isLoading={false}
    isStreamOpen={false}
  />
)

export const WithLoadMore = () => (
  <SSELogs
    filteredLogs={[mockLog]}
    filters={mockFilters}
    activeLog={undefined}
    handleActiveLog={() => {}}
    loadMore={() => {}}
    hasMore={true}
    isLoading={false}
    isStreamOpen={false}
  />
)

export const Skeleton = () => <LogsSkeleton />
