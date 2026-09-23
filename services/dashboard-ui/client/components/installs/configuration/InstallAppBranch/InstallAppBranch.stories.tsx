export default {
  title: 'Installs/Configuration/InstallAppBranch',
}

import { InstallVersionsTimelineComponent } from '@/components/install-versions/InstallVersionsTimeline'
import type { TAppBranchRun, TInstallAppConfigVersion } from '@/types'
import { InstallAppBranch } from './InstallAppBranch'

const day = 86400000
const ago = (days: number) => new Date(Date.now() - day * days).toISOString()

const run = (overrides: Partial<TAppBranchRun>): TAppBranchRun =>
  ({
    id: 'abr-1',
    status: 'success',
    created_at: ago(1),
    app_branch: { id: 'brnch-1', name: 'release' },
    head_sha: 'a1b2c3d4e5f67890',
    vcs_connection_commit: {
      sha: 'a1b2c3d4e5f67890',
      message: 'Raise the checkout worker retry budget',
      author_name: 'Example Developer',
    },
    ...overrides,
  }) as TAppBranchRun

const appliedRun = run({ id: 'abr-applied' })

const newerRun = run({
  id: 'abr-latest',
  created_at: ago(0.1),
  head_sha: 'f6e5d4c3b2a10987',
  vcs_connection_commit: {
    sha: 'f6e5d4c3b2a10987',
    message: 'Add a read replica to the payments database',
    author_name: 'Example Developer',
  },
})

const connectedVCS = {
  connected_github_vcs_config: {
    repo: 'acme/platform-configs',
    branch: 'main',
    directory: 'apps/payments',
  },
}

const publicVCS = {
  public_git_vcs_config: {
    repo: 'https://github.com/acme/example-app-configs.git',
    branch: 'main',
    directory: '.',
  },
}

const versions = [
  {
    id: 'iacv-3',
    created_at: ago(0.2),
    status: { status: 'active' },
    old_app_config_id: 'appcfg-14',
    new_app_config_id: 'appcfg-15',
    app_branch_run_id: 'abr-latest',
    app_branch_run: {
      app_branch: { id: 'brnch-1', name: 'release' },
      vcs_connection_commit: {
        sha: 'f6e5d4c3b2a10987',
        message: 'Add a read replica to the payments database',
        author_name: 'Example Developer',
      },
    },
  },
  {
    id: 'iacv-2',
    created_at: ago(4),
    status: { status: 'in-progress' },
    workflow: { status: { status: 'error' } },
    old_app_config_id: 'appcfg-13',
    new_app_config_id: 'appcfg-14',
    metadata: { triggered_by: 'cli-sync' },
  },
  {
    id: 'iacv-1',
    created_at: ago(30),
    status: { status: 'active' },
    new_app_config_id: 'appcfg-13',
  },
] as unknown as TInstallAppConfigVersion[]

const history = (
  <InstallVersionsTimelineComponent
    versions={versions}
    orgId="org-1"
    installId="inst-1"
    appId="app-1"
  />
)

const emptyHistory = <InstallVersionsTimelineComponent versions={[]} />

export const Current = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-15"
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    latestRun={appliedRun}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const UpdateAvailable = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-14"
    branchName="release"
    branchHref="#"
    branchConfig={publicVCS}
    latestRun={newerRun}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const RunInProgress = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-14"
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    latestRun={run({
      id: 'abr-running',
      status: 'in-progress',
      created_at: ago(0.01),
    })}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const RunFailed = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-14"
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    latestRun={run({ id: 'abr-failed', status: 'error', created_at: ago(0.05) })}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const NeverApplied = () => (
  <InstallAppBranch
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    latestRun={newerRun}
    latestRunHref="#"
    history={emptyHistory}
  />
)

export const NoRuns = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-13"
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    history={emptyHistory}
  />
)

export const CommitMetadataMissing = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-15"
    branchName="release"
    branchHref="#"
    branchConfig={connectedVCS}
    latestRun={run({ vcs_connection_commit: undefined })}
    latestRunHref="#"
    appliedRun={run({ vcs_connection_commit: undefined })}
    appliedRunHref="#"
    history={history}
  />
)

export const LongCommitMessage = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-15"
    branchName="release-candidate-with-a-long-branch-name"
    branchHref="#"
    branchConfig={{
      connected_github_vcs_config: {
        repo: 'acme/platform-configuration-monorepo-with-a-long-name',
        branch: 'feature/rework-the-checkout-queue-drain-path',
        directory: 'apps/payments/environments/production',
      },
    }}
    latestRun={run({
      vcs_connection_commit: {
        sha: 'f6e5d4c3b2a10987',
        message:
          'refactor: rework the checkout queue drain path so retries are bounded by a shared budget instead of a per-worker counter (#1284)',
        author_name: 'Example Developer With A Long Name',
      },
    })}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const NoSourceConfig = () => (
  <InstallAppBranch
    appliedConfigId="appcfg-15"
    branchName="release"
    branchHref="#"
    latestRun={appliedRun}
    latestRunHref="#"
    appliedRun={appliedRun}
    appliedRunHref="#"
    history={history}
  />
)

export const Loading = () => <InstallAppBranch isLoading />

export const NoBranchConnected = () => <InstallAppBranch />
