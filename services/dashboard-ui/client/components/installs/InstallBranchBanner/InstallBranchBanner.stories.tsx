export default {
  title: 'Features / Installs / Branch banner',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { viewPath, type TFixture } from '@/views/install/install-fixture'
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

const commit = (index: number): TInstallOverviewCommit => ({
  app_config_id: `cfg-${index + 2}`,
  branch_id: 'br-1',
  workflow_id: `wf-branch-${index}`,
  run_id: `run-${index}`,
  sha: `${index}bc123def4567890`,
  message: messages[index % messages.length],
  author: 'Example Developer',
  created_at: new Date(Date.now() - (index + 1) * 3_600_000).toISOString(),
  run_status: index === 0 ? 'in-progress' : 'success',
  awaiting_approval: index === 1,
})

const page = (count?: number, error = false) => {
  const base = resourcesFixture('active')
  const deployments = deploymentsFixture('results')
  const tracking: TInstallBranchTracking = {
    branch_id: 'br-1',
    target_branch: 'main',
    repo: 'acme/payments',
    status: 'current',
    selected_commit: {
      ...commit(3),
      app_config_id: 'cfg-1',
      sha: 'a1b2c3d4e5f6a7b8',
      message: 'Pin the payments chart',
      created_at: new Date(
        Date.now() - ((count ?? 0) + 1) * 3_600_000
      ).toISOString(),
    },
    commits_behind: count,
    pending_commits: Array.from(
      { length: Math.min(count ?? 0, 50) },
      (_, index) => commit(index)
    ),
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
    return base(url, init) ?? deployments(url, init)
  }
  return (
    <InstallView fixture={fixture} path={viewPath('/resources/components')} />
  )
}

export const Behind = () => page(3)
export const OneCommitBehind = () => page(1)
export const Current = () => page(0)
export const UnknownProvenance = () => page()
export const TwentyFiveCommitsBehind = () => page(25)
export const ManyCommitsBehind = () => page(64)
export const LoadFailed = () => page(0, true)
