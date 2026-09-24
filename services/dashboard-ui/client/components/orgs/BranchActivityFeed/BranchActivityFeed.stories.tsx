export default {
  title: 'Orgs/BranchActivityFeed',
}

import {
  BranchActivityFeed,
  type TBranchActivityItem,
} from './BranchActivityFeed'

const now = new Date()
const minsAgo = (n: number) =>
  new Date(now.getTime() - n * 60_000).toISOString()

const MIXED_ITEMS: TBranchActivityItem[] = [
  {
    appId: 'app-acme-payments',
    appName: 'acme-payments',
    appHref: '/orgs/org-1/apps/app-acme-payments',
    branchId: 'branch-main',
    branchName: 'main',
    branchHref: '/orgs/org-1/apps/app-acme-payments/branches/branch-main',
    runId: 'run-001',
    runStatus: 'awaiting-approval',
    runCreatedAt: minsAgo(4),
    runHref:
      '/orgs/org-1/apps/app-acme-payments/branches/branch-main/runs/run-001',
    commitMessage: 'feat: add stripe checkout flow to payment widget (#142)',
    commitSha: 'a1b2c3d',
    commitAuthor: 'Margot Ellis',
    commitAvatarUrl: 'https://avatars.githubusercontent.com/u/1',
    commitHref: 'https://github.com/acme/payments/commit/a1b2c3d',
    planGroups: [
      { name: 'canary', installs: 1, hasSelector: false },
      { name: 'enterprise', installs: 2, hasSelector: false },
    ],
    updatedInstalls: [
      {
        id: 'install-1',
        name: 'staging-example',
        group: 'canary',
        href: '/orgs/org-1/installs/install-1',
      },
      {
        id: 'install-2',
        name: 'production-acme',
        group: 'enterprise',
        href: '/orgs/org-1/installs/install-2',
      },
      {
        id: 'install-3',
        name: 'production-globex',
        group: 'enterprise',
        href: '/orgs/org-1/installs/install-3',
      },
    ],
  },
  {
    appId: 'app-acme-gateway',
    appName: 'acme-gateway',
    appHref: '/orgs/org-1/apps/app-acme-gateway',
    branchId: 'branch-feature-rate-limit',
    branchName: 'feature/rate-limit',
    branchHref:
      '/orgs/org-1/apps/app-acme-gateway/branches/branch-feature-rate-limit',
    runId: 'run-002',
    runStatus: 'failed',
    runCreatedAt: minsAgo(18),
    runHref:
      '/orgs/org-1/apps/app-acme-gateway/branches/branch-feature-rate-limit/runs/run-002',
    commitMessage: 'fix: tighten rate-limit headers for downstream services',
    commitSha: 'f4e5d6c',
    commitAuthor: 'Dev Singh',
    commitHref: 'https://github.com/acme/gateway/commit/f4e5d6c',
    planGroups: [{ name: 'customers', installs: 1, hasSelector: true }],
    updatedInstalls: [
      {
        id: 'install-4',
        name: 'development',
        group: 'customers',
        href: '/orgs/org-1/installs/install-4',
      },
    ],
  },
  {
    appId: 'app-acme-portal',
    appName: 'acme-portal',
    appHref: '/orgs/org-1/apps/app-acme-portal',
    branchId: 'branch-main',
    branchName: 'main',
    branchHref: '/orgs/org-1/apps/app-acme-portal/branches/branch-main',
    runId: 'run-003',
    runStatus: 'in-progress',
    runCreatedAt: minsAgo(2),
    runHref:
      '/orgs/org-1/apps/app-acme-portal/branches/branch-main/runs/run-003',
    commitMessage: 'chore: bump node-fetch to 3.3.2',
    commitSha: '7c8b9a0',
    commitAuthor: 'Priya Nair',
    commitAvatarUrl: 'https://avatars.githubusercontent.com/u/2',
    commitHref: 'https://github.com/acme/portal/commit/7c8b9a0',
    planGroups: [
      { name: 'canary', installs: 1, hasSelector: false },
      { name: 'rest', installs: 2, hasSelector: false },
    ],
    updatedInstalls: [
      {
        id: 'install-5',
        name: 'preview-1042',
        group: 'canary',
        href: '/orgs/org-1/installs/install-5',
      },
      {
        id: 'install-2',
        name: 'production-acme',
        group: 'rest',
        href: '/orgs/org-1/installs/install-2',
      },
      {
        id: 'install-3',
        name: 'production-globex',
        group: 'rest',
        href: '/orgs/org-1/installs/install-3',
      },
    ],
  },
  {
    appId: 'app-acme-analytics',
    appName: 'acme-analytics',
    appHref: '/orgs/org-1/apps/app-acme-analytics',
    branchId: 'branch-main',
    branchName: 'main',
    branchHref: '/orgs/org-1/apps/app-acme-analytics/branches/branch-main',
    runId: 'run-004',
    runStatus: 'success',
    runCreatedAt: minsAgo(45),
    runHref:
      '/orgs/org-1/apps/app-acme-analytics/branches/branch-main/runs/run-004',
    commitMessage: 'feat: add export-to-csv button on reports page (#89)',
    commitSha: '2d3e4f5',
    commitAuthor: 'Tomás Reyes',
    commitAvatarUrl: 'https://avatars.githubusercontent.com/u/3',
    commitHref: 'https://github.com/acme/analytics/commit/2d3e4f5',
    planGroups: [{ name: 'default', installs: 1, hasSelector: false }],
    updatedInstalls: [
      {
        id: 'install-4',
        name: 'development',
        group: 'default',
        href: '/orgs/org-1/installs/install-4',
      },
    ],
  },
  {
    appId: 'app-acme-payments',
    appName: 'acme-payments',
    appHref: '/orgs/org-1/apps/app-acme-payments',
    branchId: 'branch-hotfix-tax',
    branchName: 'hotfix/tax-calc',
    branchHref: '/orgs/org-1/apps/app-acme-payments/branches/branch-hotfix-tax',
    runId: 'run-005',
    runStatus: 'success',
    runCreatedAt: minsAgo(120),
    runHref:
      '/orgs/org-1/apps/app-acme-payments/branches/branch-hotfix-tax/runs/run-005',
    commitMessage: 'fix: correct EU VAT calculation for B2B invoices',
    commitSha: '9f0a1b2',
    commitAuthor: 'Margot Ellis',
    commitAvatarUrl: 'https://avatars.githubusercontent.com/u/1',
    commitHref: 'https://github.com/acme/payments/commit/9f0a1b2',
    planGroups: [{ name: 'enterprise', installs: 0, hasSelector: false }],
  },
]

const QUIET_ITEMS: TBranchActivityItem[] = [
  {
    appId: 'app-acme-portal',
    appName: 'acme-portal',
    appHref: '/orgs/org-1/apps/app-acme-portal',
    branchId: 'branch-main',
    branchName: 'main',
    branchHref: '/orgs/org-1/apps/app-acme-portal/branches/branch-main',
    runId: 'run-q1',
    runStatus: 'success',
    runCreatedAt: minsAgo(360),
    runHref:
      '/orgs/org-1/apps/app-acme-portal/branches/branch-main/runs/run-q1',
    commitMessage: 'docs: update README with deployment steps',
    commitSha: 'abc1234',
    commitAuthor: 'Priya Nair',
    commitAvatarUrl: 'https://avatars.githubusercontent.com/u/2',
  },
]

export const MixedStates = () => (
  <div className="max-w-3xl p-6">
    <BranchActivityFeed items={MIXED_ITEMS} />
  </div>
)

export const AllQuiet = () => (
  <div className="max-w-3xl p-6">
    <BranchActivityFeed items={QUIET_ITEMS} />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl p-6">
    <BranchActivityFeed items={[]} />
  </div>
)

export const Loading = () => (
  <div className="max-w-3xl p-6">
    <BranchActivityFeed items={[]} isLoading />
  </div>
)
