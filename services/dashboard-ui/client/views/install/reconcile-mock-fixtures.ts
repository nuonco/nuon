// Mockup fixtures for the install out-of-sync reconcile feature (ladle stories only).
// See reconcile-mock-types.ts for the proposed API shape these fixtures model.
import type { TInstallOverviewCommit } from '@/types'
import type {
  TMockDeployPlanItem,
  TMockInstallReconciliation,
} from './reconcile-mock-types'

const messages = [
  'Raise the checkout worker retry budget',
  'Pin the ledger module to 2.4.0',
  'Add a webhook signing secret input',
  'Move scheduler jobs to the shared queue',
]

export const outOfSyncCommit = (
  index: number,
  overrides: Partial<TInstallOverviewCommit> = {}
): TInstallOverviewCommit => ({
  app_config_id: `cfg-${index + 40}`,
  branch_id: 'br-1',
  workflow_id: `wf-branch-${index}`,
  run_id: `run-${index}`,
  sha: `${index}bc123def4567890abc1`,
  message: messages[index % messages.length],
  author: 'Example Developer',
  created_at: new Date(Date.now() - (index + 1) * 3_600_000).toISOString(),
  run_status: 'success',
  ...overrides,
})

export const failedUpdateDeployPlan = (): TMockDeployPlanItem[] => [
  {
    kind: 'image',
    id: 'ic-api-image',
    name: 'api (image)',
    order: 1,
    reason: 'older_build_deployed',
    reason_message: 'A newer build was never synced after the failed update.',
  },
  {
    kind: 'component',
    id: 'ic-api',
    name: 'api',
    order: 2,
    reason: 'deploy_failed',
    reason_message: 'This component failed to deploy during the last update.',
  },
  {
    kind: 'component',
    id: 'ic-worker',
    name: 'worker',
    order: 3,
    reason: 'not_attempted',
    reason_message:
      'The update workflow stopped before this component was deployed.',
  },
]

export const manualChangeDeployPlan = (): TMockDeployPlanItem[] => [
  {
    kind: 'component',
    id: 'ic-ledger',
    name: 'ledger',
    order: 1,
    reason: 'older_build_deployed',
    reason_message: 'A user manually redeployed this component with build v38.',
  },
  {
    kind: 'component',
    id: 'ic-scheduler',
    name: 'scheduler',
    order: 2,
    reason: 'skipped',
    reason_message: 'A user skipped this step during a manual deploy.',
  },
]

export const behindAndOutOfSyncDeployPlan = (): TMockDeployPlanItem[] => [
  {
    kind: 'stack',
    id: 'stk-payments',
    name: 'payments stack',
    order: 1,
    reason: 'config_changed',
    reason_message: 'The app branch moved on since the failed update.',
  },
  {
    kind: 'image',
    id: 'ic-api-image',
    name: 'api (image)',
    order: 2,
    reason: 'older_build_deployed',
    reason_message: 'A newer build was never synced after the failed update.',
  },
  {
    kind: 'component',
    id: 'ic-api',
    name: 'api',
    order: 3,
    reason: 'deploy_failed',
    reason_message: 'This component failed to deploy during the last update.',
  },
  {
    kind: 'sandbox',
    id: 'sbx-1',
    name: 'sandbox',
    order: 4,
    reason: 'dependency_changed',
    reason_message: 'Depends on the api component, which hasn\u2019t deployed yet.',
  },
]

const outOfSyncReconciliationBase: TMockInstallReconciliation = {
  status: 'out_of_sync',
  app_branch_id: 'br-1',
  app_branch_name: 'main',
  applied: { app_config_id: 'cfg-42', version: 42, commit: outOfSyncCommit(2) },
  target: { app_config_id: 'cfg-42', version: 42, commit: outOfSyncCommit(2) },
  skipped_runs: [],
  skipped_run_count: 0,
  blockers: [],
  out_of_sync: true,
}

export const failedUpdateReconciliation = (): TMockInstallReconciliation => ({
  ...outOfSyncReconciliationBase,
  failed_workflow_id: 'wf-update-failed-1',
  deploy_plan: failedUpdateDeployPlan(),
})

export const manualChangeReconciliation = (): TMockInstallReconciliation => ({
  ...outOfSyncReconciliationBase,
  deploy_plan: manualChangeDeployPlan(),
})

export const behindAndOutOfSyncReconciliation = (): TMockInstallReconciliation => ({
  ...outOfSyncReconciliationBase,
  status: 'behind',
  out_of_sync: true,
  failed_workflow_id: 'wf-update-failed-1',
  applied: { app_config_id: 'cfg-40', version: 40, commit: outOfSyncCommit(3) },
  target: { app_config_id: 'cfg-43', version: 43, commit: outOfSyncCommit(0) },
  skipped_runs: [outOfSyncCommit(0), outOfSyncCommit(1)],
  skipped_run_count: 2,
  deploy_plan: behindAndOutOfSyncDeployPlan(),
})
