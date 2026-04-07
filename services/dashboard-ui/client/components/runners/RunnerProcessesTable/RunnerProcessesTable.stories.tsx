export default {
  title: 'Runners/RunnerProcessesTable',
}

import { RunnerProcessesTable } from './RunnerProcessesTable'

const mockProcesses = [
  { id: 'proc-1', type: 'runner', composite_status: { status: 'active' }, version: '1.2.3', labels: ['primary'], started_at: '2024-01-15T08:00:00Z', created_at: '2024-01-15T08:00:00Z' },
  { id: 'proc-2', type: 'runner', composite_status: { status: 'shut-down' }, version: '1.2.2', labels: [], started_at: '2024-01-14T08:00:00Z', created_at: '2024-01-14T08:00:00Z' },
] as any[]

export const Default = () => (
  <RunnerProcessesTable processes={mockProcesses} isLoading={false} />
)

export const Loading = () => (
  <RunnerProcessesTable processes={[]} isLoading={true} />
)

export const Empty = () => (
  <RunnerProcessesTable processes={[]} isLoading={false} />
)
