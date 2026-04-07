export default {
  title: 'Runners/ProcessCard',
}

import { ProcessCard, ProcessCardSkeleton } from './ProcessCard'

const mockProcess = {
  id: 'proc-123',
  type: 'runner',
  composite_status: { status: 'active' },
  version: '1.2.3',
  labels: ['primary'],
  started_at: '2024-01-15T08:00:00Z',
  warnings: [],
} as any

export const Default = () => (
  <ProcessCard
    process={mockProcess}
    runnerId="runner-456"
    isConnected={true}
    heartbeatCreatedAt="2024-01-15T12:00:00Z"
    configuredVersion="1.2.3"
    reportedVersion="1.2.3"
    healthchecks={[]}
  />
)

export const Disconnected = () => (
  <ProcessCard
    process={mockProcess}
    runnerId="runner-456"
    isConnected={false}
    configuredVersion="1.2.3"
    reportedVersion="1.2.2"
    healthchecks={[]}
  />
)

export const WithWarnings = () => (
  <ProcessCard
    process={{ ...mockProcess, warnings: ['Version mismatch detected'] }}
    runnerId="runner-456"
    isConnected={true}
    heartbeatCreatedAt="2024-01-15T12:00:00Z"
    configuredVersion="1.2.3"
    reportedVersion="1.2.2"
    healthchecks={[]}
  />
)

export const Skeleton = () => <ProcessCardSkeleton />
