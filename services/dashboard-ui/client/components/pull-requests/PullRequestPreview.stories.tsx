export default {
  title: 'Features / Pull requests / Preview',
}

import { useEffect, useState } from 'react'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { TemplateChangesContent } from '@/components/branches/ConfigChanges'
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
      status: {
        status: 'building',
        metadata: {
          builds: [
            {
              component_id: 'cmp-api',
              component_name: 'api',
              component_type: 'docker_build',
              status: 'building',
              change_reason: 'source_changed',
            },
            {
              component_id: 'sandbox',
              component_type: 'sandbox',
              component_name: 'Sandbox',
              status: 'skipped',
              change_reason: 'no_changes',
            },
          ],
        },
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

const RUN_STEPS = [
  'fetch commit',
  'fetch app config',
  'building components and sandbox',
  'plan preview install',
] as const

const BUILD_STEP = 2

type TMockBuildStatus = 'queued' | 'building' | 'success' | 'skipped'

const mockBuilds = (api: TMockBuildStatus, worker: TMockBuildStatus) => [
  {
    component_id: 'cmp-api',
    component_name: 'api',
    component_type: 'docker_build',
    status: api,
    change_reason: 'source_changed',
  },
  {
    component_id: 'cmp-worker',
    component_name: 'worker',
    component_type: 'helm_chart',
    status: worker,
    change_reason: 'config_changed',
  },
  {
    component_id: 'sandbox',
    component_type: 'sandbox',
    component_name: 'Sandbox',
    status: 'skipped',
    change_reason: 'no_changes',
  },
]

const runningFrame = ({
  active,
  description,
  builds = mockBuilds('success', 'success'),
  finished = false,
}: {
  active: number
  description: string
  builds?: ReturnType<typeof mockBuilds>
  finished?: boolean
}): TInstallWorkflow => ({
  id: 'wf-128-running',
  created_at: '2026-10-28T09:40:00Z',
  finished,
  status: {
    status: finished ? 'success' : 'building',
    status_human_description: description,
  },
  steps:
    active < 0
      ? []
      : RUN_STEPS.map((name, index) => {
          const done = finished || index < active
          return {
            id: `step-${index}`,
            name,
            idx: index,
            group_idx: index,
            finished: done,
            status: {
              status: done ? 'success' : index === active ? 'building' : 'pending',
              metadata:
                index === BUILD_STEP && active >= BUILD_STEP
                  ? { builds }
                  : undefined,
            },
          }
        }),
  app_branch_runs: inProgressRun.app_branch_runs,
})

const runningFrames: { workflow: TInstallWorkflow; ms: number }[] = [
  {
    workflow: runningFrame({ active: -1, description: 'pending sync' }),
    ms: 2000,
  },
  {
    workflow: runningFrame({ active: 0, description: 'fetching commit' }),
    ms: 2000,
  },
  {
    workflow: runningFrame({
      active: 1,
      description: 'fetched commit a1b2c3d4',
    }),
    ms: 2500,
  },
  {
    workflow: runningFrame({
      active: BUILD_STEP,
      description: 'synced 2 components, 0 actions, 0 runbooks',
      builds: mockBuilds('building', 'queued'),
    }),
    ms: 2500,
  },
  {
    workflow: runningFrame({
      active: BUILD_STEP,
      description: 'building 2 components',
      builds: mockBuilds('success', 'building'),
    }),
    ms: 2500,
  },
  {
    workflow: runningFrame({
      active: 3,
      description: 'planning against staging',
    }),
    ms: 3500,
  },
  {
    workflow: runningFrame({
      active: 3,
      description: 'plan complete',
      finished: true,
    }),
    ms: 6000,
  },
]

const mockDiff: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 0,
    removals: 0,
    changed: 2,
    fields: [],
    entities: [
      {
        name: 'api',
        op: 'change',
        componentType: 'docker_build',
        fields: [
          { key: 'dockerfile', op: 'change', diff: "'Dockerfile' -> 'Dockerfile.retry'" },
        ],
      },
      {
        name: 'worker',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'namespace', op: 'change', diff: "'payments' -> 'payments-retry'" },
          { key: 'chart_version', op: 'change', diff: "'2.3.0' -> '2.4.1'" },
        ],
      },
    ],
  },
  {
    name: 'Install inputs',
    sectionKey: 'inputs',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 1,
    fields: [],
    entities: [
      {
        name: 'input.retry_budget',
        op: 'add',
        fields: [{ key: 'default', op: 'add', diff: `'' -> '"8"'` }],
      },
      {
        name: 'input.timeout_seconds',
        op: 'change',
        fields: [{ key: 'default', op: 'change', diff: `'"30"' -> '"45"'` }],
      },
    ],
  },
]

export const Running = () => {
  const [frame, setFrame] = useState(0)
  useEffect(() => {
    const timer = setTimeout(
      () => setFrame((current) => (current + 1) % runningFrames.length),
      runningFrames[frame].ms
    )
    return () => clearTimeout(timer)
  }, [frame])

  return (
    <div className="flex min-h-[calc(100vh-4rem)] flex-col">
    <PullRequestPreview
      source={source}
      workflows={[runningFrames[frame].workflow]}
      repo="acme/payments"
      baseRef="main"
      basePath="/org-1/apps/app-1/branches/br-1"
      branchConfig={{
        mode: 'plan-only',
        install_id: 'inst-staging',
        install_name: 'staging',
      }}
      changes={
        <TemplateChangesContent
          sections={mockDiff}
          previousSha="9f2c1ab"
          sha="a1b2c3d"
        />
      }
    />
    </div>
  )
}

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
