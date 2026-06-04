type TTitleMap = Record<string, string>

const RUNNER_STATUS_TITLES: TTitleMap = {
  active: 'Runner is healthy',
  error: 'Runner is unhealthy',
  pending: 'Runner is pending',
  provisioning: 'Runner is provisioning',
  deprovisioning: 'Runner is deprovisioning',
  deprovisioned: 'Runner is deprovisioned',
  reprovisioning: 'Runner is reprovisioning',
  offline: 'Runner is offline',
  'awaiting-install-stack-run': 'Runner is awaiting install stack run',
  unknown: 'Runner status is unknown',
}

const SANDBOX_STATUS_TITLES: TTitleMap = {
  active: 'Sandbox is provisioned',
  error: 'Sandbox has an error',
  queued: 'Sandbox is queued',
  provisioning: 'Sandbox is provisioning',
  deprovisioning: 'Sandbox is deprovisioning',
  deprovisioned: 'Sandbox is deprovisioned',
  reprovisioning: 'Sandbox is reprovisioning',
  access_error: 'Sandbox has an access error',
  deleted: 'Sandbox has been deleted',
  delete_failed: 'Sandbox deletion failed',
  empty: 'Sandbox is empty',
  unknown: 'Sandbox status is unknown',
}

const COMPONENTS_STATUS_TITLES: TTitleMap = {
  active: 'Components are deployed',
  inactive: 'Components are inactive',
  error: 'Component has an error',
  noop: 'Deployment had no changes',
  planning: 'Deployment is planning',
  syncing: 'Deployment is syncing',
  executing: 'Deployment is executing',
  cancelled: 'Deployment was cancelled',
  pending: 'Deployment is pending',
  queued: 'Deployment is queued',
  'pending-approval': 'Deployment is pending approval',
  'approval-denied': 'Deployment approval was denied',
  unknown: 'Deployment status is unknown',
}

const DEPROVISIONING_RUNNER_OVERRIDES: TTitleMap = {
  active: 'Runner waiting to teardown',
  deprovisioned: 'Runner teardown complete',
}

const DEPROVISIONED_RUNNER_OVERRIDES: TTitleMap = {
  active: 'Runner torn down',
  deprovisioned: 'Runner teardown complete',
}

const DEPROVISIONING_SANDBOX_OVERRIDES: TTitleMap = {
  active: 'Sandbox waiting to teardown',
  deprovisioned: 'Sandbox teardown complete',
  deprovisioning: 'Sandbox tearing down',
}

const DEPROVISIONED_SANDBOX_OVERRIDES: TTitleMap = {
  active: 'Sandbox torn down',
  deprovisioned: 'Sandbox teardown complete',
  deprovisioning: 'Sandbox torn down',
}

const DEPROVISIONING_COMPONENTS_OVERRIDES: TTitleMap = {
  active: 'Components waiting to teardown',
  pending: 'Components tearing down',
  executing: 'Components tearing down',
}

const DEPROVISIONED_COMPONENTS_OVERRIDES: TTitleMap = {
  active: 'Components torn down',
  pending: 'Components torn down',
  executing: 'Components torn down',
}

function getStatusTitle(
  map: TTitleMap,
  status: string,
  fallback: string
): string {
  return map[status] ?? fallback
}

export function getInstallRunnerStatusTitle(status: string): string {
  return getStatusTitle(
    RUNNER_STATUS_TITLES,
    status,
    RUNNER_STATUS_TITLES.unknown
  )
}

export function getInstallSandboxStatusTitle(status: string): string {
  return getStatusTitle(
    SANDBOX_STATUS_TITLES,
    status,
    SANDBOX_STATUS_TITLES.unknown
  )
}

export function getInstallComponentsStatusTitle(status: string): string {
  return getStatusTitle(
    COMPONENTS_STATUS_TITLES,
    status,
    COMPONENTS_STATUS_TITLES.unknown
  )
}

export function getInstallStatusTitle(
  statusKey: string,
  status: string,
  lifecycleStatus?: string
): string {
  if (lifecycleStatus === 'deprovisioning' || lifecycleStatus === 'deprovisioned') {
    const isFinished = lifecycleStatus === 'deprovisioned'
    let override: string | undefined
    switch (statusKey) {
      case 'runner_status':
        override = (isFinished ? DEPROVISIONED_RUNNER_OVERRIDES : DEPROVISIONING_RUNNER_OVERRIDES)[status]
        break
      case 'sandbox_status':
        override = (isFinished ? DEPROVISIONED_SANDBOX_OVERRIDES : DEPROVISIONING_SANDBOX_OVERRIDES)[status]
        break
      case 'composite_component_status':
        override = (isFinished ? DEPROVISIONED_COMPONENTS_OVERRIDES : DEPROVISIONING_COMPONENTS_OVERRIDES)[status]
        break
    }
    if (override) return override
  }

  switch (statusKey) {
    case 'runner_status':
      return getInstallRunnerStatusTitle(status)
    case 'sandbox_status':
      return getInstallSandboxStatusTitle(status)
    case 'composite_component_status':
      return getInstallComponentsStatusTitle(status)
    default:
      return 'Waiting on status'
  }
}
