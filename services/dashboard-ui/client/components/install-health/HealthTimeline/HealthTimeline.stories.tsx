export default {
  title: 'InstallHealth/HealthTimeline',
}

import { DateTime } from 'luxon'
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

export const InstallScope = () => (
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
)

export const InstallScopeWithNoSignalComponents = () => (
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
)

export const InstallScopeAllNoSignalComponents = () => (
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
)

export const ComponentScope = () => (
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
)

export const ComponentScopeNoTransitions = () => (
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
)

export const NoData = () => (
  <HealthTimeline
    scope="install"
    days={90}
    daily={[]}
    observedSeconds={0}
    currentHealth="unknown"
  />
)

export const Loading = () => <HealthTimeline scope="install" days={90} isLoading />
