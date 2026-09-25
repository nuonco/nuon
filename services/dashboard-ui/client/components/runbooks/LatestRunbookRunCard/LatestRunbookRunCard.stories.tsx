export default {
  title: 'Runbooks/LatestRunbookRunCard',
}

import type { TInstallRunbookRun } from '@/lib/ctl-api/installs/runbooks'
import { LatestRunbookRunCard } from './LatestRunbookRunCard'

const run: TInstallRunbookRun = {
  id: 'rbr-1',
  created_at: '2026-09-22T14:00:00Z',
  updated_at: '2026-09-22T14:08:00Z',
  status: 'success',
  install_workflow_id: 'wf-1',
  install_workflow: {
    id: 'wf-1',
    status: {
      status: 'success',
      status_human_description: 'Runbook finished.',
    },
  },
}

export const Default = () => (
  <LatestRunbookRunCard
    run={run}
    href="/org-1/installs/inst-1/history/wf-1"
  />
)

export const Flush = () => (
  <LatestRunbookRunCard
    flush
    run={run}
    href="/org-1/installs/inst-1/history/wf-1"
  />
)

export const InProgress = () => (
  <LatestRunbookRunCard
    run={{
      ...run,
      updated_at: '2026-09-22T14:02:00Z',
      install_workflow: {
        id: 'wf-1',
        status: {
          status: 'in-progress',
          status_human_description: 'Runbook is running.',
        },
      },
    }}
    href="/org-1/installs/inst-1/history/wf-1"
  />
)

export const Failed = () => (
  <LatestRunbookRunCard
    run={{
      ...run,
      install_workflow: {
        id: 'wf-1',
        status: {
          status: 'error',
          status_human_description: 'Runbook failed.',
        },
      },
    }}
    href="/org-1/installs/inst-1/history/wf-1"
  />
)

export const Empty = () => <LatestRunbookRunCard />

export const Loading = () => <LatestRunbookRunCard isLoading run={run} />
