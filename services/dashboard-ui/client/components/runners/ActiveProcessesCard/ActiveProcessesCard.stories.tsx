export default {
  title: 'Runners/ActiveProcessesCard',
}

import { ActiveProcessesCard, type TProcessWithHeartbeat } from './ActiveProcessesCard'
import type { TRunnerProcess } from '@/types'

const mockProcesses: TProcessWithHeartbeat[] = [
  {
    process: {
      id: 'proc_1',
      type: 'install',
      started_at: new Date(Date.now() - 7200000).toISOString(),
      version: 'v1.2.3',
      composite_status: { status: 'active' },
      labels: ['primary'],
    } as TRunnerProcess,
    heartbeat: {
      version: 'v1.2.3',
      created_at: new Date().toISOString(),
    },
  },
  {
    process: {
      id: 'proc_2',
      type: 'build',
      started_at: new Date(Date.now() - 600000).toISOString(),
      version: 'v1.2.3',
      composite_status: { status: 'offline' },
      labels: [],
    } as unknown as TRunnerProcess,
    heartbeat: {
      version: 'v1.2.3',
      created_at: new Date(Date.now() - 60000).toISOString(),
    },
  },
]

export const Default = () => <ActiveProcessesCard processes={mockProcesses} />

export const Empty = () => <ActiveProcessesCard processes={[]} />

export const Loading = () => <ActiveProcessesCard isLoading />
