import { CurrentAppBranchRun } from './CurrentAppBranchRun'
import type { TAppBranchRun } from '@/types'

export default {
  title: 'InstallUpdates/CurrentAppBranchRun',
}

const run = {
  id: 'abr-1',
  status: 'success',
  head_sha: 'a1b2c3d4e5f6',
  pr_number: 42,
  app_branch: { id: 'branch-1', name: 'feat/add-cache' },
  preview: { mode: 'apply' },
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6',
    message: 'Add a cache component to the deployment plan',
  },
} as TAppBranchRun

export const Applied = () => (
  <CurrentAppBranchRun run={run} orgId="org-1" appId="app-1" />
)

export const NoRunOnBranch = () => (
  <CurrentAppBranchRun branchName="feat/add-cache" />
)

export const NoBranchConnected = () => <CurrentAppBranchRun />

export const Loading = () => <CurrentAppBranchRun isLoading />
