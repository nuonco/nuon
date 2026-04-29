export default {
  title: 'Runners/ProcessManagementDropdown',
}

import { ProcessManagementDropdown } from './ProcessManagementDropdown'

const mockProcess = {
  id: 'proc-123',
  type: 'runner',
  composite_status: { status: 'active' },
  log_stream_id: 'log-1',
} as any

export const Default = () => (
  <ProcessManagementDropdown
    process={mockProcess}
    runnerId="runner-456"
    onViewSystemLogs={() => {}}
  />
)

export const InactiveProcess = () => (
  <ProcessManagementDropdown
    process={{ ...mockProcess, composite_status: { status: 'shut-down' } }}
    runnerId="runner-456"
  />
)
