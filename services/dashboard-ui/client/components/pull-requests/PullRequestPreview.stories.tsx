export default {
  title: 'Features / Pull requests / Preview',
}

import type { TInstallWorkflow, TPreviewSourcePR } from '@/types'
import { PullRequestPreview } from './PullRequestPreview'

const source: TPreviewSourcePR = {
  pr_number: 128,
  title: 'Retry checkout when the payment provider times out',
  head_ref: 'feat/checkout-retries',
  head_sha: 'a1b2c3d4e5f67890abcd1234567890abcdef12',
  url: 'https://github.com/acme/payments/pull/128',
  author: 'devon',
  is_draft: true,
  labels: [{ name: 'payments', color: '0e8a16' }],
  created_at: '2026-10-26T15:04:00Z',
}

const failedRun: TInstallWorkflow = {
  id: 'wf-128',
  created_at: '2026-10-27T18:12:00Z',
  finished: true,
  status: { status: 'error', status_human_description: 'Component builds failed.' },
  steps: [
    {
      id: 'step-commit',
      name: 'fetch commit',
      idx: 0,
      group_idx: 0,
      finished: true,
      status: { status: 'success' },
    },
    {
      id: 'step-config',
      name: 'fetch app config',
      idx: 1,
      group_idx: 1,
      finished: true,
      status: { status: 'success' },
    },
    {
      id: 'step-build',
      name: 'building components and sandbox',
      idx: 2,
      group_idx: 2,
      finished: true,
      status: {
        status: 'error',
        status_human_description: 'One or more component builds failed.',
      },
    },
    {
      id: 'step-plan',
      name: 'plan preview install',
      idx: 3,
      group_idx: 3,
      status: { status: 'pending' },
    },
  ],
  app_branch_runs: [
    {
      id: 'brun-128',
      pr_number: 128,
      created_at: '2026-10-27T18:12:00Z',
      status: 'error',
      head_sha: 'a1b2c3d4e5f67890abcd1234567890abcdef12',
      preview: {
        mode: 'plan-only',
        install_name: 'staging',
        resolved_preview_config: {
          mode: 'plan-only',
          install_name: 'staging',
        },
        branch_preview_config: {
          mode: 'plan-only',
          install_name: 'staging',
        },
      },
    },
  ],
}

export const FailedBuild = () => (
  <PullRequestPreview
    source={source}
    workflows={[failedRun]}
    repo="acme/payments"
    baseRef="main"
    basePath="/org-1/apps/app-1/branches/br-1"
    branchConfig={{
      mode: 'plan-only',
      install_id: 'inst-staging',
      install_name: 'staging',
    }}
  />
)

const inProgressRun: TInstallWorkflow = {
  id: 'wf-128-running',
  created_at: '2026-10-28T09:40:00Z',
  finished: false,
  status: {
    status: 'building',
    status_human_description: 'synced 3 components, 1 action, 0 runbooks',
  },
  steps: [
    {
      id: 'step-commit',
      name: 'fetch commit',
      idx: 0,
      group_idx: 0,
      finished: true,
      status: { status: 'success' },
    },
    {
      id: 'step-config',
      name: 'fetch app config',
      idx: 1,
      group_idx: 1,
      finished: true,
      status: { status: 'success' },
    },
    {
      id: 'step-build',
      name: 'building components and sandbox',
      idx: 2,
      group_idx: 2,
      status: { status: 'building' },
    },
    {
      id: 'step-plan',
      name: 'plan preview install',
      idx: 3,
      group_idx: 3,
      status: { status: 'pending' },
    },
  ],
  app_branch_runs: [
    {
      id: 'brun-128-running',
      pr_number: 128,
      created_at: '2026-10-28T09:40:00Z',
      status: 'running',
      head_sha: 'a1b2c3d4e5f67890abcd1234567890abcdef12',
      preview: {
        mode: 'plan-only',
        install_name: 'staging',
        resolved_preview_config: {
          mode: 'plan-only',
          install_name: 'staging',
        },
        branch_preview_config: {
          mode: 'plan-only',
          install_name: 'staging',
        },
      },
    },
  ],
}

export const InProgress = () => (
  <PullRequestPreview
    source={source}
    workflows={[inProgressRun]}
    repo="acme/payments"
    baseRef="main"
    basePath="/org-1/apps/app-1/branches/br-1"
    branchConfig={{
      mode: 'plan-only',
      install_id: 'inst-staging',
      install_name: 'staging',
    }}
  />
)

export const NoRuns = () => (
  <PullRequestPreview
    source={{ ...source, is_draft: false, labels: [] }}
    workflows={[]}
    repo="acme/payments"
    baseRef="main"
    basePath="/org-1/apps/app-1/branches/br-1"
    branchConfig={{ mode: 'plan-only' }}
  />
)
