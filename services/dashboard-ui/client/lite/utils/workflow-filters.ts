import { DateTime } from 'luxon'
import type { TWorkflowType } from '@/types/ctl-api.types'

export type TWorkflowStatusOption =
  | 'running'
  | 'queued'
  | 'succeeded'
  | 'failed'
  | 'cancelled'
  | 'warning'
  | 'skipped'

export type TWorkflowTypeGroup =
  | 'deploy'
  | 'teardown'
  | 'provision'
  | 'deprovision'
  | 'drift'
  | 'action'
  | 'runbook'
  | 'inputs'
  | 'secrets'
  | 'component-toggle'
  | 'recovery'
  | 'branch-manual'
  | 'branch-config-repo'
  | 'branch-component-repo'
  | 'branch-config'
  | 'install-sync'
  | 'config-build'

export type TWorkflowOwner = 'install' | 'app'

export type TWorkflowDatePreset = '24h' | '7d' | '30d'

export type TWorkflowPreviewOption = 'preview' | 'rollout'

export const WORKFLOW_PREVIEW_LABELS: Record<TWorkflowPreviewOption, string> = {
  preview: 'Preview',
  rollout: 'Rollout',
}

export const previewFilterParameter = (
  selected: Iterable<TWorkflowPreviewOption>
): boolean | undefined => {
  const options = new Set(selected)
  if (options.size !== 1) return undefined
  return options.has('preview')
}

export const WORKFLOW_STATUS_GROUPS: Record<
  TWorkflowStatusOption,
  readonly string[]
> = {
  running: [
    'in-progress',
    'planning',
    'applying',
    'checking-plan',
    'retrying',
    'failed-pending-retry',
  ],
  queued: ['pending', 'queued'],
  succeeded: ['success'],
  failed: ['error', 'failed-pending-retry'],
  cancelled: ['cancelled'],
  warning: ['warning'],
  skipped: ['not-attempted', 'user-skipped', 'auto-skipped', 'discarded'],
}

export const WORKFLOW_STATUS_LABELS: Record<TWorkflowStatusOption, string> = {
  running: 'Running',
  queued: 'Queued',
  succeeded: 'Succeeded',
  failed: 'Failed',
  cancelled: 'Cancelled',
  warning: 'Warning',
  skipped: 'Skipped',
}

export const WORKFLOW_TYPE_GROUPS: Record<TWorkflowType, TWorkflowTypeGroup> = {
  manual_deploy: 'deploy',
  deploy_components: 'deploy',
  teardown_component: 'teardown',
  teardown_components: 'teardown',
  provision: 'provision',
  reprovision: 'provision',
  reprovision_stack: 'provision',
  reprovision_sandbox: 'provision',
  deprovision: 'deprovision',
  deprovision_sandbox: 'deprovision',
  drift_run: 'drift',
  drift_run_reprovision_sandbox: 'drift',
  action_workflow_run: 'action',
  runbook_run: 'runbook',
  input_update: 'inputs',
  sync_secrets: 'secrets',
  component_enabled: 'component-toggle',
  component_disabled: 'component-toggle',
  recover_helm_release: 'recovery',
  app_branches_manual_update: 'branch-manual',
  app_branches_config_repo_update: 'branch-config-repo',
  app_branches_component_repo_update: 'branch-component-repo',
  app_branch_config_update: 'branch-config',
  app_install_sync: 'install-sync',
  app_config_build: 'config-build',
}

export const WORKFLOW_TYPE_LABELS: Record<TWorkflowTypeGroup, string> = {
  deploy: 'Deploy',
  teardown: 'Teardown',
  provision: 'Provision',
  deprovision: 'Deprovision',
  drift: 'Drift scan',
  action: 'Action',
  runbook: 'Runbook',
  inputs: 'Inputs',
  secrets: 'Secrets',
  'component-toggle': 'Component toggle',
  recovery: 'Recovery',
  'branch-manual': 'Manual run',
  'branch-config-repo': 'Config repo push',
  'branch-component-repo': 'Component repo push',
  'branch-config': 'Config change',
  'install-sync': 'Install sync',
  'config-build': 'Config build',
}

const INSTALL_TYPE_GROUPS: readonly TWorkflowTypeGroup[] = [
  'deploy',
  'teardown',
  'provision',
  'deprovision',
  'drift',
  'action',
  'runbook',
  'inputs',
  'secrets',
  'component-toggle',
  'recovery',
]

const APP_TYPE_GROUPS: readonly TWorkflowTypeGroup[] = [
  'branch-manual',
  'branch-config-repo',
  'branch-component-repo',
  'branch-config',
  'install-sync',
  'config-build',
]

export const WORKFLOW_DATE_LABELS: Record<TWorkflowDatePreset, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
}

const DATE_PRESET_HOURS: Record<TWorkflowDatePreset, number> = {
  '24h': 24,
  '7d': 24 * 7,
  '30d': 24 * 30,
}

export const workflowStatusOptions = (): readonly TWorkflowStatusOption[] =>
  Object.keys(WORKFLOW_STATUS_GROUPS) as TWorkflowStatusOption[]

export const workflowTypeOptions = (
  owner: TWorkflowOwner
): readonly TWorkflowTypeGroup[] =>
  owner === 'app' ? APP_TYPE_GROUPS : INSTALL_TYPE_GROUPS

export const workflowTypesForGroups = (
  groups: Iterable<TWorkflowTypeGroup>
): readonly TWorkflowType[] => {
  const wanted = new Set(groups)
  return (Object.keys(WORKFLOW_TYPE_GROUPS) as TWorkflowType[]).filter(
    (type) => wanted.has(WORKFLOW_TYPE_GROUPS[type])
  )
}

export const statusFilterParameter = (
  selected: Iterable<TWorkflowStatusOption>
): string | undefined => {
  const statuses = new Set(
    [...selected].flatMap((option) => WORKFLOW_STATUS_GROUPS[option] ?? [])
  )
  return statuses.size ? [...statuses].join(',') : undefined
}

export const typeFilterParameter = (
  selected: Iterable<TWorkflowTypeGroup>
): string | undefined => {
  const types = workflowTypesForGroups(selected)
  return types.length ? types.join(',') : undefined
}

export const createdAtGteForPreset = (
  preset: TWorkflowDatePreset | undefined,
  now: DateTime = DateTime.now()
): string | undefined =>
  preset
    ? now.minus({ hours: DATE_PRESET_HOURS[preset] }).toUTC().toISO() ??
      undefined
    : undefined

export const datePresetFilterParameter = (
  selected: Iterable<TWorkflowDatePreset>,
  now: DateTime = DateTime.now()
): string | undefined => {
  const widest = [...selected].sort(
    (left, right) => DATE_PRESET_HOURS[right] - DATE_PRESET_HOURS[left]
  )[0]
  return createdAtGteForPreset(widest, now)
}
