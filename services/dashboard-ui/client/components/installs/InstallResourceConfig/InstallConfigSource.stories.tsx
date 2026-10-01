import type { TAppBranchRun } from '@/types'
import { InstallConfigSource } from './InstallConfigSource'

export default {
  title: 'Features / Installs / Config source',
}

const run = {
  id: 'run-1',
  status: 'succeeded',
  created_at: '2026-09-18T15:04:00Z',
  head_sha: 'a1b2c3d4e5f6a7b8',
  app_branch: { id: 'brn-1', name: 'main' },
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6a7b8',
    message: 'Add checkout retry to the payments worker',
    author_name: 'Ada Lovelace',
    created_at: '2026-09-18T15:02:00Z',
  },
} as TAppBranchRun

export const Applied = () => (
  <div className="max-w-md">
    <InstallConfigSource run={run} runHref="#" />
  </div>
)

export const Behind = () => (
  <div className="max-w-md">
    <InstallConfigSource behind run={run} runHref="#" />
  </div>
)

export const Missing = () => (
  <div className="max-w-md">
    <InstallConfigSource />
  </div>
)

export const Loading = () => (
  <div className="max-w-md">
    <InstallConfigSource isLoading />
  </div>
)
