import type { TRunnerJob } from '@/types'

export type TJobStatus =
  | 'finished'
  | 'failed'
  | 'timed-out'
  | 'queued'
  | 'in-progress'
  | 'not-attempted'
  | 'available'
  | 'cancelled'

export type TJobGroup =
  | 'build'
  | 'sandbox'
  | 'sync'
  | 'deploy'
  | 'actions'
  | 'operations'

export function getJobHref(job: TRunnerJob): string {
  const { group, metadata } = job ?? {}
  switch (group) {
    case 'build':
      return `apps/${metadata?.app_id}/components/${metadata?.component_id}/builds/${metadata?.component_build_id}`
    case 'sandbox':
      return `installs/${metadata?.install_id}/sandbox/${metadata?.sandbox_run_id}`
    case 'sync':
    case 'deploy':
      return `installs/${metadata?.install_id}/components/${metadata?.component_id}/deploys/${metadata?.deploy_id}`
    case 'actions':
      return `installs/${metadata?.install_id}/actions/${metadata?.action_workflow_id}/${metadata?.action_workflow_run_id}`
    default:
      return ''
  }
}

export function getJobName(job: TRunnerJob): string {
  const { group, metadata, type } = job ?? {}
  switch (group) {
    case 'build':
    case 'sync':
    case 'deploy':
      return metadata?.component_name ?? 'Unknown'
    case 'sandbox':
      return metadata?.sandbox_run_type ?? 'Unknown'
    case 'actions':
      return metadata?.action_workflow_name ?? 'Unknown'
    case 'operations':
      return type ?? 'Unknown'
    default:
      return 'Unknown'
  }
}

export function getJobExecutionStatus(job: TRunnerJob): string {
  const statusHandlers: Record<TJobGroup, (j: TRunnerJob) => string> = {
    build: getBuildJobExecutionStatus,
    sandbox: getSandboxJobExecutionStatus,
    sync: getSyncJobExecutionStatus,
    deploy: getDeployJobExecutionStatus,
    actions: getActionsJobExecutionStatus,
    operations: getOperationsJobExecutionStatus,
  }
  return statusHandlers[job.group]?.(job) ?? 'Unknown'
}

const statusMessagesByGroup: Record<TJobGroup, Record<TJobStatus, string>> = {
  build: {
    finished: 'component built successfully',
    failed: 'component build failed',
    'timed-out': 'component build timed out',
    queued: 'component build queued',
    'in-progress': 'component build is being built',
    'not-attempted': 'component build not attempted',
    available: 'component build starting soon',
    cancelled: 'component build canceled',
  },
  sandbox: {
    finished: 'sandbox provisioned successfully',
    failed: 'sandbox provisioning failed',
    'timed-out': 'sandbox provisioning timed out',
    queued: 'sandbox provisioning queued',
    'in-progress': 'sandbox is being provisioned',
    'not-attempted': 'sandbox provisioning not attempted',
    available: 'sandbox provisioning starting soon',
    cancelled: 'sandbox provisioning canceled',
  },
  sync: {
    finished: 'component synced successfully',
    failed: 'component sync failed',
    'timed-out': 'component sync timed out',
    queued: 'component sync queued',
    'in-progress': 'component is syncing',
    'not-attempted': 'component sync not attempted',
    available: 'component sync starting soon',
    cancelled: 'component sync canceled',
  },
  deploy: {
    finished: 'component deployed successfully',
    failed: 'component deployment failed',
    'timed-out': 'component deployment timed out',
    queued: 'component deployment queued',
    'in-progress': 'component is being deployed',
    'not-attempted': 'component deployment not attempted',
    available: 'component deployment starting soon',
    cancelled: 'component deployment canceled',
  },
  actions: {
    finished: 'action completed successfully',
    failed: 'action failed',
    'timed-out': 'action timed out',
    queued: 'action queued',
    'in-progress': 'action is running',
    'not-attempted': 'action not attempted',
    available: 'action starting soon',
    cancelled: 'action canceled',
  },
  operations: {
    finished: 'operation completed successfully',
    failed: 'operation failed',
    'timed-out': 'operation timed out',
    queued: 'operation queued',
    'in-progress': 'operation is running',
    'not-attempted': 'operation not attempted',
    available: 'operation starting soon',
    cancelled: 'operation canceled',
  },
}

function getBuildJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.build[job.status as TJobStatus] ?? 'Unknown'
}
function getSandboxJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.sandbox[job.status as TJobStatus] ?? 'Unknown'
}
function getSyncJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.sync[job.status as TJobStatus] ?? 'Unknown'
}
function getDeployJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.deploy[job.status as TJobStatus] ?? 'Unknown'
}
function getActionsJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.actions[job.status as TJobStatus] ?? 'Unknown'
}
function getOperationsJobExecutionStatus(job: TRunnerJob) {
  return statusMessagesByGroup.operations[job.status as TJobStatus] ?? 'Unknown'
}
