import { expect, test } from 'bun:test'
import { toRecentUpdateItem } from './map-recent-updates'

const source = {
  orgId: 'org-1',
  nestedInstalls: false,
  appId: 'app-1',
  appName: 'acme-payments',
  branchId: 'branch-1',
  branchName: 'main',
  runId: 'run-1',
  workflowId: 'wf-1',
  status: 'running',
  createdAt: '2026-09-01T10:00:00Z',
  commitMessage: 'feat: add checkout',
  commitSha: 'a1b2c3d4',
  commitAuthor: 'Example Developer',
  repo: 'acme/payments',
  configGroups: [
    { id: 'group-1', name: 'canary', order: 0, hasSelector: false },
    { id: 'group-2', name: 'enterprise', order: 1, hasSelector: true },
  ],
  groupRuns: [
    {
      install_group_id: 'group-1',
      install_group_name: 'canary',
      total_installs: 1,
      installs: [
        {
          install_id: 'inst-1',
          status: 'success',
          workflow_id: 'wf-inst-1',
        },
      ],
    },
  ],
  installsById: {
    'inst-1': { name: 'staging-example', health: 'healthy' },
    'inst-2': { name: 'production-acme', health: 'degraded' },
  },
  approvals: [
    {
      id: 'approval-1',
      type: 'app_branch_plan' as const,
      owner_id: 'inst-2',
      owner_type: 'installs',
      workflow_step: { install_workflow_id: 'wf-inst-1' },
    },
    {
      id: 'approval-other',
      type: 'terraform_plan' as const,
      owner_id: 'inst-9',
      owner_type: 'installs',
      workflow_step: { install_workflow_id: 'wf-other' },
    },
  ],
}

test('maps a branch run into a recent update card', () => {
  const item = toRecentUpdateItem(source)

  expect(item?.appName).toBe('acme-payments')
  expect(item?.branchName).toBe('main')
  expect(item?.appHref).toBe('/org-1/apps/app-1')
  expect(item?.branchHref).toBe('/org-1/apps/app-1/branches/branch-1')
  expect(item?.runHref).toBe('/org-1/apps/app-1/branches/branch-1/runs/wf-1')
  expect(item?.commitHref).toBe(
    'https://github.com/acme/payments/commit/a1b2c3d4'
  )
  expect(item?.runStatus).toBe('awaiting-approval')
  expect(item?.planGroups?.map((group) => group.name)).toEqual([
    'canary',
    'enterprise',
  ])
  expect(item?.updatedInstalls).toEqual([
    {
      id: 'inst-1',
      name: 'staging-example',
      group: 'canary',
      href: '/org-1/installs/inst-1',
      runStatus: 'success',
      health: 'healthy',
      rolledOut: true,
    },
  ])
  expect(item?.pendingApprovals).toEqual([
    {
      id: 'approval-1',
      installName: 'production-acme',
      type: 'install group plan',
      href: '/org-1/installs/inst-2',
    },
  ])
})

test('keeps a running status when nothing is awaiting approval', () => {
  const item = toRecentUpdateItem({ ...source, approvals: [] })
  expect(item?.runStatus).toBe('running')
})

test('skips a run without a timestamp', () => {
  expect(toRecentUpdateItem({ ...source, createdAt: undefined })).toBeUndefined()
})
