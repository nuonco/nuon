export default {
  title: 'Features / Installs / Management / Reconcile install',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import type { TInstallReconciliation } from '@/types'
import {
  behindAndOutOfSyncReconciliation,
  failedUpdateReconciliation,
  manualChangeReconciliation,
} from '@/views/install/reconcile-mock-fixtures'
import type { TMockInstallReconciliation } from '@/views/install/reconcile-mock-types'
import { ReconcileInstallModal } from './ReconcileInstall'

const noop = () => {}

const commit = (index: number) => ({
  app_config_id: `cfg-${index}`,
  branch_id: 'br-1',
  workflow_id: `wf-${index}`,
  run_id: `run-${index}`,
  sha: `${index}bc123def4567890`,
  message: `Update acme-api settings ${index}`,
  author: 'Example Developer',
  created_at: new Date(Date.now() - index * 3_600_000).toISOString(),
  run_status: 'success',
})

const behind: TInstallReconciliation = {
  status: 'behind',
  app_branch_id: 'br-1',
  app_branch_name: 'main',
  applied: { app_config_id: 'cfg-1', version: 3, commit: commit(5) },
  target: { app_config_id: 'cfg-4', version: 6, commit: commit(1) },
  skipped_runs: [commit(1), commit(2), commit(3)],
  skipped_run_count: 3,
  blockers: [],
}

const getRunHref = () => undefined

const Story = (props: Partial<Parameters<typeof ReconcileInstallModal>[0]>) => (
  <ModalStory>
    <ReconcileInstallModal
      installName="acme-prod"
      reconciliation={behind}
      isLoading={false}
      isPending={false}
      repo="acme/payments"
      getRunHref={getRunHref}
      onSubmit={noop}
      {...props}
    />
  </ModalStory>
)

const DryRunStory = ({
  reconciliation,
}: {
  reconciliation: TMockInstallReconciliation
}) => {
  const [planOnly, setPlanOnly] = useState(false)
  return (
    <ModalStory>
      <ReconcileInstallModal
        installName="acme-prod"
        reconciliation={reconciliation}
        isLoading={false}
        isPending={false}
        repo="acme/payments"
        getRunHref={getRunHref}
        planOnly={planOnly}
        onPlanOnlyChange={setPlanOnly}
        onSubmit={noop}
      />
    </ModalStory>
  )
}

export const Behind = () => <Story />

export const Pending = () => <Story isPending />

export const Loading = () => <Story isLoading reconciliation={undefined} />

export const Blocked = () => (
  <Story
    reconciliation={{
      ...behind,
      status: 'blocked',
      blockers: [
        {
          code: 'install_busy',
          message: 'Another workflow is running on this install.',
          workflow_id: 'wf-busy',
        },
      ],
    }}
  />
)

export const InProgress = () => (
  <Story
    reconciliation={{
      ...behind,
      status: 'in_progress',
      in_flight_workflow_id: 'wf-1',
    }}
  />
)

export const Current = () => (
  <Story
    reconciliation={{
      ...behind,
      status: 'current',
      applied: behind.target,
      skipped_runs: [],
      skipped_run_count: 0,
    }}
  />
)

export const SubmitFailed = () => (
  <Story
    error={{
      error: 'The app branch moved to a newer config. Review the changes again.',
      description: '',
      user_error: true,
    }}
  />
)

export const OutOfSyncFailedUpdate = () => (
  <DryRunStory reconciliation={failedUpdateReconciliation()} />
)

export const OutOfSyncManualChange = () => (
  <DryRunStory reconciliation={manualChangeReconciliation()} />
)

export const BehindAndOutOfSync = () => (
  <DryRunStory reconciliation={behindAndOutOfSyncReconciliation()} />
)

export const OutOfSyncDryRun = () => (
  <Story
    reconciliation={failedUpdateReconciliation()}
    planOnly
    onPlanOnlyChange={() => {}}
  />
)
