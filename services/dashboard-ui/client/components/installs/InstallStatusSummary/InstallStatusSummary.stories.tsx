export default {
  title: 'Features / Installs / Install status summary',
}

import { Card } from '@/components/common/Card'
import type { TInstallStatus, TInstallStatusAxis } from '@/types'
import { InstallStatusSummary } from './InstallStatusSummary'

const axis = (
  status: string,
  description: string,
  extra?: TInstallStatusAxis['metadata']
): TInstallStatusAxis => ({
  status,
  status_human_description: description,
  metadata: extra,
})

const deployed = axis('active', 'All deployed', {
  counts: { deployed: 4, failed: 0, progressing: 0, not_deployed: 0 },
})
const healthyResources = axis('healthy', 'All healthy', {
  counts: {
    healthy: 12,
    progressing: 0,
    degraded: 0,
    unhealthy: 0,
    unknown: 0,
  },
})
const healthyChecks = axis('healthy', 'All healthy', {
  counts: { healthy: 4, progressing: 0, degraded: 0, unhealthy: 0, unknown: 0 },
})

const Frame = ({ status }: { status?: TInstallStatus }) => (
  <Card
    className="!p-4 !shadow-none max-w-xl"
    aria-label="Install status summary"
  >
    <InstallStatusSummary status={status} />
  </Card>
)

export const AllHealthy = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: healthyChecks,
    }}
  />
)

export const DeploymentsInProgress = () => (
  <Frame
    status={{
      deployments: axis('pending', 'In progress', {
        counts: { deployed: 2, failed: 0, progressing: 1, not_deployed: 0 },
      }),
      resources: healthyResources,
      health_checks: healthyChecks,
    }}
  />
)

export const DeploymentsPartial = () => (
  <Frame
    status={{
      deployments: axis('pending', '8 of 13 deployed', {
        counts: { deployed: 8, failed: 0, progressing: 0, not_deployed: 5 },
      }),
      resources: healthyResources,
      health_checks: healthyChecks,
    }}
  />
)

export const DeploymentsNotDeployed = () => (
  <Frame
    status={{
      deployments: axis('pending', 'Not deployed', {
        counts: { deployed: 0, failed: 0, progressing: 0, not_deployed: 4 },
      }),
      resources: axis('unknown', 'No resources'),
      health_checks: axis('unknown', 'No health checks'),
    }}
  />
)

export const DeploymentsNone = () => (
  <Frame
    status={{
      deployments: axis('pending', 'No deployments', {
        counts: { deployed: 0, failed: 0, progressing: 0, not_deployed: 0 },
      }),
      resources: axis('unknown', 'No resources'),
      health_checks: axis('unknown', 'No health checks'),
    }}
  />
)

export const DeploymentFailed = () => (
  <Frame
    status={{
      deployments: axis('error', '1 deployment failed', {
        counts: { deployed: 3, failed: 1, progressing: 0, not_deployed: 0 },
      }),
      resources: healthyResources,
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesDegraded = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('degraded', 'Degraded', {
        counts: {
          healthy: 10,
          progressing: 0,
          degraded: 2,
          unhealthy: 0,
          unknown: 0,
        },
      }),
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesUnhealthy = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('unhealthy', 'Unhealthy', {
        counts: {
          healthy: 8,
          progressing: 0,
          degraded: 0,
          unhealthy: 1,
          unknown: 0,
        },
      }),
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesProgressing = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('progressing', 'Progressing', {
        counts: {
          healthy: 9,
          progressing: 3,
          degraded: 0,
          unhealthy: 0,
          unknown: 0,
        },
      }),
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesUnknown = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('unknown', 'Unknown', {
        counts: {
          healthy: 0,
          progressing: 0,
          degraded: 0,
          unhealthy: 0,
          unknown: 4,
        },
      }),
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesNone = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('unknown', 'No resources'),
      health_checks: healthyChecks,
    }}
  />
)

export const ResourcesClusterUnavailable = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: axis('unknown', 'Cluster unavailable', {
        cluster_access_error: 'runner has no cluster credentials',
      }),
      health_checks: healthyChecks,
    }}
  />
)

export const HealthChecksDegraded = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: axis('degraded', 'Degraded', {
        counts: {
          healthy: 3,
          progressing: 0,
          degraded: 1,
          unhealthy: 0,
          unknown: 0,
        },
      }),
    }}
  />
)

export const HealthChecksUnhealthy = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: axis('unhealthy', 'Unhealthy', {
        counts: {
          healthy: 2,
          progressing: 0,
          degraded: 0,
          unhealthy: 1,
          unknown: 0,
        },
      }),
    }}
  />
)

export const HealthChecksProgressing = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: axis('progressing', 'Progressing', {
        counts: {
          healthy: 2,
          progressing: 1,
          degraded: 0,
          unhealthy: 0,
          unknown: 0,
        },
      }),
    }}
  />
)

export const HealthChecksUnknown = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: axis('unknown', 'Unknown', {
        counts: {
          healthy: 1,
          progressing: 0,
          degraded: 0,
          unhealthy: 0,
          unknown: 2,
        },
      }),
    }}
  />
)

export const HealthChecksNone = () => (
  <Frame
    status={{
      deployments: deployed,
      resources: healthyResources,
      health_checks: axis('unknown', 'No health checks'),
    }}
  />
)

export const Loading = () => (
  <Card className="!p-4 !shadow-none max-w-xl">
    <InstallStatusSummary loading />
  </Card>
)

export const Unavailable = () => (
  <Card className="!p-4 !shadow-none max-w-xl">
    <InstallStatusSummary unavailable />
  </Card>
)
