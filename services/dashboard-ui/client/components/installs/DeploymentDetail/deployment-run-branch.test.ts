import { expect, test } from 'bun:test'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { deploymentRunBranch } from './deployment-run-branch'

const installBranch = {
  id: 'br-install',
  name: 'main',
}

const runBranch = {
  id: 'br-run',
  name: 'release',
  run_id: 'run-1',
  sha: 'abc123def456',
}

test('uses the workflow branch run instead of the install branch', () => {
  const branch = deploymentRunBranch(
    { app_branch: installBranch } as TInstallDeploymentRecord,
    {
      app_branch_runs: [
        {
          id: 'run-9',
          status: 'success',
          created_at: '2026-10-10T10:00:00Z',
          app_branch: { id: 'br-feature', name: 'feature' },
          head_sha: 'fff111aaa222',
          vcs_connection_commit: {
            sha: 'fff111aaa222',
            message: 'Ship the feature\n\nLonger body',
            author_name: 'Ada',
            author_avatar_url: 'https://example.com/ada.png',
          },
        },
      ],
    } as TWorkflow
  )

  expect(branch).toEqual({
    id: 'br-feature',
    name: 'feature',
    runId: 'run-9',
    status: 'success',
    sha: 'fff111aaa222',
    message: 'Ship the feature',
    author: 'Ada',
    avatarUrl: 'https://example.com/ada.png',
    createdAt: '2026-10-10T10:00:00Z',
    repo: undefined,
  })
})

test('uses the deployment branch run when the workflow has none', () => {
  expect(
    deploymentRunBranch(
      { app_branch: runBranch } as TInstallDeploymentRecord,
      { app_branch_runs: [] } as TWorkflow
    )
  ).toEqual({
    id: 'br-run',
    name: 'release',
    runId: 'run-1',
    sha: 'abc123def456',
  })
})

test('ignores the install tracking branch', () => {
  expect(
    deploymentRunBranch(
      { app_branch: installBranch } as TInstallDeploymentRecord,
      undefined
    )
  ).toBeUndefined()
})
