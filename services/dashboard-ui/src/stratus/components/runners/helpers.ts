import type { TRunnerJob } from '@/types'

export function getJobHref(job: TRunnerJob): string {
  let hrefPath: string

  switch (job?.group) {
    case 'build':
      hrefPath = `apps/${job?.metadata?.app_id}/components/${job?.metadata?.component_id}/builds/${job?.metadata?.component_build_id}`
      break
    case 'sandbox':
      hrefPath = `installs/${job?.metadata?.install_id}/sandbox/${job?.metadata?.sandbox_run_id}`
      break
    case 'sync':
      hrefPath = `installs/${job?.metadata?.install_id}/components/${job?.metadata?.component_id}/deploys/${job?.metadata?.deploy_id}`
      break
    case 'deploy':
      hrefPath = `installs/${job?.metadata?.install_id}/components/${job?.metadata?.component_id}/deploys/${job?.metadata?.deploy_id}`
      break
    case 'actions':
      hrefPath = `installs/${job?.metadata?.install_id}/actions/${job?.metadata?.action_workflow_id}/${job?.metadata?.action_workflow_run_id}`
      break
    default:
      hrefPath = ''
  }

  return hrefPath
}

export function getJobName(job: TRunnerJob): string {
  let name: string

  switch (job?.group) {
    case 'build':
      name = job?.metadata?.component_name
      break
    case 'sandbox':
      name = job?.metadata?.sandbox_run_type
      break
    case 'sync':
      name = job?.metadata?.component_name
      break
    case 'deploy':
      name = job?.metadata?.component_name
      break
    case 'actions':
      name = job?.metadata?.action_workflow_name
      break
    case 'operations':
      name = job?.type
      break
    default:
      name = 'Unknown'
  }

  return name
}

type TJobStatus =
  | 'finished'
  | 'failed'
  | 'timed-out'
  | 'queued'
  | 'in-progress'
  | 'not-attempted'
  | 'available'
  | 'cancelled'

type TJobGroup =
  | 'build'
  | 'sandbox'
  | 'sync'
  | 'deploy'
  | 'actions'
  | 'operations'

export function getJobExecutionStatus(job: TRunnerJob): string {
  const statusHandlers: Record<TJobGroup, (job: TRunnerJob) => string> = {
    build: getBuildJobExecutionStatus,
    sandbox: getSandboxJobExecutionStatus,
    sync: getSyncJobExecutionStatus,
    deploy: getDeployJobExecutionStatus,
    actions: getActionsJobExecutionStatus,
    operations: getOperationsJobExecutionStatus,
  }

  return statusHandlers[job.group]?.(job) ?? 'Unknown'
}

function getBuildJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'component built successfully',
    failed: 'component build failed',
    'timed-out': 'component build timed out',
    queued: 'component build queued',
    'in-progress': 'component build is being built',
    'not-attempted': 'component build not attempted',
    available: 'component build starting soon',
    cancelled: 'component build canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

function getSandboxJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'sandbox provisioned successfully',
    failed: 'sandbox provisioning failed',
    'timed-out': 'sandbox provisioning timed out',
    queued: 'sandbox provisioning queued',
    'in-progress': 'sandbox is being provisioned',
    'not-attempted': 'sandbox provisioning not attempted',
    available: 'sandbox provisioning starting soon',
    cancelled: 'sandbox provisioning canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

function getSyncJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'component synced successfully',
    failed: 'component sync failed',
    'timed-out': 'component sync timed out',
    queued: 'component sync queued',
    'in-progress': 'component is syncing',
    'not-attempted': 'component sync not attempted',
    available: 'component sync starting soon',
    cancelled: 'component sync canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

function getDeployJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'component deployed successfully',
    failed: 'component deployment failed',
    'timed-out': 'component deployment timed out',
    queued: 'component deployment queued',
    'in-progress': 'component is being deployed',
    'not-attempted': 'component deployment not attempted',
    available: 'component deployment starting soon',
    cancelled: 'component deployment canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

function getActionsJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'action completed successfully',
    failed: 'action failed',
    'timed-out': 'action timed out',
    queued: 'action queued',
    'in-progress': 'action is running',
    'not-attempted': 'action not attempted',
    available: 'action starting soon',
    cancelled: 'action canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

function getOperationsJobExecutionStatus(job: TRunnerJob) {
  const statusMessages: Record<TJobStatus, string> = {
    finished: 'operation completed successfully',
    failed: 'operation failed',
    'timed-out': 'operation timed out',
    queued: 'operation queued',
    'in-progress': 'operation is running',
    'not-attempted': 'operation not attempted',
    available: 'operation starting soon',
    cancelled: 'operation canceled',
  }

  return statusMessages[job.status] ?? 'Unknown'
}

export const RECENT_ACTIVITY_SEARCH_PARAM = 'recent-activity'
export const RECENT_ACTIVITY_LIMIT = 10
export const RECENT_ACTIVITY_GROUPS = [
  'actions',
  'build',
  'deploy',
  'operations',
  'sandbox',
  'sync',
]
