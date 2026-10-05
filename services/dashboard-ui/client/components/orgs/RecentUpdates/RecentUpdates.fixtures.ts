import type { TBranchActivityItem } from '@/components/orgs/BranchActivityFeed'
import {
  toRecentUpdateItem,
  type TRecentConfigGroup,
  type TRecentGroupRun,
  type TRecentUpdatePayloadItem,
} from './map-recent-updates'

const now = Date.now()
const minsAgo = (mins: number) => new Date(now - mins * 60_000).toISOString()

const group = (
  id: string,
  name: string,
  order: number,
  hasSelector = false
): TRecentConfigGroup => ({ id, name, order, hasSelector })

const groupRun = (
  id: string,
  name: string,
  installs: Array<[installId: string, status: string]>
): TRecentGroupRun => ({
  install_group_id: id,
  install_group_name: name,
  total_installs: installs.length,
  installs: installs.map(([installId, status]) => ({
    install_id: installId,
    status,
    workflow_id: `wf-${installId}`,
  })),
})

const INSTALLS: TRecentUpdatePayloadItem['installsById'] = {
  'inst-staging': { name: 'staging-example', health: 'healthy' },
  'inst-acme': { name: 'production-acme', health: 'healthy' },
  'inst-globex': { name: 'production-globex', health: 'degraded' },
  'inst-initech': { name: 'production-initech', health: 'healthy' },
  'inst-umbrella': { name: 'production-umbrella', health: 'unhealthy' },
  'inst-hooli': { name: 'production-hooli', health: 'healthy' },
  'inst-preview': { name: 'preview-1042', health: 'unknown' },
  'inst-dev': { name: 'development', health: 'healthy' },
}

const commit = (
  sha: string,
  message: string,
  author: string,
  repo: string
): Partial<TRecentUpdatePayloadItem> => ({
  commitSha: sha,
  commitMessage: message,
  commitAuthor: author,
  repo,
})

export const AWAITING_APPROVAL: TRecentUpdatePayloadItem = {
  appId: 'app-payments',
  appName: 'acme-payments',
  branchId: 'branch-payments-main',
  branchName: 'main',
  runId: 'run-awaiting',
  workflowId: 'wf-run-awaiting',
  status: 'running',
  createdAt: minsAgo(4),
  ...commit(
    'a1b2c3d4',
    'feat: add card-on-file checkout flow (#142)',
    'Example Developer',
    'acme/payments'
  ),
  configGroups: [
    group('grp-canary', 'canary', 0),
    group('grp-enterprise', 'enterprise', 1),
  ],
  groupRuns: [
    groupRun('grp-canary', 'canary', [['inst-staging', 'success']]),
    groupRun('grp-enterprise', 'enterprise', [
      ['inst-acme', 'in-progress'],
      ['inst-globex', 'in-progress'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [
    {
      id: 'approval-acme',
      type: 'terraform_plan',
      owner_type: 'install_workflow_steps',
      workflow_step: {
        install_workflow_id: 'wf-inst-acme',
        owner_type: 'installs',
        owner_id: 'inst-acme',
      },
    },
    {
      id: 'approval-globex',
      type: 'helm_approval',
      owner_type: 'install_workflow_steps',
      workflow_step: { install_workflow_id: 'wf-inst-globex' },
    },
  ],
}

export const BRANCH_PLAN_APPROVAL: TRecentUpdatePayloadItem = {
  appId: 'app-billing',
  appName: 'acme-billing',
  branchId: 'branch-billing-main',
  branchName: 'main',
  runId: 'run-branch-plan',
  workflowId: 'wf-run-branch-plan',
  status: 'running',
  createdAt: minsAgo(9),
  awaitingApproval: true,
  ...commit(
    'b7c8d9e0',
    'feat: prorate seat changes mid-cycle',
    'Example Maintainer',
    'acme/billing'
  ),
  configGroups: [group('grp-billing-default', 'default', 0)],
  groupRuns: [],
  installsById: INSTALLS,
  approvals: [
    {
      id: 'approval-branch-plan',
      type: 'app_branch_plan',
      owner_id: 'run-branch-plan',
      owner_type: 'app_branch_runs',
    },
  ],
}

export const FAILED: TRecentUpdatePayloadItem = {
  appId: 'app-gateway',
  appName: 'acme-gateway',
  branchId: 'branch-gateway-rate-limit',
  branchName: 'feature/rate-limit',
  runId: 'run-failed',
  workflowId: 'wf-run-failed',
  status: 'failed',
  createdAt: minsAgo(18),
  ...commit(
    'f4e5d6c7',
    'fix: tighten rate-limit headers for downstream services',
    'Example Contributor',
    'acme/gateway'
  ),
  configGroups: [
    group('grp-gw-canary', 'canary', 0),
    group('grp-gw-customers', 'customers', 1, true),
  ],
  groupRuns: [
    groupRun('grp-gw-canary', 'canary', [['inst-dev', 'success']]),
    groupRun('grp-gw-customers', 'customers', [
      ['inst-umbrella', 'error'],
      ['inst-hooli', 'not-attempted'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const FAILED_BEFORE_ROLLOUT: TRecentUpdatePayloadItem = {
  appId: 'app-search',
  appName: 'acme-search',
  branchId: 'branch-search-main',
  branchName: 'main',
  runId: 'run-build-failed',
  workflowId: 'wf-run-build-failed',
  status: 'error',
  createdAt: minsAgo(32),
  ...commit(
    'c0ffee12',
    'chore: upgrade search indexer base image',
    'Example Developer',
    'acme/search'
  ),
  configGroups: [
    group('grp-search-canary', 'canary', 0),
    group('grp-search-rest', 'rest', 1),
  ],
  groupRuns: [],
  installsById: INSTALLS,
  approvals: [],
}

export const RUNNING: TRecentUpdatePayloadItem = {
  appId: 'app-portal',
  appName: 'acme-portal',
  branchId: 'branch-portal-main',
  branchName: 'main',
  runId: 'run-running',
  workflowId: 'wf-run-running',
  status: 'running',
  createdAt: minsAgo(2),
  ...commit(
    '7c8b9a01',
    'chore: bump node-fetch to 3.3.2',
    'Example Maintainer',
    'acme/portal'
  ),
  configGroups: [
    group('grp-portal-canary', 'canary', 0),
    group('grp-portal-rest', 'rest', 1),
  ],
  groupRuns: [
    groupRun('grp-portal-canary', 'canary', [['inst-preview', 'success']]),
    groupRun('grp-portal-rest', 'rest', [
      ['inst-acme', 'in-progress'],
      ['inst-globex', 'queued'],
      ['inst-initech', 'queued'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const PENDING: TRecentUpdatePayloadItem = {
  appId: 'app-portal',
  appName: 'acme-portal',
  branchId: 'branch-portal-release',
  branchName: 'release/2026-10',
  runId: 'run-pending',
  workflowId: 'wf-run-pending',
  status: 'pending',
  createdAt: minsAgo(1),
  ...commit(
    'd1e2f3a4',
    'release: cut October release branch',
    'Example Release Bot',
    'acme/portal'
  ),
  configGroups: [
    group('grp-release-canary', 'canary', 0),
    group('grp-release-rest', 'rest', 1),
  ],
  groupRuns: [],
  installsById: INSTALLS,
  approvals: [],
}

export const SUCCEEDED: TRecentUpdatePayloadItem = {
  appId: 'app-analytics',
  appName: 'acme-analytics',
  branchId: 'branch-analytics-main',
  branchName: 'main',
  runId: 'run-success',
  workflowId: 'wf-run-success',
  status: 'success',
  createdAt: minsAgo(45),
  ...commit(
    '2d3e4f56',
    'feat: add export-to-csv button on reports page (#89)',
    'Example Contributor',
    'acme/analytics'
  ),
  configGroups: [
    group('grp-an-canary', 'canary', 0),
    group('grp-an-early', 'early-access', 1),
    group('grp-an-ga', 'general', 2),
  ],
  groupRuns: [
    groupRun('grp-an-canary', 'canary', [['inst-staging', 'success']]),
    groupRun('grp-an-early', 'early-access', [
      ['inst-acme', 'success'],
      ['inst-globex', 'success'],
    ]),
    groupRun('grp-an-ga', 'general', [
      ['inst-initech', 'success'],
      ['inst-umbrella', 'success'],
      ['inst-hooli', 'success'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const MANUAL_NO_COMMIT: TRecentUpdatePayloadItem = {
  appId: 'app-billing',
  appName: 'acme-billing',
  branchId: 'branch-billing-hotfix',
  branchName: 'hotfix/tax-calc',
  runId: 'run-manual',
  workflowId: 'wf-run-manual',
  status: 'success',
  createdAt: minsAgo(120),
  configGroups: [group('grp-billing-hotfix', 'default', 0)],
  groupRuns: [
    groupRun('grp-billing-hotfix', 'default', [['inst-dev', 'success']]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const CANCELLED: TRecentUpdatePayloadItem = {
  appId: 'app-search',
  appName: 'acme-search',
  branchId: 'branch-search-vector',
  branchName: 'experiment/vector-index',
  runId: 'run-cancelled',
  workflowId: 'wf-run-cancelled',
  status: 'cancelled',
  createdAt: minsAgo(75),
  ...commit(
    'e5f6a7b8',
    'wip: try approximate nearest-neighbour index',
    'Example Developer',
    'acme/search'
  ),
  configGroups: [
    group('grp-vec-canary', 'canary', 0),
    group('grp-vec-rest', 'rest', 1),
  ],
  groupRuns: [
    groupRun('grp-vec-canary', 'canary', [['inst-preview', 'success']]),
    groupRun('grp-vec-rest', 'rest', [
      ['inst-acme', 'cancelled'],
      ['inst-globex', 'not-attempted'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const LABEL_SELECTOR_NO_MATCHES: TRecentUpdatePayloadItem = {
  appId: 'app-notify',
  appName: 'acme-notify',
  branchId: 'branch-notify-main',
  branchName: 'main',
  runId: 'run-no-matches',
  workflowId: 'wf-run-no-matches',
  status: 'success',
  createdAt: minsAgo(200),
  ...commit(
    'a9b8c7d6',
    'feat: add SMS fallback channel',
    'Example Maintainer',
    'acme/notify'
  ),
  configGroups: [group('grp-notify-eu', 'eu-customers', 0, true)],
  groupRuns: [groupRun('grp-notify-eu', 'eu-customers', [])],
  installsById: INSTALLS,
  approvals: [],
}

export const MANY_GROUPS: TRecentUpdatePayloadItem = {
  appId: 'app-edge',
  appName: 'acme-edge',
  branchId: 'branch-edge-main',
  branchName: 'main',
  runId: 'run-rings',
  workflowId: 'wf-run-rings',
  status: 'running',
  createdAt: minsAgo(12),
  ...commit(
    '0a1b2c3d',
    'perf: cache TLS session tickets at the edge',
    'Example Contributor',
    'acme/edge'
  ),
  configGroups: [
    group('grp-ring-0', 'ring-0', 0),
    group('grp-ring-1', 'ring-1', 1),
    group('grp-ring-2', 'ring-2', 2),
    group('grp-ring-3', 'ring-3', 3),
    group('grp-ring-4', 'ring-4', 4),
    group('grp-ring-5', 'ring-5', 5),
  ],
  groupRuns: [
    groupRun('grp-ring-0', 'ring-0', [['inst-dev', 'success']]),
    groupRun('grp-ring-1', 'ring-1', [['inst-staging', 'success']]),
    groupRun('grp-ring-2', 'ring-2', [
      ['inst-acme', 'success'],
      ['inst-globex', 'in-progress'],
    ]),
    groupRun('grp-ring-3', 'ring-3', [['inst-initech', 'queued']]),
    groupRun('grp-ring-4', 'ring-4', [['inst-umbrella', 'queued']]),
    groupRun('grp-ring-5', 'ring-5', [['inst-hooli', 'queued']]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const NO_INSTALL_GROUPS: TRecentUpdatePayloadItem = {
  appId: 'app-docs',
  appName: 'acme-docs',
  branchId: 'branch-docs-main',
  branchName: 'main',
  runId: 'run-no-groups',
  workflowId: 'wf-run-no-groups',
  status: 'success',
  createdAt: minsAgo(300),
  ...commit(
    'abc12345',
    'docs: update README with deployment steps',
    'Example Maintainer',
    'acme/docs'
  ),
  configGroups: [],
  groupRuns: [],
  installsById: INSTALLS,
  approvals: [],
}

export const LONG_NAMES: TRecentUpdatePayloadItem = {
  appId: 'app-ingest',
  appName: 'acme-data-platform-ingest-pipeline',
  branchId: 'branch-ingest-long',
  branchName: 'feature/partitioned-backfill-for-late-arriving-events',
  runId: 'run-long',
  workflowId: 'wf-run-long',
  status: 'success',
  createdAt: minsAgo(600),
  ...commit(
    'fedcba98',
    'feat: partition the backfill job by tenant and event day so late-arriving events no longer force a full reprocess of the warehouse tables',
    'Example Developer With A Long Name',
    'acme/data-platform'
  ),
  configGroups: [group('grp-ingest-default', 'default', 0)],
  groupRuns: [
    groupRun('grp-ingest-default', 'default', [
      ['inst-acme', 'success'],
      ['inst-globex', 'success'],
    ]),
  ],
  installsById: INSTALLS,
  approvals: [],
}

export const toStoryItems = (
  updates: TRecentUpdatePayloadItem[]
): TBranchActivityItem[] =>
  updates.flatMap((update) => {
    const item = toRecentUpdateItem({
      ...update,
      orgId: 'org-1',
      nestedInstalls: false,
    })
    return item ? [item] : []
  })

export const ALL_UPDATES = [
  PENDING,
  RUNNING,
  AWAITING_APPROVAL,
  BRANCH_PLAN_APPROVAL,
  MANY_GROUPS,
  FAILED,
  FAILED_BEFORE_ROLLOUT,
  SUCCEEDED,
  CANCELLED,
  MANUAL_NO_COMMIT,
  LABEL_SELECTOR_NO_MATCHES,
  NO_INSTALL_GROUPS,
  LONG_NAMES,
]
