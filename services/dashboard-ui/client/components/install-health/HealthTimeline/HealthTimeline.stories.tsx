export default {
  title: 'InstallHealth/HealthTimeline',
}

import type { ReactNode } from 'react'
import { DateTime } from 'luxon'
import { Card } from '@/components/common/Card'
import type {
  THealthTimelineDay,
  TInstallComponentHealthTransition,
  TInstallHealthTimelineComponent,
} from '@/types'
import { HealthTimeline } from './HealthTimeline'

function buildDaily(days: number): THealthTimelineDay[] {
  return Array.from({ length: days }, (_, i) => {
    const date = DateTime.now()
      .minus({ days: days - 1 - i })
      .toISODate()!

    if (i % 19 === 0) {
      return {
        date,
        health: '',
        unhealthy_seconds: 0,
        degraded_seconds: 0,
        unknown_seconds: 86400,
        observed_seconds: 0,
      }
    }
    if (i % 23 === 0) {
      return {
        date,
        health: 'unhealthy',
        unhealthy_seconds: 5400,
        degraded_seconds: 0,
        unknown_seconds: 0,
        observed_seconds: 86400,
      }
    }
    if (i % 11 === 0) {
      return {
        date,
        health: 'degraded',
        unhealthy_seconds: 0,
        degraded_seconds: 1800,
        unknown_seconds: 0,
        observed_seconds: 86400,
      }
    }
    return {
      date,
      health: 'healthy',
      unhealthy_seconds: 0,
      degraded_seconds: 0,
      unknown_seconds: 0,
      observed_seconds: 86400,
    }
  })
}

function buildSparseDaily(
  totalDays: number,
  observedDays: number,
  {
    unhealthyOffsets = [],
    noSignalOffsets = [],
  }: { unhealthyOffsets?: number[]; noSignalOffsets?: number[] } = {}
): THealthTimelineDay[] {
  return Array.from({ length: totalDays }, (_, i) => {
    const daysAgo = totalDays - 1 - i
    const date = DateTime.now().minus({ days: daysAgo }).toISODate()!
    const isMonitored = daysAgo < observedDays

    if (!isMonitored || noSignalOffsets.includes(daysAgo)) {
      return {
        date,
        health: '',
        unhealthy_seconds: 0,
        degraded_seconds: 0,
        unknown_seconds: 86400,
        observed_seconds: 0,
      }
    }

    const observedSeconds = daysAgo === 0 ? 43200 : 86400

    if (unhealthyOffsets.includes(daysAgo)) {
      return {
        date,
        health: 'unhealthy',
        unhealthy_seconds: 3600,
        degraded_seconds: 0,
        unknown_seconds: 0,
        observed_seconds: observedSeconds,
      }
    }

    return {
      date,
      health: 'healthy',
      unhealthy_seconds: 0,
      degraded_seconds: 0,
      unknown_seconds: 0,
      observed_seconds: observedSeconds,
    }
  })
}

function totalObservedSeconds(daily: THealthTimelineDay[]): number {
  return daily.reduce((total, day) => total + day.observed_seconds, 0)
}

function uptimePercentOf(daily: THealthTimelineDay[]): number {
  const observed = totalObservedSeconds(daily)
  if (!observed) return 0
  const bad = daily.reduce(
    (total, day) => total + day.unhealthy_seconds + day.degraded_seconds,
    0
  )
  return Math.round(((observed - bad) / observed) * 10000) / 100
}

const firstDayDaily = buildSparseDaily(90, 1)
const twoDayDaily = buildSparseDaily(90, 2)
const firstWeekDaily = buildSparseDaily(90, 7, { unhealthyOffsets: [3] })
const gappedDaily = buildSparseDaily(90, 12, {
  unhealthyOffsets: [2],
  noSignalOffsets: [5, 6],
})
const halfWindowDaily = buildSparseDaily(90, 45, { unhealthyOffsets: [8, 30] })
const unmonitoredDaily = buildSparseDaily(90, 0)

const mockComponents: TInstallHealthTimelineComponent[] = [
  {
    install_component_id: 'icmp1',
    component_name: 'api',
    current_health: 'healthy',
    uptime_percent: 99.98,
  },
  {
    install_component_id: 'icmp2',
    component_name: 'worker',
    current_health: 'unhealthy',
    uptime_percent: 96.4,
  },
  {
    install_component_id: 'icmp3',
    component_name: 'database',
    current_health: 'degraded',
    uptime_percent: 99.1,
  },
]

const mockNoSignalComponents: TInstallHealthTimelineComponent[] = [
  {
    install_component_id: 'icmp4',
    component_name: 'networking',
    current_health: 'not-applicable',
    uptime_percent: 0,
  },
  {
    install_component_id: 'icmp5',
    component_name: 'secrets',
    current_health: 'unknown',
    uptime_percent: 0,
  },
  {
    install_component_id: 'icmp6',
    component_name: 'dns',
    current_health: 'not-applicable',
    uptime_percent: 0,
  },
]

const mockTransitions: TInstallComponentHealthTransition[] = [
  {
    from_health: 'unhealthy',
    to_health: 'healthy',
    message: 'Pod api-7d9f8 recovered',
    root_resource_kind: 'Deployment',
    root_resource_namespace: 'default',
    root_resource_name: 'api',
    observed_at: DateTime.now().minus({ hours: 1 }).toISO()!,
  },
  {
    from_health: 'healthy',
    to_health: 'unhealthy',
    message: 'Pod api-7d9f8 is crash looping',
    root_resource_kind: 'Deployment',
    root_resource_namespace: 'default',
    root_resource_name: 'api',
    correlated_deploy_id: 'dep_9f8a3c',
    diagnosis: 'Container was OOMKilled after hitting its memory limit.',
    observed_at: DateTime.now().minus({ hours: 3 }).toISO()!,
  },
  {
    from_health: 'healthy',
    to_health: 'degraded',
    message: 'Elevated latency on readiness probe',
    observed_at: DateTime.now().minus({ days: 2 }).toISO()!,
  },
]

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="p-6 max-w-4xl">
    <Card>{children}</Card>
  </div>
)

export const InstallScope = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={buildDaily(90)}
      uptimePercent={99.42}
      observedSeconds={90 * 86400}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeWithNoSignalComponents = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={buildDaily(90)}
      uptimePercent={99.42}
      observedSeconds={90 * 86400}
      currentHealth="healthy"
      components={[...mockComponents, ...mockNoSignalComponents]}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeAllNoSignalComponents = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={buildDaily(90)}
      uptimePercent={99.42}
      observedSeconds={90 * 86400}
      currentHealth="unknown"
      components={mockNoSignalComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeFirstDayOfHistory = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={firstDayDaily}
      uptimePercent={uptimePercentOf(firstDayDaily)}
      observedSeconds={totalObservedSeconds(firstDayDaily)}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeTwoDaysOfHistory = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={twoDayDaily}
      uptimePercent={uptimePercentOf(twoDayDaily)}
      observedSeconds={totalObservedSeconds(twoDayDaily)}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeFirstWeekOfHistory = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={firstWeekDaily}
      uptimePercent={uptimePercentOf(firstWeekDaily)}
      observedSeconds={totalObservedSeconds(firstWeekDaily)}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeHalfWindowOfHistory = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={halfWindowDaily}
      uptimePercent={uptimePercentOf(halfWindowDaily)}
      observedSeconds={totalObservedSeconds(halfWindowDaily)}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeWithSignalGap = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={gappedDaily}
      uptimePercent={uptimePercentOf(gappedDaily)}
      observedSeconds={totalObservedSeconds(gappedDaily)}
      currentHealth="healthy"
      components={mockComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const InstallScopeNoObservationsYet = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={unmonitoredDaily}
      uptimePercent={0}
      observedSeconds={0}
      currentHealth="unknown"
      components={mockNoSignalComponents}
      componentBasePath="/org123/installs/inst123/components"
    />
  </Frame>
)

export const ComponentScopeFirstDayOfHistory = () => (
  <Frame>
    <HealthTimeline
      scope="component"
      days={90}
      daily={firstDayDaily}
      uptimePercent={uptimePercentOf(firstDayDaily)}
      observedSeconds={totalObservedSeconds(firstDayDaily)}
      currentHealth="healthy"
      transitions={mockTransitions.slice(0, 1)}
      deployBasePath="/org123/installs/inst123/components/comp123/deploys"
    />
  </Frame>
)

export const ComponentScope = () => (
  <Frame>
    <HealthTimeline
      scope="component"
      days={90}
      daily={buildDaily(90)}
      uptimePercent={96.4}
      observedSeconds={90 * 86400}
      currentHealth="unhealthy"
      transitions={mockTransitions}
      deployBasePath="/org123/installs/inst123/components/comp123/deploys"
    />
  </Frame>
)

export const ComponentScopeNoTransitions = () => (
  <Frame>
    <HealthTimeline
      scope="component"
      days={90}
      daily={buildDaily(90)}
      uptimePercent={100}
      observedSeconds={90 * 86400}
      currentHealth="healthy"
      transitions={[]}
      deployBasePath="/org123/installs/inst123/components/comp123/deploys"
    />
  </Frame>
)

export const NoData = () => (
  <Frame>
    <HealthTimeline
      scope="install"
      days={90}
      daily={[]}
      observedSeconds={0}
      currentHealth="unknown"
    />
  </Frame>
)

export const Loading = () => (
  <Frame>
    <HealthTimeline scope="install" days={90} isLoading />
  </Frame>
)
