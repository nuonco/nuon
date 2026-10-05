export default { title: 'Features / Operations / Activity list' }

import { PanelStory } from '@/components/__stories__/helpers'
import type { TInstallActivity } from '@/types'
import { ActivityDetailPanel } from './ActivityDetailPanel'
import {
  ActivityListPresenter,
  type IActivityFilter,
} from './ActivityListPresenter'

const ORG_ID = 'org-1'
const INSTALL_ID = 'inst-1'

const hoursAgo = (hours: number) =>
  new Date(Date.now() - hours * 3600_000).toISOString()

const ACTION: TInstallActivity = {
  id: 'actrun-1',
  type: 'action_run',
  status: 'finished',
  created_at: hoursAgo(2),
  title: 'Restart workers',
  summary: 'Action run finished.',
  workflow: {
    id: 'wf-1',
    type: 'action_workflow_run',
    name: 'Restart workers',
  },
  action: {
    run_id: 'actrun-1',
    action_workflow_id: 'act-1',
    name: 'Restart workers',
    trigger_type: 'manual',
  },
}

const RUNBOOK: TInstallActivity = {
  id: 'rbr-1',
  type: 'runbook_run',
  status: 'queued',
  created_at: hoursAgo(26),
  title: 'Reconcile drift',
  summary: 'Runbook run queued.',
  workflow: { id: 'wf-2', type: 'runbook_run', name: 'Reconcile drift' },
  runbook: { run_id: 'rbr-1', runbook_id: 'rb-1', name: 'Reconcile drift' },
}

const POLICY: TInstallActivity = {
  id: 'pol-1',
  type: 'policy_check',
  status: 'error',
  created_at: hoursAgo(50),
  title: 'payments-api',
  summary: 'Policy checks failed',
  policy: {
    report_id: 'pol-1',
    owner_type: 'install_deploys',
    owner_id: 'dep-1',
    component_name: 'payments-api',
    deny_count: 2,
    warn_count: 1,
    pass_count: 4,
  },
}

const ITEMS = [ACTION, RUNBOOK, POLICY]

const DEFAULT_FILTER: IActivityFilter = {
  search: '',
  status: new Set(),
  type: new Set(),
}

const presenterProps = {
  orgId: ORG_ID,
  installId: INSTALL_ID,
  search: '',
  filter: DEFAULT_FILTER,
  onSearchChange: () => {},
  onStatusChange: () => {},
  onTypeChange: () => {},
  onDateChange: () => {},
  onClearFilters: () => {},
}

export const Default = () => (
  <div className="max-w-3xl mx-auto p-6">
    <ActivityListPresenter
      {...presenterProps}
      activity={ITEMS}
      isLoading={false}
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 20 }}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-3xl mx-auto p-6">
    <ActivityListPresenter
      {...presenterProps}
      activity={[]}
      isLoading
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 5 }}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl mx-auto p-6">
    <ActivityListPresenter
      {...presenterProps}
      activity={[]}
      isLoading={false}
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 20 }}
    />
  </div>
)

export const EmptyFiltered = () => (
  <div className="max-w-3xl mx-auto p-6">
    <ActivityListPresenter
      {...presenterProps}
      activity={[]}
      isLoading={false}
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 20 }}
      filter={{ ...DEFAULT_FILTER, status: new Set(['failed']) }}
    />
  </div>
)

export const DetailPanelStory = () => (
  <PanelStory label="Open activity details">
    <ActivityDetailPanel
      activity={POLICY}
      orgId={ORG_ID}
      installId={INSTALL_ID}
    />
  </PanelStory>
)
