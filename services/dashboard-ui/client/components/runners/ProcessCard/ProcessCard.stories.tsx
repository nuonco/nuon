export default {
  title: 'Runners/ProcessCard',
}

import type { TRunnerHealthCheck } from '@/types'
import { ProcessCard, ProcessCardSkeleton } from './ProcessCard'

const now = new Date()
const minutesAgo = (m: number) => new Date(now.getTime() - m * 60000).toISOString()
const hoursAgo = (h: number) => new Date(now.getTime() - h * 3600000).toISOString()

const baseProcess = {
  id: 'rpru3x3uwgappsn730x5lbz8bf',
  type: 'install',
  composite_status: { status: 'active' },
  version: 'de111c8',
  labels: [] as string[],
  started_at: hoursAgo(6),
  warnings: [] as string[],
} as any

const mngProcess = {
  ...baseProcess,
  id: 'rpr4cg772hl2ekwvsxla8b4xwc',
  type: 'mng',
  version: 'development',
  labels: ['Local Runner'],
} as any

const healthyChecks: TRunnerHealthCheck[] = Array.from({ length: 30 }, (_, i) => ({
  id: `hc-${i}`,
  status_code: 0,
  minute_bucket: minutesAgo(30 - i),
})) as TRunnerHealthCheck[]

const mixedChecks: TRunnerHealthCheck[] = Array.from({ length: 30 }, (_, i) => ({
  id: `hc-${i}`,
  status_code: i >= 20 && i <= 25 ? 1 : i >= 26 ? 900 : 0,
  minute_bucket: minutesAgo(30 - i),
})) as TRunnerHealthCheck[]

const unknownChecks: TRunnerHealthCheck[] = Array.from({ length: 30 }, (_, i) => ({
  id: `hc-${i}`,
  status_code: 900,
  minute_bucket: minutesAgo(30 - i),
})) as TRunnerHealthCheck[]

export const ActiveConnected = () => (
  <ProcessCard
    process={baseProcess}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={healthyChecks}
  />
)

export const ActiveDisconnected = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      composite_status: { status: 'offline' },
    }}
    isConnected={false}
    heartbeatCreatedAt={minutesAgo(3)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={mixedChecks}
  />
)

export const VersionMismatch = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      warnings: [
        'Reported runner version (de111c8) does not match configured version (ab98f21). Please update the runner to the correct version.',
      ],
    }}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="ab98f21"
    reportedVersion="de111c8"
    healthchecks={healthyChecks}
  />
)

export const Initializing = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      warnings: [
        'This runner is still initializing and will not process jobs until its first health check',
      ],
    }}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={[]}
  />
)

export const Pending = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      composite_status: { status: 'pending' },
      labels: ['Local Runner'],
      started_at: hoursAgo(21),
    }}
    isConnected={false}
    configuredVersion="main"
    reportedVersion="development"
    healthchecks={[]}
  />
)

export const MngProcess = () => (
  <ProcessCard
    process={mngProcess}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="latest"
    reportedVersion="development"
    healthchecks={healthyChecks}
  />
)

export const MngVersionMismatch = () => (
  <ProcessCard
    process={{
      ...mngProcess,
      warnings: [
        'Reported runner version (development) does not match configured version (v1.2.0). Please update the runner to the correct version.',
      ],
    }}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="v1.2.0"
    reportedVersion="development"
    healthchecks={healthyChecks}
  />
)

export const MultipleWarnings = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      composite_status: { status: 'offline' },
      warnings: [
        'Runner is offline and will be marked inactive in 5 minutes',
        'Reported runner version (de111c8) does not match configured version (ab98f21). Please update the runner to the correct version.',
      ],
    }}
    isConnected={false}
    heartbeatCreatedAt={minutesAgo(3)}
    configuredVersion="ab98f21"
    reportedVersion="de111c8"
    healthchecks={mixedChecks}
  />
)

export const NoHealthData = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      labels: ['Local Runner'],
    }}
    isConnected={false}
    configuredVersion="main"
    reportedVersion="-"
    healthchecks={[]}
  />
)

export const AllUnknownHealth = () => (
  <ProcessCard
    process={baseProcess}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={unknownChecks}
  />
)

export const FewHealthChecks = () => (
  <ProcessCard
    process={baseProcess}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={
      Array.from({ length: 5 }, (_, i) => ({
        id: `hc-${i}`,
        status_code: 0,
        minute_bucket: minutesAgo(4 - i),
      })) as TRunnerHealthCheck[]
    }
  />
)

export const PendingShutdown = () => (
  <ProcessCard
    process={{
      ...baseProcess,
      composite_status: { status: 'pending-shutdown', status_description: 'Shutting down for version update' },
      warnings: ['Shutting down for version update'],
    }}
    isConnected
    heartbeatCreatedAt={minutesAgo(0)}
    configuredVersion="de111c8"
    reportedVersion="de111c8"
    healthchecks={healthyChecks}
  />
)

export const SideBySide = () => (
  <div className="grid grid-cols-2 gap-6 items-start max-w-5xl">
    <ProcessCard
      process={{
        ...baseProcess,
        warnings: [
          'This runner is still initializing and will not process jobs until its first health check',
        ],
      }}
      isConnected
      heartbeatCreatedAt={minutesAgo(0)}
      configuredVersion="main"
      reportedVersion="development"
      healthchecks={[]}
    />
    <ProcessCard
      process={{
        ...mngProcess,
        composite_status: { status: 'pending' },
      }}
      isConnected={false}
      configuredVersion="main"
      reportedVersion="development"
      healthchecks={[]}
    />
  </div>
)

export const Skeleton = () => <ProcessCardSkeleton />
