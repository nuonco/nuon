export default {
  title: 'Installs/InstallHealth',
}

import { DateTime } from 'luxon'
import { HealthTimelineComponent } from '@/components/install-health/HealthTimeline'
import {
  groupComponentResources,
  groupSandboxResources,
  InstallResourcesTableComponent,
} from '@/components/install-resources/InstallResourcesTable'
import { healthFacetCounts } from '@/components/install-resources/InstallResourcesTable/InstallResourcesTable'
import type {
  THealthTimelineDay,
  TInstallHealthTimelineComponent,
  TInstallResource,
} from '@/types'
import { InstallHealth } from './InstallHealth'

const components: TInstallHealthTimelineComponent[] = [
  {
    install_component_id: 'instcmp-api',
    component_id: 'cmp-api',
    component_name: 'api',
    current_health: 'healthy',
    uptime_percent: 99.98,
    observed_seconds: 30 * 86400,
  },
  {
    install_component_id: 'instcmp-worker',
    component_id: 'cmp-worker',
    component_name: 'worker',
    current_health: 'healthy',
    uptime_percent: 99.91,
    observed_seconds: 30 * 86400,
  },
]

const resources: TInstallResource[] = [
  {
    install_component_id: 'instcmp-api',
    component_id: 'cmp-api',
    source: 'component',
    kind: 'Deployment',
    namespace: 'payments',
    name: 'api',
    health: 'healthy',
    provider: 'kubernetes',
    observed_at: DateTime.now().minus({ minutes: 1 }).toISO()!,
  },
  {
    install_component_id: 'instcmp-api',
    component_id: 'cmp-api',
    source: 'component',
    kind: 'Service',
    namespace: 'payments',
    name: 'api',
    health: 'healthy',
    provider: 'kubernetes',
    observed_at: DateTime.now().minus({ minutes: 1 }).toISO()!,
  },
  {
    install_component_id: 'instcmp-worker',
    component_id: 'cmp-worker',
    source: 'component',
    kind: 'Deployment',
    namespace: 'payments',
    name: 'worker',
    health: 'healthy',
    provider: 'kubernetes',
    observed_at: DateTime.now().minus({ minutes: 1 }).toISO()!,
  },
  {
    source: 'sandbox',
    owner_name: 'cert-manager',
    kind: 'Deployment',
    namespace: 'cert-manager',
    name: 'cert-manager',
    health: 'healthy',
    provider: 'kubernetes',
    observed_at: DateTime.now().minus({ minutes: 1 }).toISO()!,
  },
]

const unhealthyResources: TInstallResource[] = resources.map((resource) =>
  resource.name === 'api' && resource.kind === 'Deployment'
    ? {
        ...resource,
        health: 'unhealthy',
        message: 'Available replicas are below the desired count.',
      }
    : resource.name === 'worker'
      ? {
          ...resource,
          health: 'degraded',
          message: 'One pod is not ready.',
        }
      : resource
)

const daily = (
  currentHealth: string,
  incidentOffsets: number[] = []
): THealthTimelineDay[] =>
  Array.from({ length: 30 }, (_, index) => {
    const daysAgo = 29 - index
    const incident = incidentOffsets.includes(daysAgo)
    const today = daysAgo === 0

    return {
      date: DateTime.now().minus({ days: daysAgo }).toISODate()!,
      health: incident ? currentHealth : 'healthy',
      unhealthy_seconds: incident && currentHealth === 'unhealthy' ? 7200 : 0,
      degraded_seconds: incident && currentHealth === 'degraded' ? 3600 : 0,
      unknown_seconds: 0,
      observed_seconds: today ? 43200 : 86400,
    }
  })

const noObservations = (): THealthTimelineDay[] =>
  Array.from({ length: 30 }, (_, index) => ({
    date: DateTime.now().minus({ days: 29 - index }).toISODate()!,
    health: 'unknown',
    unhealthy_seconds: 0,
    degraded_seconds: 0,
    unknown_seconds: 86400,
    observed_seconds: 0,
  }))

const componentNames = {
  'instcmp-api': 'api',
  'instcmp-worker': 'worker',
}

const ResourceTable = ({
  isLoading = false,
  values = resources,
}: {
  isLoading?: boolean
  values?: TInstallResource[]
}) => (
  <InstallResourcesTableComponent
    componentGroups={groupComponentResources(values, componentNames)}
    sandboxGroups={groupSandboxResources(values)}
    healthCounts={healthFacetCounts(values)}
    isLoading={isLoading}
    kind=""
    namespace=""
    health=""
    search=""
    kindOptions={['Deployment', 'Service']}
    namespaceOptions={['cert-manager', 'payments']}
    onKindChange={() => {}}
    onNamespaceChange={() => {}}
    onHealthChange={() => {}}
  />
)

const Timeline = ({
  clusterAccessError,
  currentHealth = 'healthy',
  timeline = daily('healthy'),
  timelineComponents = components,
}: {
  clusterAccessError?: string
  currentHealth?: string
  timeline?: THealthTimelineDay[]
  timelineComponents?: TInstallHealthTimelineComponent[]
}) => (
  <HealthTimelineComponent
    scope="install"
    days={30}
    daily={timeline}
    uptimePercent={
      currentHealth === 'unhealthy'
        ? 97.42
        : currentHealth === 'degraded'
          ? 99.61
          : timeline.some((day) => day.observed_seconds > 0)
            ? 100
            : 0
    }
    observedSeconds={timeline.reduce(
      (total, day) => total + day.observed_seconds,
      0
    )}
    currentHealth={currentHealth}
    components={timelineComponents}
    componentBasePath="/org-1/installs/inst-1/resources/components"
    getComponentHref={(componentId) =>
      `/org-1/installs/inst-1/resources/components#${componentId}`
    }
    clusterAccessError={clusterAccessError}
  />
)

export const Healthy = () => (
  <InstallHealth
    timeline={<Timeline />}
    resources={<ResourceTable />}
  />
)

export const Degraded = () => (
  <InstallHealth
    timeline={
      <Timeline
        currentHealth="degraded"
        timeline={daily('degraded', [0, 8])}
        timelineComponents={[
          components[0],
          {
            ...components[1],
            current_health: 'degraded',
            uptime_percent: 98.72,
          },
        ]}
      />
    }
    resources={<ResourceTable values={unhealthyResources} />}
  />
)

export const Unhealthy = () => (
  <InstallHealth
    timeline={
      <Timeline
        currentHealth="unhealthy"
        timeline={daily('unhealthy', [0, 1, 14])}
        timelineComponents={[
          {
            ...components[0],
            current_health: 'unhealthy',
            uptime_percent: 94.2,
          },
          {
            ...components[1],
            current_health: 'degraded',
            uptime_percent: 98.72,
          },
        ]}
      />
    }
    resources={<ResourceTable values={unhealthyResources} />}
  />
)

export const ClusterAccessError = () => (
  <InstallHealth
    timeline={
      <Timeline
        currentHealth="unknown"
        timeline={noObservations()}
        timelineComponents={components.map((component) => ({
          ...component,
          current_health: 'unknown',
          uptime_percent: 0,
          observed_seconds: 0,
        }))}
        clusterAccessError="The runner could not authenticate to the cluster."
      />
    }
    resources={<ResourceTable values={[]} />}
  />
)

export const NoObservations = () => (
  <InstallHealth
    timeline={
      <Timeline
        currentHealth="unknown"
        timeline={noObservations()}
        timelineComponents={components.map((component) => ({
          ...component,
          current_health: 'unknown',
          uptime_percent: 0,
          observed_seconds: 0,
        }))}
      />
    }
    resources={<ResourceTable values={[]} />}
  />
)

export const EmptyResources = () => (
  <InstallHealth
    timeline={<Timeline />}
    resources={<ResourceTable values={[]} />}
  />
)

export const Loading = () => (
  <InstallHealth
    timeline={<HealthTimelineComponent scope="install" days={30} isLoading />}
    resources={<ResourceTable isLoading values={[]} />}
  />
)
