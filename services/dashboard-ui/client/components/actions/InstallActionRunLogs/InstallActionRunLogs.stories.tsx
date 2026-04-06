export default {
  title: 'Actions/InstallActionRunLogs',
}

import type { TActionConfig } from '@/types'
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
  log_attributes: { workflow_step_name: 'build' },
})) as any

export const Vertical = () => (
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
    filters={{}}
  />
)

export const Horizontal = () => (
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
    filters={{}}
  />
)

export const Empty = () => (
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
    filters={{}}
  />
)
