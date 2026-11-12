export default {
  title: 'Views / Installs / Reconcile',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import type { TFixture } from './install-fixture'
import { viewPath } from './install-fixture'
import { resourcesFixture } from './page-fixtures'
import type {
  TInstallOverviewCommit,
  TInstallReconciliation,
} from '@/types'

const commit = (index: number): TInstallOverviewCommit => ({
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

const reconciliationOf = (
  overrides: Partial<TInstallReconciliation> = {}
): TInstallReconciliation => ({
  status: 'behind',
  app_branch_id: 'br-1',
  app_branch_name: 'main',
  applied: { app_config_id: 'cfg-1', version: 3, commit: commit(5) },
  target: { app_config_id: 'cfg-4', version: 6, commit: commit(1) },
  skipped_runs: [commit(1), commit(2), commit(3)],
  skipped_run_count: 3,
  blockers: [],
  ...overrides,
})

const page = (runsBehind: number, reconciliation: TInstallReconciliation) => {
  const base = resourcesFixture('active')
  const fixture: TFixture = (url, init) => {
    if (url.pathname.endsWith('/reconciliation')) {
      return { body: reconciliation }
    }
    if (
      url.pathname.includes('/app-configs/') &&
      url.pathname.endsWith('/diff')
    ) {
      return { body: {} }
    }
    const reply = base(url, init)
    if (
      url.pathname.endsWith('/orgs/current') &&
      reply?.body &&
      typeof reply.body === 'object'
    ) {
      const org = reply.body as { features?: Record<string, boolean> }
      return {
        ...reply,
        body: {
          ...org,
          features: { ...org.features, 'install-reconciliation': true },
        },
      }
    }
    if (
      url.pathname.endsWith('/overview') &&
      reply?.body &&
      typeof reply.body === 'object'
    ) {
      const body = reply.body as {
        branch_tracking?: Record<string, unknown>
      }
      return {
        ...reply,
        body: {
          ...body,
          branch_tracking: {
            ...body.branch_tracking,
            branch_id: 'br-1',
            target_branch: 'main',
            repo: 'acme/payments',
            status: 'current',
            runs_behind: runsBehind,
            successful_runs:
              runsBehind > 0 ? [commit(1), commit(2), commit(3)] : [],
          },
        },
      }
    }
    return reply
  }
  return (
    <InstallView fixture={fixture} path={viewPath('/resources/components')} />
  )
}

const behind = reconciliationOf()

export const Available = () => page(0, behind)

export const WithRunsBehind = () => page(3, behind)

export const Blocked = () =>
  page(
    0,
    reconciliationOf({
      status: 'blocked',
      blockers: [
        {
          code: 'install_busy',
          message: 'Another workflow is running on this install.',
          workflow_id: 'wf-busy',
        },
      ],
    })
  )

export const InProgress = () =>
  page(
    0,
    reconciliationOf({ status: 'in_progress', in_flight_workflow_id: 'wf-1' })
  )
