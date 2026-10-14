export default {
  title: 'Features / Installs / Branch banner',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture } from '@/views/app/app-fixtures'
import {
  VIEW_APP_ID,
  viewPath,
  type TFixture,
} from '@/views/install/install-fixture'
import {
  deploymentsFixture,
  resourcesFixture,
} from '@/views/install/page-fixtures'
import type { TInstallBranchTracking, TInstallOverviewCommit } from '@/types'

const messages = [
  'Enable request tracing for the API',
  'Pin the ledger module to 2.4.0',
  'Raise the checkout worker retry budget',
]

const runCommit = (
  index: number,
  runStatus = 'success'
): TInstallOverviewCommit => ({
  app_config_id: `cfg-${index + 2}`,
  branch_id: 'br-1',
  workflow_id: `wf-branch-${index}`,
  run_id: `run-${index}`,
  sha: `${index}bc123def4567890`,
  message: messages[index % messages.length],
  author: 'Example Developer',
  created_at: new Date(Date.now() - (index + 1) * 3_600_000).toISOString(),
  run_status: runStatus,
  metadata:
    index % 3 === 0
      ? { trigger: 'push' }
      : index % 3 === 1
        ? { trigger: 'manual' }
        : { trigger: 'tag', tag: 'v2.4.0' },
})

const page = (
  count?: number,
  error = false,
  excludedStatuses: string[] = []
) => {
  const base = resourcesFixture('active')
  const deployments = deploymentsFixture('results')
  const appRuns = appFixture('runs')
  const history = [
    ...excludedStatuses.map((status, index) => runCommit(index, status)),
    ...Array.from({ length: count ?? 0 }, (_, index) =>
      runCommit(index + excludedStatuses.length)
    ),
  ]
  const tracking: TInstallBranchTracking = {
    branch_id: 'br-1',
    target_branch: 'main',
    repo: 'acme/payments',
    status: 'current',
    selected_commit: {
      ...runCommit(3),
      run_id: 'run-selected',
      workflow_id: 'wf-selected',
      app_config_id: 'cfg-1',
      sha: 'a1b2c3d4e5f6a7b8',
      message: 'Pin the payments chart',
      created_at: new Date(
        Date.now() - (history.length + 1) * 3_600_000
      ).toISOString(),
    },
    runs_behind: count,
    successful_runs: history
      .filter((run) => run.run_status === 'success')
      .slice(0, 50),
  }
  const fixture: TFixture = (url, init) => {
    if (url.pathname.endsWith('/overview')) {
      return error
        ? { body: { error: 'Unable to load branch tracking' }, status: 500 }
        : {
            body: {
              branch_tracking: tracking,
              config_drift: { components: [] },
            },
          }
    }
    if (url.pathname === `/v1/apps/${VIEW_APP_ID}/configs`) {
      return {
        paginated: true,
        body: [{ id: history[0]?.app_config_id ?? 'cfg-1', app_id: VIEW_APP_ID }],
      }
    }
    if (url.pathname === `/v1/apps/${VIEW_APP_ID}/branches/br-1/runs`) {
      return {
        paginated: true,
        body: [...history, tracking.selected_commit!].map((run) => ({
          id: run.workflow_id,
          created_at: run.created_at,
          owner_type: 'app_branches',
          type: 'app_branches_manual_update',
          finished: !['pending', 'running'].includes(run.run_status ?? ''),
          status: {
            status:
              run.run_status === 'failed'
                ? 'error'
                : run.run_status === 'running'
                  ? 'in-progress'
                  : run.run_status,
          },
          app_branch_runs: [
            {
              id: run.run_id,
              status: run.run_status,
              head_sha: run.sha,
              created_at: run.created_at,
              app_config_id: run.app_config_id,
              metadata: run.metadata,
              app_branch: { id: 'br-1', name: 'main' },
              vcs_connection_commit: {
                sha: run.sha,
                message: run.message,
                author_name: run.author,
                created_at: run.created_at,
              },
            },
          ],
        })),
      }
    }
    if (
      url.pathname.endsWith('/orgs/current') ||
      url.pathname.startsWith(`/v1/apps/${VIEW_APP_ID}`)
    ) {
      return appRuns(url, init)
    }
    return base(url, init) ?? deployments(url, init)
  }
  return (
    <InstallView fixture={fixture} path={viewPath('/resources/components')} />
  )
}

export const Behind = () => page(3)
export const OneRunBehind = () => page(1)
export const Current = () => page(0)
export const UnknownProvenance = () => page()
export const TwentyFiveRunsBehind = () => page(25)
export const ManyRunsBehind = () => page(64)
export const FailedAndCanceledRuns = () =>
  page(3, false, ['failed', 'cancelled', 'running'])
export const OnlyFailedAndCanceledRuns = () =>
  page(0, false, ['failed', 'cancelled'])
export const LoadFailed = () => page(0, true)
