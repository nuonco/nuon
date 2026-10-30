export default {
  title: 'Features / Installs / Management / Reconcile app version',
}

import { DateTime } from 'luxon'
import { ModalStory } from '@/components/__stories__/helpers'
import { ReconcileAppVersionModal } from './ReconcileAppVersion'
import { PendingScheduledReconcile } from './PendingScheduledReconcile'
import { ScheduledReconcileHistory } from './ScheduledReconcileHistory'
import type { TInstallScheduledReconcile } from '@/types'

const noop = () => {}

const pending: TInstallScheduledReconcile = {
  id: 'schedreconcile_01hpending',
  scheduled_at: DateTime.now().plus({ hours: 2 }).toISO() ?? undefined,
  status: 'pending',
  plan_only: false,
}

const history: TInstallScheduledReconcile[] = [
  {
    id: 'schedreconcile_01hfired',
    scheduled_at: DateTime.now().minus({ days: 1 }).toISO() ?? undefined,
    status: 'fired',
    plan_only: false,
    workflow_id: 'wf_01hworkflow',
  },
  {
    id: 'schedreconcile_01hnochange',
    scheduled_at: DateTime.now().minus({ days: 2 }).toISO() ?? undefined,
    status: 'no-changes',
    plan_only: true,
  },
  {
    id: 'schedreconcile_01hcancelled',
    scheduled_at: DateTime.now().minus({ days: 3 }).toISO() ?? undefined,
    status: 'cancelled',
    plan_only: false,
  },
  {
    id: 'schedreconcile_01hfailed',
    scheduled_at: DateTime.now().minus({ days: 4 }).toISO() ?? undefined,
    status: 'failed',
    status_reason: 'Install was deprovisioned before the schedule fired.',
    plan_only: false,
  },
]

export const OutOfDate = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const NeverProvisioned = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId={undefined}
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={0}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const InSync = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hinsync"
      targetConfigId="appcfg_01hinsync"
      unappliedCount={0}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithConflictError = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={{
        error:
          'The install is already on its pinned app config; there is nothing to reconcile.',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const ScheduleForm = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={null}
      history={history}
      getWorkflowHref={(workflowId) =>
        `/org-1/installs/install-1/workflows/${workflowId}`
      }
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithPendingSchedule = () => (
  <ModalStory>
    <ReconcileAppVersionModal
      installName="prod-acme"
      appliedConfigId="appcfg_01hpendingapplied"
      targetConfigId="appcfg_01hpendingtarget"
      unappliedCount={3}
      isPending={false}
      error={null}
      pending={pending}
      history={history}
      getWorkflowHref={(workflowId) =>
        `/org-1/installs/install-1/workflows/${workflowId}`
      }
      onSubmit={noop}
      onCancelPending={noop}
    />
  </ModalStory>
)

export const PendingBanner = () => (
  <div className="p-6">
    <PendingScheduledReconcile pending={pending} onCancel={noop} />
  </div>
)

export const History = () => (
  <div className="p-6">
    <ScheduledReconcileHistory
      history={history}
      getWorkflowHref={(workflowId) =>
        `/org-1/installs/install-1/workflows/${workflowId}`
      }
    />
  </div>
)
